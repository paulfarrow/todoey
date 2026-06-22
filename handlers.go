package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) handleInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if m.mode == modeGoto || m.mode == modeMoveTask {
			if c := m.gotoCompletion(); c != "" {
				m.input.set(c)
			}
		}
		if m.mode == modeAdd {
			if c := m.addTaskCompletion(); c != "" {
				idx := strings.LastIndex(m.input.val(), "#")
				m.input.set(m.input.val()[:idx+1] + c)
			}
		}
	case "enter":
		text := strings.TrimSpace(m.input.val())
		origMode := m.mode
		m.mode = modeNormal
		m.input.clear()
		switch origMode {
		case modeSearch:
			if text != "" {
				m.searchQuery = text
				m.status = "Searching..."
				return m, searchTasks(text)
			}
			m.searchQuery = ""
			m.refreshGen++
			return m, m.refreshTasks()
		case modeGoto:
			lower := strings.ToLower(text)
			if lower == "today" {
				m.projectCursor = 0
				m.searchQuery = ""
				m.status = "Loading..."
				m.ensureProjVisible()
				m.refreshGen++
				return m, m.refreshTasks()
			}
			for i, p := range m.projects {
				if strings.ToLower(p.Name) == lower {
					m.projectCursor = i + 1
					m.searchQuery = ""
					m.status = "Loading..."
					m.ensureProjVisible()
					m.refreshGen++
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
				m.detailField.set(text)
				m.input.clear()
				return m, nil
			}
		case modeAddDesc:
			content := m.detailField.val()
			m.detailField.clear()
			if text != "" {
				content += " // " + text
			}
			return m, createTask(content)
		}
	case "esc", "ctrl+c":
		m.mode = modeNormal
		m.input.clear()
		m.detailField.clear()
	default:
		m.input.handle(msg)
	}
	return m, nil
}

func (m model) handleNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		if !m.filterOverdue && m.searchQuery == "" && m.mode == modeNormal && len(m.selected) == 0 {
			m.mode = modeConfirmQuit
			return m, nil
		}
		return m.handleNormal(tea.KeyMsg{Type: tea.KeyEsc})

	case "V":
		if len(m.tasks) > 0 {
			if m.mode == modeVisual {
				m.mode = modeNormal
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
			m.ensureTaskVisible()
		}
	case "k", "up":
		if m.taskCursor > 0 {
			m.taskCursor--
			if m.mode == modeVisual {
				m.selected = visualRange(m.visualAnchor, m.taskCursor)
			}
			m.ensureTaskVisible()
		}

	case "J":
		if m.projectCursor < len(m.projects) {
			m.projectCursor++
			m.searchQuery = ""
			m.status = "Loading..."
			m.taskScroll = 0
			m.ensureProjVisible()
			m.refreshGen++
			return m, m.refreshTasks()
		}
	case "K":
		if m.projectCursor > 0 {
			m.projectCursor--
			m.searchQuery = ""
			m.status = "Loading..."
			m.taskScroll = 0
			m.ensureProjVisible()
			m.refreshGen++
			return m, m.refreshTasks()
		}

	case "T":
		m.projectCursor = 0
		m.searchQuery = ""
		m.status = "Loading..."
		m.taskScroll = 0
		m.ensureProjVisible()
		m.refreshGen++
		return m, m.refreshTasks()

	case "O":
		if m.filterOverdue {
			m.filterOverdue = false
			if m.allTasks != nil {
				m.tasks = m.allTasks
				m.allTasks = nil
			}
		} else {
			m.filterOverdue = true
			m.allTasks = m.tasks
			m.tasks = m.overdueOnly(m.allTasks)
		}
		m.taskCursor = 0
		m.taskScroll = 0
		m.selected = nil
		m.mode = modeNormal

	case "g":
		m.taskCursor = 0
		m.ensureTaskVisible()
	case "G":
		if len(m.tasks) > 0 {
			m.taskCursor = len(m.tasks) - 1
			m.ensureTaskVisible()
		}

	case "}":
		// Jump to first task of next project group (Today view)
		if m.projectCursor == 0 && len(m.tasks) > 0 {
			curPID := m.tasks[m.taskCursor].ProjectID
			for i := m.taskCursor + 1; i < len(m.tasks); i++ {
				if m.tasks[i].ProjectID != curPID {
					m.taskCursor = i
					m.ensureTaskVisible()
					break
				}
			}
		}
	case "{":
		// Jump to first task of previous project group (Today view)
		if m.projectCursor == 0 && len(m.tasks) > 0 && m.taskCursor > 0 {
			curPID := m.tasks[m.taskCursor].ProjectID
			// Find start of current group
			groupStart := m.taskCursor
			for groupStart > 0 && m.tasks[groupStart-1].ProjectID == curPID {
				groupStart--
			}
			if groupStart > 0 {
				// Jump to start of previous group
				prevPID := m.tasks[groupStart-1].ProjectID
				dest := groupStart - 1
				for dest > 0 && m.tasks[dest-1].ProjectID == prevPID {
					dest--
				}
				m.taskCursor = dest
			} else {
				m.taskCursor = 0
			}
			m.ensureTaskVisible()
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
			m.ensureTaskVisible()
		}
	case "esc", "ctrl+c":
		if m.filterOverdue {
			m.filterOverdue = false
			if m.allTasks != nil {
				m.tasks = m.allTasks
				m.allTasks = nil
			}
			m.taskCursor = 0
			m.taskScroll = 0
			m.selected = nil
			return m, nil
		}
		if m.searchQuery != "" {
			m.searchQuery = ""
			m.status = "Loading..."
			m.refreshGen++
			return m, m.refreshTasks()
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
		m.input.clear()
	case "/":
		m.mode = modeSearch
		m.input.clear()
	case "c":
		m.mode = modeGoto
		m.input.clear()
	case "alt+m":
		if len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
			m.mode = modeMoveTask
			m.input.clear()
		}
	case "W":
		if len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
			openTaskInBrowser(m.tasks[m.taskCursor].ID)
			m.status = "Opened in browser"
		}
	case "r":
		if len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
			m.mode = modeReschedule
			m.input.clear()
		}
	case "alt+r":
		m.searchQuery = ""
		m.status = "Refreshing..."
		m.refreshGen++
		return m, m.refreshTasks()
	}
	return m, nil
}

func (m model) handleDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	t := m.tasks[m.taskCursor]

	if m.mode == modeDetailEditContent || m.mode == modeDetailEditDesc ||
		m.mode == modeDetailReschedule || m.mode == modeDetailMove {
		switch msg.String() {
		case "esc":
			m.mode = modeDetail
			m.detailField.clear()
		case "enter":
			text := strings.TrimSpace(m.detailField.val())
			origMode := m.mode
			m.mode = modeDetail
			m.detailField.clear()
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
				lower := strings.ToLower(m.detailField.val())
				for _, p := range m.projects {
					if strings.HasPrefix(strings.ToLower(p.Name), lower) {
						m.detailField.set(p.Name)
						break
					}
				}
			}
		default:
			m.detailField.handle(msg)
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

	switch msg.String() {
	case "esc", "q":
		m.mode = modeNormal
	case "e":
		m.mode = modeDetailEditContent
		m.detailField.set(t.Content)
	case "E":
		m.mode = modeDetailEditDesc
		m.detailField.set(t.Description)
	case "r":
		m.mode = modeDetailReschedule
		m.detailField.clear()
	case "x":
		m.mode = modeNormal
		return m, closeTask(t.ID, t.Content)
	case "d":
		m.mode = modeDetailConfirmDelete
	case "W":
		openTaskInBrowser(t.ID)
		m.status = "Opened in browser"
	case "alt+m":
		m.mode = modeDetailMove
		m.detailField.clear()
	}
	return m, nil
}
