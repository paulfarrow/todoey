package main

import (
	"flag"
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
	modeReschedule
	modeDetailEditContent
	modeDetailEditDesc
	modeDetailReschedule
	modeDetailMove
	modeDetailConfirmDelete
	modeAddDesc
)

type model struct {
	cfg           config
	projects      []project
	tasks         []task
	projectCursor int
	taskCursor    int
	selected      map[int]bool
	visualAnchor  int
	refreshGen    int
	status        string
	input         string
	detailField   string
	mode          inputMode
	searchQuery   string
	width         int
	height        int
}

type tickMsg struct{ gen int }

type dataMsg struct {
	projects      []project
	tasks         []task
	projectCursor int // cursor value at time of fetch; -1 means ignore check
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

func (m model) Init() tea.Cmd {
	if m.cfg.AutoRefresh && m.cfg.RefreshInterval > 0 {
		gen := m.refreshGen
		tick := tea.Tick(time.Duration(m.cfg.RefreshInterval)*time.Second, func(time.Time) tea.Msg { return tickMsg{gen} })
		return tea.Batch(fetchInitialData(), tick)
	}
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
		if msg.projectCursor != -1 && msg.projectCursor != m.projectCursor {
			return m, nil // stale result, discard
		}
		m.tasks = msg.tasks
		if n := len(msg.tasks); m.taskCursor >= n {
			m.taskCursor = max(0, n-1)
		}
		m.selected = nil
		m.status = fmt.Sprintf("%d tasks", len(msg.tasks))

	case statusMsg:
		m.status = string(msg)
		m.refreshGen++; return m, m.refreshTasks()

	case tickMsg:
		if msg.gen == m.refreshGen {
			m.refreshGen++; return m, m.refreshTasks()
		}

	case errMsg:
		m.status = string(msg)

	case tea.KeyMsg:
		if m.mode == modeDetail ||
			m.mode == modeDetailEditContent ||
			m.mode == modeDetailEditDesc ||
			m.mode == modeDetailReschedule ||
			m.mode == modeDetailMove ||
			m.mode == modeDetailConfirmDelete {
			return m.handleDetail(msg)
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
	var fetch tea.Cmd
	if m.projectCursor == 0 {
		fetch = fetchTodayTasks(0)
	} else if m.projectCursor-1 < len(m.projects) {
		fetch = fetchProjectTasks(m.projects[m.projectCursor-1].ID, m.projectCursor)
	} else {
		fetch = fetchTodayTasks(0)
	}
	if m.cfg.AutoRefresh && m.cfg.RefreshInterval > 0 {
		gen := m.refreshGen
		tick := tea.Tick(time.Duration(m.cfg.RefreshInterval)*time.Second, func(time.Time) tea.Msg { return tickMsg{gen} })
		return tea.Batch(fetch, tick)
	}
	return fetch
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

func (m model) addTaskCompletion() string {
	idx := strings.LastIndex(m.input, "#")
	if idx < 0 {
		return ""
	}
	fragment := strings.ToLower(m.input[idx+1:])
	if fragment == "" {
		return ""
	}
	for _, p := range m.projects {
		if strings.HasPrefix(strings.ToLower(p.Name), fragment) {
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
		if m.mode == modeAdd {
			if c := m.addTaskCompletion(); c != "" {
				idx := strings.LastIndex(m.input, "#")
				m.input = m.input[:idx+1] + c
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
			m.refreshGen++; return m, m.refreshTasks()
		case modeGoto:
			lower := strings.ToLower(text)
			if lower == "today" {
				m.projectCursor = 0
				m.searchQuery = ""
				m.status = "Loading..."
				m.refreshGen++; return m, m.refreshTasks()
			}
			for i, p := range m.projects {
				if strings.ToLower(p.Name) == lower {
					m.projectCursor = i + 1
					m.searchQuery = ""
					m.status = "Loading..."
					m.refreshGen++; return m, m.refreshTasks()
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
		case modeReschedule:
			if text != "" {
				targets := m.selectedTasks()
				m.selected = nil
				var cmds []tea.Cmd
				for _, t := range targets {
					cmds = append(cmds, rescheduleTask(t.ID, t.Content, text))
				}
				return m, tea.Batch(cmds...)
			}
		case modeAdd:
			if text != "" {
				m.mode = modeAddDesc
				m.detailField = text
				m.input = ""
				return m, nil
			}
		case modeAddDesc:
			content := m.detailField
			m.detailField = ""
			if text != "" {
				content += " // " + text
			}
			return m, createTask(content)
		}
	case "esc", "ctrl+c":
		m.mode = modeNormal
		m.input = ""
		m.detailField = ""
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
			m.refreshGen++; return m, m.refreshTasks()
		}
	case "K":
		if m.projectCursor > 0 {
			m.projectCursor--
			m.searchQuery = ""
			m.status = "Loading..."
			m.refreshGen++; return m, m.refreshTasks()
		}

	case "T":
		m.projectCursor = 0
		m.searchQuery = ""
		m.status = "Loading..."
		m.refreshGen++; return m, m.refreshTasks()

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
		if m.searchQuery != "" {
			m.searchQuery = ""
			m.status = "Loading..."
			m.refreshGen++; return m, m.refreshTasks()
		}
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
		if len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
			m.mode = modeReschedule
			m.input = ""
		}
	case "alt+r":
		m.searchQuery = ""
		m.status = "Refreshing..."
		m.refreshGen++; return m, m.refreshTasks()
	}
	return m, nil
}

func (m model) handleDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	t := m.tasks[m.taskCursor]

	// Text-input sub-modes
	if m.mode == modeDetailEditContent || m.mode == modeDetailEditDesc ||
		m.mode == modeDetailReschedule || m.mode == modeDetailMove {
		switch msg.String() {
		case "esc":
			m.mode = modeDetail
			m.detailField = ""
		case "enter":
			text := strings.TrimSpace(m.detailField)
			origMode := m.mode
			m.mode = modeDetail
			m.detailField = ""
			switch origMode {
			case modeDetailEditContent:
				if text != "" {
					return m, updateTask(t.ID, t.Content, map[string]string{"content": text})
				}
			case modeDetailEditDesc:
				return m, updateTask(t.ID, t.Content, map[string]string{"description": text})
			case modeDetailReschedule:
				if text != "" {
					return m, rescheduleTask(t.ID, t.Content, text)
				}
			case modeDetailMove:
				lower := strings.ToLower(text)
				for _, p := range m.projects {
					if strings.ToLower(p.Name) == lower {
						return m, moveTask(t.ID, t.Content, p.ID)
					}
				}
			}
		case "tab":
			if m.mode == modeDetailMove {
				lower := strings.ToLower(m.detailField)
				for _, p := range m.projects {
					if strings.HasPrefix(strings.ToLower(p.Name), lower) {
						m.detailField = p.Name
						break
					}
				}
			}
		case "backspace":
			if len(m.detailField) > 0 {
				m.detailField = m.detailField[:len(m.detailField)-1]
			}
		default:
			if len(msg.String()) == 1 || msg.String() == " " {
				m.detailField += msg.String()
			}
		}
		return m, nil
	}

	if m.mode == modeDetailConfirmDelete {
		switch msg.String() {
		case "y", "Y":
			m.mode = modeNormal
			return m, deleteTask(t.ID, t.Content)
		default:
			m.mode = modeDetail
		}
		return m, nil
	}

	// modeDetail — action key dispatch
	switch msg.String() {
	case "esc", "q":
		m.mode = modeNormal
	case "e":
		m.mode = modeDetailEditContent
		m.detailField = t.Content
	case "E":
		m.mode = modeDetailEditDesc
		m.detailField = t.Description
	case "r":
		m.mode = modeDetailReschedule
		m.detailField = ""
	case "x":
		m.mode = modeNormal
		return m, closeTask(t.ID, t.Content)
	case "d":
		m.mode = modeDetailConfirmDelete
	case "alt+m":
		m.mode = modeDetailMove
		m.detailField = ""
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

	b.WriteString("\n")
	switch m.mode {
	case modeDetailEditContent:
		b.WriteString(titleStyle.Render("  Edit content: ") + m.detailField + "█\n")
	case modeDetailEditDesc:
		b.WriteString(titleStyle.Render("  Edit description: ") + m.detailField + "█\n")
	case modeDetailReschedule:
		b.WriteString(titleStyle.Render("  Reschedule to: ") + m.detailField + "█\n")
	case modeDetailMove:
		ghost := ""
		lower := strings.ToLower(m.detailField)
		if lower != "" {
			for _, p := range m.projects {
				if strings.HasPrefix(strings.ToLower(p.Name), lower) {
					ghost = dimStyle.Render(p.Name[len(m.detailField):])
					break
				}
			}
		}
		b.WriteString(titleStyle.Render("  Move to project: ") + m.detailField + ghost + "█\n")
	case modeDetailConfirmDelete:
		b.WriteString(errorStyle.Render("  Delete this task? ") + normalStyle.Render("[y] yes  [any] cancel") + "\n")
	default:
		b.WriteString(helpStyle.Render("e:edit content  E:edit desc  r:reschedule  x:complete  d:delete  alt+m:move  q/esc:back") + "\n")
	}

	return style.Render(b.String())
}

func (m model) View() string {
	if (m.mode == modeDetail || m.mode == modeDetailEditContent || m.mode == modeDetailEditDesc ||
		m.mode == modeDetailReschedule || m.mode == modeDetailMove || m.mode == modeDetailConfirmDelete) &&
		len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
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
		main.WriteString(titleStyle.Render("◆ Search: "+m.searchQuery) + "  " + dimStyle.Render("(esc to return, J/K to browse projects)") + "\n\n")
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
		ghost := ""
		if c := m.addTaskCompletion(); c != "" {
			idx := strings.LastIndex(m.input, "#")
			fragment := m.input[idx+1:]
			ghost = dimStyle.Render(c[len(fragment):])
		}
		main.WriteString("\n" + titleStyle.Render("  New task: ") + m.input + ghost + "█\n")
	} else if m.mode == modeAddDesc {
		main.WriteString("\n" + dimStyle.Render("  Task: "+m.detailField) + "\n")
		main.WriteString(titleStyle.Render("  Description (enter to skip): ") + m.input + "█\n")
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
	} else if m.mode == modeReschedule && len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
		main.WriteString("\n" + titleStyle.Render("  Reschedule to: ") + m.input + "█\n")
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
			helpStyle.Render("j/k:tasks  J/K:projects  x:complete  d:delete  a:add  /:search  c:goto  alt+m:move  r:reschedule  alt+r:refresh  g/G:top/bottom  V:visual  q:quit"),
	)
	if m.mode == modeVisual {
		footer = footerStyle.Width(footerWidth - 2).Render(
			markedCursorStyle.Render(" VISUAL ") + "  " + statusRendered + "\n" +
				helpStyle.Render("j/k:extend selection  V/esc:exit  x:complete  d:delete  alt+m:move r:reschedule"),
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
	addTask := flag.String("a", "", "")
	flag.StringVar(addTask, "add-task", "", "Add a task and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: todoist-tui [-a, --add-task \"task text\"]\n")
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
