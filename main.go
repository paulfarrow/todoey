package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170"))
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	statusStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type view int

const (
	viewToday view = iota
	viewInbox
	viewProjects
	viewAdd
)

type task struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Due     *struct {
		Date string `json:"date"`
	} `json:"due"`
	Priority  int    `json:"priority"`
	ProjectID string `json:"projectId"`
}

type taskResponse struct {
	Results []task `json:"results"`
}

type project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type projectResponse struct {
	Results []project `json:"results"`
}

type model struct {
	tasks    []task
	projects []project
	cursor   int
	view     view
	status   string
	input    string
	inputOn  bool
	width    int
	height   int
}

type tasksMsg []task
type projectsMsg []project
type statusMsg string
type errMsg string

func fetchTasks(args ...string) tea.Cmd {
	return func() tea.Msg {
		cmdArgs := append([]string{"task", "list", "--json"}, args...)
		out, err := exec.Command("td", cmdArgs...).Output()
		if err != nil {
			return errMsg(fmt.Sprintf("Error: %v", err))
		}
		var resp taskResponse
		if err := json.Unmarshal(out, &resp); err != nil {
			return errMsg(fmt.Sprintf("Parse error: %v", err))
		}
		return tasksMsg(resp.Results)
	}
}

func fetchToday() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("td", "today", "--json").Output()
		if err != nil {
			return errMsg(fmt.Sprintf("Error: %v", err))
		}
		var resp taskResponse
		if err := json.Unmarshal(out, &resp); err != nil {
			return errMsg(fmt.Sprintf("Parse error: %v", err))
		}
		return tasksMsg(resp.Results)
	}
}

func fetchInbox() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("td", "inbox", "--json").Output()
		if err != nil {
			return errMsg(fmt.Sprintf("Error: %v", err))
		}
		var resp taskResponse
		if err := json.Unmarshal(out, &resp); err != nil {
			return errMsg(fmt.Sprintf("Parse error: %v", err))
		}
		return tasksMsg(resp.Results)
	}
}

func fetchProjects() tea.Cmd {
	return func() tea.Msg {
		out, err := exec.Command("td", "project", "list", "--json").Output()
		if err != nil {
			return errMsg(fmt.Sprintf("Error: %v", err))
		}
		var resp projectResponse
		if err := json.Unmarshal(out, &resp); err != nil {
			return errMsg(fmt.Sprintf("Parse error: %v", err))
		}
		return projectsMsg(resp.Results)
	}
}

func completeTask(name string) tea.Cmd {
	return func() tea.Msg {
		_, err := exec.Command("td", "task", "complete", name).Output()
		if err != nil {
			return errMsg(fmt.Sprintf("Error completing: %v", err))
		}
		return statusMsg(fmt.Sprintf("Completed: %s", name))
	}
}

func addTask(text string) tea.Cmd {
	return func() tea.Msg {
		_, err := exec.Command("td", "add", text).Output()
		if err != nil {
			return errMsg(fmt.Sprintf("Error adding: %v", err))
		}
		return statusMsg(fmt.Sprintf("Added: %s", text))
	}
}

func (m model) Init() tea.Cmd {
	return fetchToday()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tasksMsg:
		m.tasks = msg
		m.projects = nil
		m.cursor = 0
		m.status = fmt.Sprintf("%d tasks", len(msg))

	case projectsMsg:
		m.projects = msg
		m.tasks = nil
		m.cursor = 0
		m.status = fmt.Sprintf("%d projects", len(msg))

	case statusMsg:
		m.status = string(msg)
		// Refresh current view
		return m, m.refreshCmd()

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

func (m model) refreshCmd() tea.Cmd {
	switch m.view {
	case viewToday:
		return fetchToday()
	case viewInbox:
		return fetchInbox()
	case viewProjects:
		return fetchProjects()
	default:
		return fetchToday()
	}
}

func (m model) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		text := strings.TrimSpace(m.input)
		m.inputOn = false
		m.input = ""
		if text != "" {
			return m, addTask(text)
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

	// Vim navigation
	case "j", "down":
		m.cursor++
		if m.cursor >= m.listLen() {
			m.cursor = m.listLen() - 1
		}
	case "k", "up":
		m.cursor--
		if m.cursor < 0 {
			m.cursor = 0
		}
	case "g":
		m.cursor = 0
	case "G":
		m.cursor = m.listLen() - 1

	// Views
	case "alt+1":
		m.view = viewToday
		m.status = "Loading..."
		return m, fetchToday()
	case "alt+2":
		m.view = viewInbox
		m.status = "Loading..."
		return m, fetchInbox()
	case "alt+3":
		m.view = viewProjects
		m.status = "Loading..."
		return m, fetchProjects()

	// Actions
	case "x":
		if len(m.tasks) > 0 && m.cursor < len(m.tasks) {
			return m, completeTask(m.tasks[m.cursor].Content)
		}
	case "a":
		m.inputOn = true
		m.input = ""
	case "r":
		m.status = "Refreshing..."
		return m, m.refreshCmd()
	}
	return m, nil
}

func (m model) listLen() int {
	if m.view == viewProjects {
		return max(len(m.projects), 1)
	}
	return max(len(m.tasks), 1)
}

func (m model) View() string {
	var b strings.Builder

	// Header
	tabs := []string{"[1]Today", "[2]Inbox", "[3]Projects"}
	for i, t := range tabs {
		if view(i) == m.view {
			b.WriteString(titleStyle.Render(t))
		} else {
			b.WriteString(dimStyle.Render(t))
		}
		b.WriteString("  ")
	}
	b.WriteString("\n\n")

	// Content
	if m.view == viewProjects {
		for i, p := range m.projects {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
				b.WriteString(selectedStyle.Render(cursor + p.Name))
			} else {
				b.WriteString(normalStyle.Render(cursor + p.Name))
			}
			b.WriteString("\n")
		}
	} else {
		if len(m.tasks) == 0 {
			b.WriteString(dimStyle.Render("  No tasks"))
			b.WriteString("\n")
		}
		for i, t := range m.tasks {
			cursor := "  "
			line := t.Content
			if t.Due != nil && t.Due.Date != "" {
				line += dimStyle.Render(" (" + t.Due.Date + ")")
			}
			if i == m.cursor {
				cursor = "> "
				b.WriteString(selectedStyle.Render(cursor+t.Content) + dimStyle.Render(dueStr(t)))
			} else {
				b.WriteString(normalStyle.Render(cursor+t.Content) + dimStyle.Render(dueStr(t)))
			}
			b.WriteString("\n")
		}
	}

	// Input
	if m.inputOn {
		b.WriteString("\n")
		b.WriteString(titleStyle.Render("Add task: ") + m.input + "█")
		b.WriteString("\n")
	}

	// Status bar
	b.WriteString("\n")
	if strings.HasPrefix(m.status, "Error") {
		b.WriteString(errorStyle.Render(m.status))
	} else {
		b.WriteString(statusStyle.Render(m.status))
	}
	b.WriteString("\n")

	// Help
	b.WriteString(helpStyle.Render("j/k:navigate  x:complete  a:add  r:refresh  g/G:top/bottom  1-3:views  q:quit"))

	return b.String()
}

func dueStr(t task) string {
	if t.Due != nil && t.Due.Date != "" {
		return " (" + t.Due.Date + ")"
	}
	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	p := tea.NewProgram(model{status: "Loading..."}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
