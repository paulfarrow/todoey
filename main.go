package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("203"))
	selectedStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")).Background(lipgloss.Color("52"))
	markedStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")).Background(lipgloss.Color("30"))
	markedCursorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")).Background(lipgloss.Color("37"))
	normalStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	statusStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	sidebarStyle      = lipgloss.NewStyle().
				BorderRight(true).
				BorderStyle(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("52")).
				Background(lipgloss.Color("234")).
				Padding(0, 1)
	paneStyle   = lipgloss.NewStyle().Padding(0, 2)
	footerStyle = lipgloss.NewStyle().
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("237")).
			Padding(0, 1)
)

var api TodoistAPI = NewTodoistClient()

var projectColors = []lipgloss.Color{"75", "215", "114", "183", "87", "222", "159", "210"}

func projectTagStyle(idx int) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(projectColors[idx%len(projectColors)])
}

func (m model) projectTag(projectID string) string {
	for i, p := range m.projects {
		if p.ID == projectID {
			s := projectTagStyle(i)
			return s.Render("#") + s.Bold(true).Render(p.Name)
		}
	}
	return ""
}

func main() {
	addTask := flag.String("a", "", "")
	flag.StringVar(addTask, "add-task", "", "Add a task and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: todoist [-a, --add-task \"task text\"]\n")
	}
	flag.Parse()

	if *addTask != "" {
		t, err := api.CreateTask(*addTask)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		line := "Task added: " + t.Content
		if t.Due != nil && t.Due.Date != "" {
			line += " (" + t.Due.Date + ")"
		}
		projects, _ := api.GetProjects()
		for _, p := range projects {
			if p.ID == t.ProjectID {
				line += " #" + p.Name
				break
			}
		}
		fmt.Println(line)
		return
	}

	p := tea.NewProgram(model{cfg: loadConfig(), status: "Loading..."}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
