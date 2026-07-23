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
	modeDetailAddComment
	modeDetailViewComments
	modeAddSubTask
	modeAddDesc
	modeConfirmQuit
)

type model struct {
	cfg           config
	projects      []project
	tasks         []task
	allTasks      []task // unfiltered tasks, set when filterOverdue is active
	sections      []section
	comments      []comment
	subTasks      []task // sub-tasks of the currently viewed task in detail mode
	subTaskCursor int    // selected sub-task index in detail view
	detailTask    *task  // task currently shown in detail view (nil = use tasks[taskCursor])
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
	countBuf      int // vim-style numeric prefix accumulator
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

	case sectionsMsg:
		m.sections = msg.sections

	case commentsMsg:
		m.comments = msg.comments
		m.mode = modeDetailViewComments

	case subTasksMsg:
		m.subTasks = msg.subTasks

	case tea.KeyMsg:
		if m.mode == modeDetail ||
			m.mode == modeDetailEditContent ||
			m.mode == modeDetailEditDesc ||
			m.mode == modeDetailReschedule ||
			m.mode == modeDetailMove ||
			m.mode == modeDetailConfirmDelete ||
			m.mode == modeDetailAddComment ||
			m.mode == modeDetailViewComments {
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

// currentDetailTask returns the task being shown in the detail view.
func (m model) currentDetailTask() task {
	if m.detailTask != nil {
		return *m.detailTask
	}
	return m.tasks[m.taskCursor]
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
		projID := m.projects[m.projectCursor-1].ID
		fetch = tea.Batch(fetchProjectTasks(projID, m.projectCursor), fetchSections(projID))
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
	fragment := m.input.val()[idx+1:]
	if fragment == "" {
		return ""
	}
	// Unescape backslash-spaces for matching against project names
	unescaped := strings.ReplaceAll(fragment, "\\ ", " ")
	lower := strings.ToLower(unescaped)
	for _, p := range m.projects {
		if strings.HasPrefix(strings.ToLower(p.Name), lower) {
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
	h := m.availableHeight()
	h -= 3 // header: blank line + title + divider
	if m.filterOverdue {
		h -= 2 // overdue indicator + blank line
	}
	// Reserve space for scroll indicators only if there's room
	totalLines := m.taskTotalLines()
	if totalLines > h && h > 4 {
		h -= 2 // ▲ and ▼ indicators
	}
	if h < 3 {
		h = 3
	}
	return h
}

// taskTotalLines returns the total number of lines the task list would occupy
// including project group headers on the Today view and section headers in project views.
func (m model) taskTotalLines() int {
	if len(m.tasks) == 0 {
		return 0
	}
	lines := len(m.tasks)
	if m.projectCursor == 0 {
		lastPID := ""
		for _, t := range m.tasks {
			if t.ProjectID != lastPID {
				lastPID = t.ProjectID
				lines += 2 // blank line + header
			}
		}
	} else {
		lastSID := ""
		for _, t := range m.tasks {
			if t.SectionID != lastSID {
				lastSID = t.SectionID
				// Only add lines if there's a named section
				for _, s := range m.sections {
					if s.ID == t.SectionID {
						lines += 2 // blank line + section header
						break
					}
				}
			}
		}
	}
	return lines
}

// taskCursorLine returns the line index (0-based) of the current taskCursor
// within the full task line list (accounting for project headers and section headers).
func (m model) taskCursorLine() int {
	line := 0
	lastPID := ""
	lastSID := ""
	for i, t := range m.tasks {
		if m.projectCursor == 0 && t.ProjectID != lastPID {
			lastPID = t.ProjectID
			line += 2 // blank line + header
		}
		if m.projectCursor != 0 && t.SectionID != lastSID {
			lastSID = t.SectionID
			for _, s := range m.sections {
				if s.ID == t.SectionID {
					line += 2
					break
				}
			}
		}
		if i == m.taskCursor {
			return line
		}
		line++
	}
	return line
}

// taskCursorLineWithHeader returns the line where the cursor's section starts
// (including the project header or section header above the first task in a group).
func (m model) taskCursorLineWithHeader() int {
	line := 0
	lastPID := ""
	lastSID := ""
	headerLine := 0
	for i, t := range m.tasks {
		if m.projectCursor == 0 && t.ProjectID != lastPID {
			lastPID = t.ProjectID
			headerLine = line
			line += 2
		}
		if m.projectCursor != 0 && t.SectionID != lastSID {
			prevSID := lastSID
			lastSID = t.SectionID
			for _, s := range m.sections {
				if s.ID == t.SectionID {
					headerLine = line
					line += 2
					break
				}
			}
			_ = prevSID
		}
		if i == m.taskCursor {
			if i == 0 || (m.projectCursor == 0 && m.tasks[i-1].ProjectID != t.ProjectID) {
				return headerLine
			}
			if m.projectCursor != 0 && (i == 0 || m.tasks[i-1].SectionID != t.SectionID) {
				return headerLine
			}
			return line
		}
		line++
	}
	return line
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

// ensureTaskVisible adjusts taskScroll (line offset) so taskCursor is visible.
func (m *model) ensureTaskVisible() {
	visible := m.taskListHeight()
	cursorLine := m.taskCursorLine()
	if cursorLine < m.taskScroll {
		// Scroll up: include the group header if cursor is first in group
		m.taskScroll = m.taskCursorLineWithHeader()
	} else if cursorLine >= m.taskScroll+visible {
		m.taskScroll = cursorLine - visible + 1
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
