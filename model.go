package main

import (
	"fmt"
	"slices"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

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
	modeConfirmQuit
)

type model struct {
	cfg           config
	projects      []project
	tasks         []task
	allTasks      []task // unfiltered tasks, set when filterOverdue is active
	projectCursor int
	taskCursor    int
	selected      map[int]bool
	visualAnchor  int
	refreshGen    int
	status        string
	input         textInput
	detailField   textInput
	mode          inputMode
	searchQuery   string
	filterOverdue bool
	width         int
	height        int
	taskScroll    int // first visible task index
	projScroll    int // first visible project index in sidebar
}

type tickMsg struct{ gen int }

type dataMsg struct {
	projects      []project
	tasks         []task
	projectCursor int // cursor value at time of fetch; -1 means ignore check
}
type statusMsg string
type errMsg string

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
		m.ensureTaskVisible()
		m.ensureProjVisible()

	case dataMsg:
		if msg.projects != nil {
			m.projects = msg.projects
		}
		if msg.projectCursor != -1 && msg.projectCursor != m.projectCursor {
			return m, nil // stale result, discard
		}
		tasks := msg.tasks
		if m.projectCursor == 0 {
			order := make(map[string]int, len(m.projects))
			for i, p := range m.projects {
				order[p.ID] = i
			}
			slices.SortStableFunc(tasks, func(a, b task) int {
				return order[a.ProjectID] - order[b.ProjectID]
			})
		}
		m.tasks = tasks
		if m.filterOverdue {
			m.allTasks = tasks
			m.tasks = m.overdueOnly(tasks)
		} else {
			m.allTasks = nil
		}
		if n := len(m.tasks); m.taskCursor >= n {
			m.taskCursor = max(0, n-1)
		}
		m.selected = nil
		m.status = fmt.Sprintf("%d tasks", len(m.tasks))
		m.ensureTaskVisible()
		m.ensureProjVisible()

	case statusMsg:
		m.status = string(msg)
		m.refreshGen++
		return m, m.refreshTasks()

	case tickMsg:
		if msg.gen == m.refreshGen {
			m.refreshGen++
			return m, m.refreshTasks()
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
		if m.mode == modeConfirmQuit {
			switch msg.String() {
			case "y", "Y":
				return m, tea.Quit
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

func (m model) overdueOnly(tasks []task) []task {
	today := time.Now().Format("2006-01-02")
	var out []task
	for _, t := range tasks {
		if t.Due != nil && t.Due.Date != "" && t.Due.Date < today {
			out = append(out, t)
		}
	}
	return out
}

func (m model) refreshTasks() tea.Cmd {
	if m.cfg.AutoRefresh && m.cfg.RefreshInterval > 0 && m.searchQuery != "" {
		gen := m.refreshGen
		return tea.Tick(time.Duration(m.cfg.RefreshInterval)*time.Second, func(time.Time) tea.Msg { return tickMsg{gen} })
	}
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
	lower := strings.ToLower(m.input.val())
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
	idx := strings.LastIndex(m.input.val(), "#")
	if idx < 0 {
		return ""
	}
	fragment := strings.ToLower(m.input.val()[idx+1:])
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

// availableHeight returns the number of lines available for content (excluding footer).
func (m model) availableHeight() int {
	if m.height == 0 {
		return 100 // no terminal size yet; don't constrain
	}
	h := m.height - 4 // footer takes ~4 lines (border + 2 content lines + spacing)
	if h < 5 {
		h = 5
	}
	return h
}

// taskListHeight returns lines available for the task list in the main pane.
func (m model) taskListHeight() int {
	h := m.availableHeight() - 4 // header (~3 lines) + 1 for overdue/prompt margin
	if m.filterOverdue {
		h -= 2
	}
	// Reserve space for scroll indicators if list is scrollable
	if len(m.tasks) > h {
		h -= 2 // ▲ and ▼ indicators
	}
	if h < 3 {
		h = 3
	}
	return h
}

// sidebarHeight returns lines available for project items in the sidebar.
func (m model) sidebarHeight() int {
	h := m.availableHeight() - 3 // header: blank line + title + divider
	// Reserve space for scroll indicators if list is scrollable
	totalItems := 1 + len(m.projects)
	if totalItems > h {
		h -= 2 // ▲ and ▼ indicators
	}
	if h < 3 {
		h = 3
	}
	return h
}

// ensureTaskVisible adjusts taskScroll so taskCursor is visible.
func (m *model) ensureTaskVisible() {
	visible := m.taskListHeight()
	if m.taskCursor < m.taskScroll {
		m.taskScroll = m.taskCursor
	} else if m.taskCursor >= m.taskScroll+visible {
		m.taskScroll = m.taskCursor - visible + 1
	}
	if m.taskScroll < 0 {
		m.taskScroll = 0
	}
}

// ensureProjVisible adjusts projScroll so projectCursor is visible.
func (m *model) ensureProjVisible() {
	visible := m.sidebarHeight()
	// projectCursor 0 = "Today" which is item index 0, projects are 1..N
	idx := m.projectCursor
	if idx < m.projScroll {
		m.projScroll = idx
	} else if idx >= m.projScroll+visible {
		m.projScroll = idx - visible + 1
	}
	if m.projScroll < 0 {
		m.projScroll = 0
	}
}
