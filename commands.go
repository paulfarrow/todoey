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
