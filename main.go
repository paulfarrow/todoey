package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170"))
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	statusStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	sidebarStyle  = lipgloss.NewStyle().BorderRight(true).BorderStyle(lipgloss.NormalBorder()).Padding(0, 1)
	paneStyle     = lipgloss.NewStyle().Padding(0, 1)
)

var api = NewTodoistClient()

type model struct {
	projects      []project
	tasks         []task
	projectCursor int
	taskCursor    int
	status        string
	input         string
	inputOn       bool
	width         int
	height        int
}

type dataMsg struct {
	projects []project
	tasks    []task
}
type statusMsg string
type errMsg string

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
		return dataMsg{projects: projects, tasks: tasks}
	}
}

func fetchTodayTasks() tea.Cmd {
	return func() tea.Msg {
		tasks, err := api.GetTodayTasks()
		if err != nil {
			return errMsg(fmt.Sprintf("Error: %v", err))
		}
		return dataMsg{tasks: tasks}
	}
}

func fetchProjectTasks(projectID string) tea.Cmd {
	return func() tea.Msg {
		tasks, err := api.GetTasksByProject(projectID)
		if err != nil {
			return errMsg(fmt.Sprintf("Error: %v", err))
		}
		return dataMsg{tasks: tasks}
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

func createTask(content, projectID string) tea.Cmd {
	return func() tea.Msg {
		if err := api.CreateTask(content, projectID); err != nil {
			return errMsg(fmt.Sprintf("Error adding: %v", err))
		}
		return statusMsg(fmt.Sprintf("Added: %s", content))
	}
}

func (m model) Init() tea.Cmd {
	return fetchInitialData()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case dataMsg:
		if msg.projects != nil {
			m.projects = msg.projects
		}
		m.tasks = msg.tasks
		m.taskCursor = 0
		m.status = fmt.Sprintf("%d tasks", len(msg.tasks))

	case statusMsg:
		m.status = string(msg)
		return m, m.refreshTasks()

	case errMsg:
		m.status = string(msg)

	case tea.KeyMsg:
		if m.inputOn {
			return m.handleInput(msg)
		}
		return m.handleNormal(msg)
	}
	return m, nil
}

func (m model) refreshTasks() tea.Cmd {
	if m.projectCursor == 0 {
		return fetchTodayTasks()
	}
	if m.projectCursor-1 < len(m.projects) {
		return fetchProjectTasks(m.projects[m.projectCursor-1].ID)
	}
	return fetchTodayTasks()
}

func (m model) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		text := strings.TrimSpace(m.input)
		m.inputOn = false
		m.input = ""
		if text != "" {
			var pid string
			if m.projectCursor > 0 && m.projectCursor-1 < len(m.projects) {
				pid = m.projects[m.projectCursor-1].ID
			}
			return m, createTask(text, pid)
		}
	case "esc":
		m.inputOn = false
		m.input = ""
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	default:
		if len(msg.String()) == 1 || msg.String() == " " {
			m.input += msg.String()
		}
	}
	return m, nil
}

func (m model) handleNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "j", "down":
		if m.taskCursor < len(m.tasks)-1 {
			m.taskCursor++
		}
	case "k", "up":
		if m.taskCursor > 0 {
			m.taskCursor--
		}

	case "J":
		if m.projectCursor < len(m.projects) {
			m.projectCursor++
			m.status = "Loading..."
			return m, m.refreshTasks()
		}
	case "K":
		if m.projectCursor > 0 {
			m.projectCursor--
			m.status = "Loading..."
			return m, m.refreshTasks()
		}

	case "g":
		m.taskCursor = 0
	case "G":
		if len(m.tasks) > 0 {
			m.taskCursor = len(m.tasks) - 1
		}

	case "x":
		if len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
			t := m.tasks[m.taskCursor]
			return m, closeTask(t.ID, t.Content)
		}
	case "a":
		m.inputOn = true
		m.input = ""
	case "r":
		m.status = "Refreshing..."
		return m, m.refreshTasks()
	}
	return m, nil
}

func (m model) View() string {
	sidebarWidth := 24
	var sidebar strings.Builder
	sidebar.WriteString(titleStyle.Render("Projects") + "\n\n")

	if m.projectCursor == 0 {
		sidebar.WriteString(selectedStyle.Render("> Today") + "\n")
	} else {
		sidebar.WriteString(normalStyle.Render("  Today") + "\n")
	}

	for i, p := range m.projects {
		name := p.Name
		if len(name) > sidebarWidth-4 {
			name = name[:sidebarWidth-4]
		}
		if i+1 == m.projectCursor {
			sidebar.WriteString(selectedStyle.Render("> "+name) + "\n")
		} else {
			sidebar.WriteString(normalStyle.Render("  "+name) + "\n")
		}
	}

	var main strings.Builder
	if m.projectCursor == 0 {
		main.WriteString(titleStyle.Render("Today") + "\n\n")
	} else if m.projectCursor-1 < len(m.projects) {
		main.WriteString(titleStyle.Render(m.projects[m.projectCursor-1].Name) + "\n\n")
	}
	if len(m.tasks) == 0 {
		main.WriteString(dimStyle.Render("  No tasks") + "\n")
	}
	for i, t := range m.tasks {
		if i == m.taskCursor {
			main.WriteString(selectedStyle.Render("> "+t.Content) + dimStyle.Render(dueStr(t)) + "\n")
		} else {
			main.WriteString(normalStyle.Render("  "+t.Content) + dimStyle.Render(dueStr(t)) + "\n")
		}
	}

	if m.inputOn {
		main.WriteString("\n" + titleStyle.Render("Add task: ") + m.input + "█\n")
	}

	sideRendered := sidebarStyle.Width(sidebarWidth).Render(sidebar.String())
	mainRendered := paneStyle.Render(main.String())
	content := lipgloss.JoinHorizontal(lipgloss.Top, sideRendered, mainRendered)

	var footer strings.Builder
	if strings.HasPrefix(m.status, "Error") {
		footer.WriteString(errorStyle.Render(m.status))
	} else {
		footer.WriteString(statusStyle.Render(m.status))
	}
	footer.WriteString("\n")
	footer.WriteString(helpStyle.Render("j/k:tasks  J/K:projects  x:complete  a:add  r:refresh  g/G:top/bottom  q:quit"))

	return content + "\n" + footer.String()
}

func dueStr(t task) string {
	if t.Due != nil && t.Due.Date != "" {
		return " (" + t.Due.Date + ")"
	}
	return ""
}

func main() {
	p := tea.NewProgram(model{status: "Loading..."}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
