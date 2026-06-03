package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const baseURL = "https://api.todoist.com/api/v1"

type task struct {
	ID          string `json:"id"`
	Content     string `json:"content"`
	Description string `json:"description"`
	Due         *struct {
		Date        string `json:"date"`
		String      string `json:"string"`
		IsRecurring bool   `json:"is_recurring"`
	} `json:"due"`
	Deadline *struct {
		Date string `json:"date"`
	} `json:"deadline"`
	Duration *struct {
		Amount int    `json:"amount"`
		Unit   string `json:"unit"`
	} `json:"duration"`
	Priority  int      `json:"priority"`
	ProjectID string   `json:"project_id"`
	SectionID string   `json:"section_id"`
	ParentID  string   `json:"parent_id"`
	Labels    []string `json:"labels"`
	AddedAt   string   `json:"added_at"`
	UpdatedAt string   `json:"updated_at"`
}

type project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TodoistClient struct {
	token  string
	client *http.Client
}

func NewTodoistClient() *TodoistClient {
	cfg := loadConfig()
	return &TodoistClient{token: cfg.APIToken, client: &http.Client{}}
}

func (c *TodoistClient) do(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}
	return data, nil
}

type paginatedTasks struct {
	Results    []task  `json:"results"`
	NextCursor *string `json:"next_cursor"`
}

type paginatedProjects struct {
	Results    []project `json:"results"`
	NextCursor *string   `json:"next_cursor"`
}

func (c *TodoistClient) GetProjects() ([]project, error) {
	data, err := c.do("GET", "/projects?limit=200", nil)
	if err != nil {
		return nil, err
	}
	var resp paginatedProjects
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}

func (c *TodoistClient) GetTodayTasks() ([]task, error) {
	data, err := c.do("GET", "/tasks/filter?query=today%20%7C%20overdue&limit=200", nil)
	if err != nil {
		return nil, err
	}
	var resp paginatedTasks
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}

func (c *TodoistClient) GetTasksByProject(projectID string) ([]task, error) {
	data, err := c.do("GET", "/tasks?project_id="+projectID+"&limit=200", nil)
	if err != nil {
		return nil, err
	}
	var resp paginatedTasks
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}

func (c *TodoistClient) DeleteTask(taskID string) error {
	_, err := c.do("DELETE", "/tasks/"+taskID, nil)
	return err
}

func (c *TodoistClient) CloseTask(taskID string) error {
	_, err := c.do("POST", "/tasks/"+taskID+"/close", nil)
	return err
}

func (c *TodoistClient) MoveTask(taskID, projectID string) error {
	_, err := c.do("POST", "/tasks/"+taskID+"/move", map[string]string{"project_id": projectID})
	return err
}

func (c *TodoistClient) SearchTasks(query string) ([]task, error) {
	data, err := c.do("GET", "/tasks/filter?query="+url.QueryEscape("search: "+query)+"&limit=200", nil)
	if err != nil {
		return nil, err
	}
	var resp paginatedTasks
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return resp.Results, nil
}

func (c *TodoistClient) RescheduleTask(taskID, dueString string) error {
	_, err := c.do("POST", "/tasks/"+taskID, map[string]string{"due_string": dueString})
	return err
}

func (c *TodoistClient) CreateTask(text string) (*task, error) {
	body := map[string]string{"text": text}
	data, err := c.do("POST", "/tasks/quick", body)
	if err != nil {
		return nil, err
	}
	var t task
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (c *TodoistClient) UpdateTask(taskID string, fields map[string]string) error {
	_, err := c.do("POST", "/tasks/"+taskID, fields)
	return err
}
