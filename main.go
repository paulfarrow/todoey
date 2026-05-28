package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("203"))
	selectedStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")).Background(lipgloss.Color("52"))
	markedStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")).Background(lipgloss.Color("30"))
	markedCursorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231")).Background(lipgloss.Color("37"))
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	statusStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("78"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	sidebarStyle  = lipgloss.NewStyle().
			BorderRight(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(lipgloss.Color("52")).
			Background(lipgloss.Color("234")).
			Padding(0, 1)
	paneStyle  = lipgloss.NewStyle().Padding(0, 2)
	footerStyle = lipgloss.NewStyle().
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("237")).
			Padding(0, 1)
)

var api = NewTodoistClient()

// projectColors cycles through distinct foreground colors for project tags.
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

type inputMode int

const (
	modeNormal inputMode = iota
	modeAdd
	modeSearch
	modeGoto
	modeDetail
	modeConfirmDelete
	modeMoveTask
	modeVisual
)

type model struct {
	projects      []project
	tasks         []task
	projectCursor int
	taskCursor    int
	selected      map[int]bool
	visualAnchor  int
	status        string
	input         string
	mode          inputMode
	searchQuery   string
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
		return dataMsg{tasks: tasks}
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
		m.selected = nil
		m.status = fmt.Sprintf("%d tasks", len(msg.tasks))

	case statusMsg:
		m.status = string(msg)
		return m, m.refreshTasks()

	case errMsg:
		m.status = string(msg)

	case tea.KeyMsg:
		if m.mode == modeDetail {
			if msg.String() == "esc" || msg.String() == "ctrl+c" || msg.String() == "q" || msg.String() == "enter" {
				m.mode = modeNormal
			}
			return m, nil
		}
		if m.mode == modeConfirmDelete {
			switch msg.String() {
			case "y", "Y":
				targets := m.selectedTasks()
				m.selected = nil
				m.mode = modeNormal
				var cmds []tea.Cmd
				for _, t := range targets {
					cmds = append(cmds, deleteTask(t.ID, t.Content))
				}
				return m, tea.Batch(cmds...)
			default:
				m.mode = modeNormal
			}
			return m, nil
		}
		if m.mode != modeNormal && m.mode != modeVisual {
			return m.handleInput(msg)
		}
		return m.handleNormal(msg)
	}
	return m, nil
}

func visualRange(anchor, cursor int) map[int]bool {
	out := make(map[int]bool)
	lo, hi := anchor, cursor
	if lo > hi {
		lo, hi = hi, lo
	}
	for i := lo; i <= hi; i++ {
		out[i] = true
	}
	return out
}

func (m model) selectedTasks() []task {
	if len(m.selected) == 0 {
		if m.taskCursor < len(m.tasks) {
			return []task{m.tasks[m.taskCursor]}
		}
		return nil
	}
	var out []task
	for i, t := range m.tasks {
		if m.selected[i] {
			out = append(out, t)
		}
	}
	return out
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

func (m model) gotoCompletion() string {
	lower := strings.ToLower(m.input)
	if lower == "" {
		return ""
	}
	if strings.HasPrefix("today", lower) {
		return "Today"
	}
	for _, p := range m.projects {
		if strings.HasPrefix(strings.ToLower(p.Name), lower) {
			return p.Name
		}
	}
	return ""
}

func (m model) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if m.mode == modeGoto || m.mode == modeMoveTask {
			if c := m.gotoCompletion(); c != "" {
				m.input = c
			}
		}
	case "enter":
		text := strings.TrimSpace(m.input)
		origMode := m.mode
		m.mode = modeNormal
		m.input = ""
		switch origMode {
		case modeSearch:
			if text != "" {
				m.searchQuery = text
				m.status = "Searching..."
				return m, searchTasks(text)
			}
			m.searchQuery = ""
			return m, m.refreshTasks()
		case modeGoto:
			lower := strings.ToLower(text)
			if lower == "today" {
				m.projectCursor = 0
				m.searchQuery = ""
				m.status = "Loading..."
				return m, m.refreshTasks()
			}
			for i, p := range m.projects {
				if strings.ToLower(p.Name) == lower {
					m.projectCursor = i + 1
					m.searchQuery = ""
					m.status = "Loading..."
					return m, m.refreshTasks()
				}
			}
		case modeMoveTask:
			lower := strings.ToLower(text)
			for _, p := range m.projects {
				if strings.ToLower(p.Name) == lower {
					targets := m.selectedTasks()
					m.selected = nil
					var cmds []tea.Cmd
					for _, t := range targets {
						cmds = append(cmds, moveTask(t.ID, t.Content, p.ID))
					}
					return m, tea.Batch(cmds...)
				}
			}
		case modeAdd:
			if text != "" {
				var pid string
				if m.projectCursor > 0 && m.projectCursor-1 < len(m.projects) {
					pid = m.projects[m.projectCursor-1].ID
				}
				return m, createTask(text, pid)
			}
		}
	case "esc", "ctrl+c":
		m.mode = modeNormal
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
	case "q":
		return m, tea.Quit

	case "V":
		if len(m.tasks) > 0 {
			if m.mode == modeVisual {
				m.mode = modeNormal // stop auto-selecting, keep selection
			} else {
				m.mode = modeVisual
				m.visualAnchor = m.taskCursor
				if m.selected == nil {
					m.selected = make(map[int]bool)
				}
				m.selected[m.taskCursor] = true
			}
		}

	case "j", "down":
		if m.taskCursor < len(m.tasks)-1 {
			m.taskCursor++
			if m.mode == modeVisual {
				m.selected = visualRange(m.visualAnchor, m.taskCursor)
			}
		}
	case "k", "up":
		if m.taskCursor > 0 {
			m.taskCursor--
			if m.mode == modeVisual {
				m.selected = visualRange(m.visualAnchor, m.taskCursor)
			}
		}

	case "J":
		if m.projectCursor < len(m.projects) {
			m.projectCursor++
			m.searchQuery = ""
			m.status = "Loading..."
			return m, m.refreshTasks()
		}
	case "K":
		if m.projectCursor > 0 {
			m.projectCursor--
			m.searchQuery = ""
			m.status = "Loading..."
			return m, m.refreshTasks()
		}

	case "g":
		m.taskCursor = 0
	case "G":
		if len(m.tasks) > 0 {
			m.taskCursor = len(m.tasks) - 1
		}

	case "v", " ":
		if len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
			if m.selected == nil {
				m.selected = make(map[int]bool)
			}
			m.selected[m.taskCursor] = !m.selected[m.taskCursor]
			if !m.selected[m.taskCursor] {
				delete(m.selected, m.taskCursor)
			}
			if m.taskCursor < len(m.tasks)-1 {
				m.taskCursor++
			}
		}
	case "esc", "ctrl+c":
		m.mode = modeNormal
		m.selected = nil

	case "x":
		if len(m.tasks) == 0 {
			break
		}
		targets := m.selectedTasks()
		m.selected = nil
		m.mode = modeNormal
		var cmds []tea.Cmd
		for _, t := range targets {
			cmds = append(cmds, closeTask(t.ID, t.Content))
		}
		return m, tea.Batch(cmds...)
	case "d":
		if len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
			m.mode = modeConfirmDelete
		}
	case "enter":
		if len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
			m.mode = modeDetail
		}
	case "a":
		m.mode = modeAdd
		m.input = ""
	case "/":
		m.mode = modeSearch
		m.input = ""
	case "c":
		m.mode = modeGoto
		m.input = ""
	case "alt+m":
		if len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
			m.mode = modeMoveTask
			m.input = ""
		}
	case "r":
		m.searchQuery = ""
		m.status = "Refreshing..."
		return m, m.refreshTasks()
	}
	return m, nil
}

func (m model) viewDetail() string {
	t := m.tasks[m.taskCursor]

	w := m.width
	if w == 0 {
		w = 80
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("52")).
		Padding(1, 2).
		Width(w - 4)

	field := func(label, value string) string {
		if value == "" {
			return ""
		}
		return dimStyle.Render(label+": ") + normalStyle.Render(value) + "\n"
	}

	priorities := map[int]string{4: "p1 (urgent)", 3: "p2 (high)", 2: "p3 (medium)", 1: "p4 (normal)"}

	var b strings.Builder
	b.WriteString(titleStyle.Render("◆ "+t.Content) + "\n\n")

	if t.Description != "" {
		b.WriteString(normalStyle.Render(t.Description) + "\n\n")
	}

	if t.Due != nil {
		s := t.Due.Date
		if t.Due.String != "" {
			s = t.Due.String + " (" + t.Due.Date + ")"
		}
		if t.Due.IsRecurring {
			s += " 🔁"
		}
		b.WriteString(field("Due", s))
	}
	if t.Deadline != nil && t.Deadline.Date != "" {
		b.WriteString(field("Deadline", t.Deadline.Date))
	}
	if t.Duration != nil {
		b.WriteString(field("Duration", fmt.Sprintf("%d %s", t.Duration.Amount, t.Duration.Unit)))
	}
	b.WriteString(field("Priority", priorities[t.Priority]))
	b.WriteString(field("Project", m.projectTag(t.ProjectID)))
	if len(t.Labels) > 0 {
		b.WriteString(field("Labels", strings.Join(t.Labels, ", ")))
	}
	if t.SectionID != "" {
		b.WriteString(field("Section ID", t.SectionID))
	}
	if t.ParentID != "" {
		b.WriteString(field("Parent ID", t.ParentID))
	}
	b.WriteString(field("ID", t.ID))
	b.WriteString(field("Added", t.AddedAt))
	b.WriteString(field("Updated", t.UpdatedAt))

	b.WriteString("\n" + helpStyle.Render("esc / enter: back"))

	return style.Render(b.String())
}

func (m model) View() string {
	if m.mode == modeDetail && len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
		return m.viewDetail()
	}
	sidebarWidth := 24
	var sidebar strings.Builder
	sidebar.WriteString(titleStyle.Render("◆ Projects") + "\n\n")

	itemWidth := sidebarWidth - 2
	renderItem := func(label string, selected bool) string {
		if selected {
			return selectedStyle.Width(itemWidth).Render(" " + label)
		}
		return normalStyle.Render("  " + label)
	}

	sidebar.WriteString(renderItem("Today", m.projectCursor == 0) + "\n")
	for i, p := range m.projects {
		name := p.Name
		if len(name) > sidebarWidth-4 {
			name = name[:sidebarWidth-4]
		}
		sidebar.WriteString(renderItem(name, i+1 == m.projectCursor) + "\n")
	}

	var main strings.Builder
	if m.searchQuery != "" {
		main.WriteString(titleStyle.Render("◆ Search: "+m.searchQuery) + "  " + dimStyle.Render("(r to clear, J/K to browse projects)") + "\n\n")
	} else if m.projectCursor == 0 {
		main.WriteString(titleStyle.Render("◆ Today") + "\n\n")
	} else if m.projectCursor-1 < len(m.projects) {
		main.WriteString(titleStyle.Render("◆ "+m.projects[m.projectCursor-1].Name) + "\n\n")
	}
	if len(m.tasks) == 0 {
		main.WriteString(dimStyle.Render("  No tasks") + "\n")
	}
	today := time.Now().Format("2006-01-02")
	for i, t := range m.tasks {
		due := dueStr(t)
		overdue := t.Due != nil && t.Due.Date != "" && t.Due.Date < today
		dueRendered := dimStyle.Render(due)
		if overdue {
			dueRendered = errorStyle.Render(due)
		}
		tag := m.projectTag(t.ProjectID)
		if tag != "" {
			tag = " " + tag
		}
		isCursor := i == m.taskCursor
		isMarked := m.selected[i]
		var line string
		switch {
		case isCursor && isMarked:
			line = markedCursorStyle.Render(" ▶ "+t.Content) + dueRendered + tag
		case isCursor:
			line = selectedStyle.Render(" ○ "+t.Content) + dueRendered + tag
		case isMarked:
			line = markedStyle.Render(" ● "+t.Content) + dueRendered + tag
		default:
			line = normalStyle.Render("  ○ "+t.Content) + dueRendered + tag
		}
		main.WriteString(line + "\n")
	}

	if m.mode == modeAdd {
		main.WriteString("\n" + titleStyle.Render("  New task: ") + m.input + "█\n")
	} else if m.mode == modeSearch {
		main.WriteString("\n" + titleStyle.Render("  Search: ") + m.input + "█\n")
	} else if m.mode == modeGoto {
		ghost := ""
		if c := m.gotoCompletion(); c != "" && c != m.input {
			ghost = dimStyle.Render(c[len(m.input):])
		}
		main.WriteString("\n" + titleStyle.Render("  Go to project: ") + m.input + ghost + "█\n")
	} else if m.mode == modeConfirmDelete && len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
		var deletePrompt string
		if len(m.selected) > 1 {
			deletePrompt = fmt.Sprintf("  Delete %d tasks? ", len(m.selected))
		} else {
			deletePrompt = "  Delete \"" + m.tasks[m.taskCursor].Content + "\"? "
		}
		main.WriteString("\n" + errorStyle.Render(deletePrompt) + normalStyle.Render("[y] yes  [any] cancel") + "\n")
	} else if m.mode == modeMoveTask && len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
		ghost := ""
		if c := m.gotoCompletion(); c != "" && c != m.input {
			ghost = dimStyle.Render(c[len(m.input):])
		}
		main.WriteString("\n" + titleStyle.Render("  Move to project: ") + m.input + ghost + "█\n")
	}

	sideRendered := sidebarStyle.Width(sidebarWidth).Render(sidebar.String())
	mainRendered := paneStyle.Render(main.String())
	content := lipgloss.JoinHorizontal(lipgloss.Top, sideRendered, mainRendered)

	statusText := m.status
	statusRendered := statusStyle.Render(statusText)
	if strings.HasPrefix(m.status, "Error") {
		statusRendered = errorStyle.Render(statusText)
	}
	footerWidth := m.width
	if footerWidth == 0 {
		footerWidth = 80
	}
	footer := footerStyle.Width(footerWidth - 2).Render(
		statusRendered + "\n" +
			helpStyle.Render("j/k:tasks  J/K:projects  x:complete  d:delete  a:add  /:search  c:goto  alt+m:move  r:refresh  g/G:top/bottom  V:visual  q:quit"),
	)
	if m.mode == modeVisual {
		footer = footerStyle.Width(footerWidth - 2).Render(
			markedCursorStyle.Render(" VISUAL ") + "  " + statusRendered + "\n" +
				helpStyle.Render("j/k:extend selection  V/esc:exit  x:complete  d:delete  alt+m:move"),
		)
	}

	return content + "\n" + footer
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
