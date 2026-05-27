package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const baseURL = "https://api.todoist.com/api/v1"

type task struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Due     *struct {
		Date string `json:"date"`
	} `json:"due"`
	Priority  int    `json:"priority"`
	ProjectID string `json:"project_id"`
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
	token := os.Getenv("TODOIST_API_TOKEN")
	return &TodoistClient{token: token, client: &http.Client{}}
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

func (c *TodoistClient) CloseTask(taskID string) error {
	_, err := c.do("POST", "/tasks/"+taskID+"/close", nil)
	return err
}

func (c *TodoistClient) CreateTask(content, projectID string) error {
	body := map[string]string{"content": content}
	if projectID != "" {
		body["project_id"] = projectID
	}
	_, err := c.do("POST", "/tasks", body)
	return err
}
