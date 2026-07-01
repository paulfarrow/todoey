package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func fetchInitialData() tea.Cmd {
	return func() tea.Msg {
		projects, err := api.GetProjects()
		if err != nil {
			return errMsg(fmt.Sprintf("Error: %v", err))
		}
		tasks, err := api.GetTodayTasks()
		if err != nil {
			return errMsg(fmt.Sprintf("Error: %v", err))
		}
		return dataMsg{projects: projects, tasks: tasks, projectCursor: 0}
	}
}

func fetchTodayTasks(cursor int) tea.Cmd {
	return func() tea.Msg {
		tasks, err := api.GetTodayTasks()
		if err != nil {
			return errMsg(fmt.Sprintf("Error: %v", err))
		}
		return dataMsg{tasks: tasks, projectCursor: cursor}
	}
}

func fetchProjectTasks(projectID string, cursor int) tea.Cmd {
	return func() tea.Msg {
		tasks, err := api.GetTasksByProject(projectID)
		if err != nil {
			return errMsg(fmt.Sprintf("Error: %v", err))
		}
		return dataMsg{tasks: tasks, projectCursor: cursor}
	}
}

func moveTask(taskID, taskContent, projectID string) tea.Cmd {
	return func() tea.Msg {
		if err := api.MoveTask(taskID, projectID); err != nil {
			return errMsg(fmt.Sprintf("Error moving: %v", err))
		}
		return statusMsg(fmt.Sprintf("Moved: %s", taskContent))
	}
}

func searchTasks(query string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := api.SearchTasks(query)
		if err != nil {
			return errMsg(fmt.Sprintf("Error: %v", err))
		}
		return dataMsg{tasks: tasks, projectCursor: -1}
	}
}

func deleteTask(id, content string) tea.Cmd {
	return func() tea.Msg {
		if err := api.DeleteTask(id); err != nil {
			return errMsg(fmt.Sprintf("Error deleting: %v", err))
		}
		return statusMsg(fmt.Sprintf("Deleted: %s", content))
	}
}

func closeTask(id, content string) tea.Cmd {
	return func() tea.Msg {
		if err := api.CloseTask(id); err != nil {
			return errMsg(fmt.Sprintf("Error completing: %v", err))
		}
		return statusMsg(fmt.Sprintf("Completed: %s", content))
	}
}

func rescheduleTask(id, content, dueString string) tea.Cmd {
	return func() tea.Msg {
		if err := api.RescheduleTask(id, dueString); err != nil {
			return errMsg(fmt.Sprintf("Error rescheduling: %v", err))
		}
		return statusMsg(fmt.Sprintf("Rescheduled: %s", content))
	}
}

func updateTask(id, content string, fields map[string]string) tea.Cmd {
	return func() tea.Msg {
		if err := api.UpdateTask(id, fields); err != nil {
			return errMsg(fmt.Sprintf("Error updating: %v", err))
		}
		return statusMsg(fmt.Sprintf("Updated: %s", content))
	}
}

func createTask(content string) tea.Cmd {
	return func() tea.Msg {
		if _, err := api.CreateTask(content); err != nil {
			return errMsg(fmt.Sprintf("Error adding: %v", err))
		}
		return statusMsg(fmt.Sprintf("Added: %s", content))
	}
}

type sectionsMsg struct {
	sections []section
}

type commentsMsg struct {
	comments []comment
}

type subTasksMsg struct {
	subTasks []task
}

func fetchSections(projectID string) tea.Cmd {
	return func() tea.Msg {
		sections, err := api.GetSections(projectID)
		if err != nil {
			return errMsg(fmt.Sprintf("Error fetching sections: %v", err))
		}
		return sectionsMsg{sections: sections}
	}
}

func fetchComments(taskID string) tea.Cmd {
	return func() tea.Msg {
		comments, err := api.GetComments(taskID)
		if err != nil {
			return errMsg(fmt.Sprintf("Error fetching comments: %v", err))
		}
		return commentsMsg{comments: comments}
	}
}

func addComment(taskID, content string) tea.Cmd {
	return func() tea.Msg {
		if _, err := api.AddComment(taskID, content); err != nil {
			return errMsg(fmt.Sprintf("Error adding comment: %v", err))
		}
		return statusMsg(fmt.Sprintf("Comment added"))
	}
}

func createSubTask(content, parentID string) tea.Cmd {
	return func() tea.Msg {
		if _, err := api.CreateSubTask(content, parentID); err != nil {
			return errMsg(fmt.Sprintf("Error adding sub-task: %v", err))
		}
		return statusMsg(fmt.Sprintf("Sub-task added: %s", content))
	}
}

func fetchSubTasks(parentID string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := api.GetSubTasks(parentID)
		if err != nil {
			return errMsg(fmt.Sprintf("Error fetching sub-tasks: %v", err))
		}
		if tasks == nil {
			tasks = []task{}
		}
		return subTasksMsg{subTasks: tasks}
	}
}
