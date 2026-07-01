package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m model) viewDetail() string {
	t := m.tasks[m.taskCursor]

	w := m.width
	if w == 0 {
		w = 80
	}

	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("195")).Italic(true)

	divider := dividerStyle.Render(strings.Repeat("─", w-8)) + "\n"

	row := func(icon, label, value string) string {
		if value == "" {
			return ""
		}
		return labelStyle.Render(icon+" "+label) + "  " + valueStyle.Render(value) + "\n"
	}

	priorities := map[int]string{4: "● Urgent", 3: "● High", 2: "● Medium", 1: "● Normal"}
	priorityColors := map[int]string{4: "203", 3: "215", 2: "220", 1: "245"}

	var b strings.Builder

	b.WriteString(titleStyle.Render(t.Content) + "\n")
	b.WriteString(divider)

	if t.Description != "" {
		b.WriteString(descStyle.Render(t.Description) + "\n")
		b.WriteString(divider)
	}

	if t.Due != nil && t.Due.Date != "" {
		s := t.Due.Date
		if t.Due.String != "" {
			s = t.Due.String + " (" + t.Due.Date + ")"
		}
		if t.Due.IsRecurring {
			s += "  🔁"
		}
		b.WriteString(row("📅", "Due", s))
	}
	if t.Deadline != nil && t.Deadline.Date != "" {
		b.WriteString(row("⏳", "Deadline", t.Deadline.Date))
	}
	if t.Duration != nil {
		b.WriteString(row("⏱", "Duration", fmt.Sprintf("%d %s", t.Duration.Amount, t.Duration.Unit)))
	}
	if p, ok := priorities[t.Priority]; ok && t.Priority > 1 {
		colored := lipgloss.NewStyle().Foreground(lipgloss.Color(priorityColors[t.Priority])).Render(p)
		b.WriteString(labelStyle.Render("⚑ Priority") + "  " + colored + "\n")
	}
	if proj := m.projectTag(t.ProjectID); proj != "" {
		b.WriteString(labelStyle.Render("Project:") + "  " + proj + "\n")
	}
	if len(t.Labels) > 0 {
		labels := ""
		for _, l := range t.Labels {
			labels += dimStyle.Render("@") + valueStyle.Render(l) + "  "
		}
		b.WriteString(labelStyle.Render("🏷 Labels") + "  " + strings.TrimSpace(labels) + "\n")
	}

	// Sub-tasks
	if m.subTasks != nil {
		b.WriteString("\n")
		if len(m.subTasks) == 0 {
			b.WriteString(labelStyle.Render("📋 Sub-tasks") + "  " + dimStyle.Render("none (S to add)") + "\n")
		} else {
			b.WriteString(labelStyle.Render("📋 Sub-tasks") + "  " + dimStyle.Render("(j/k:select  x:complete  enter:open)") + "\n")
			for i, st := range m.subTasks {
				checkbox := "○"
				stStyle := valueStyle
				if st.Priority >= 3 {
					stStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("215"))
				}
				prefix := "    "
				if i == m.subTaskCursor {
					prefix = "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("78")).Render("▸") + " "
					stStyle = stStyle.Bold(true)
				}
				b.WriteString(prefix + dimStyle.Render(checkbox) + " " + stStyle.Render(st.Content))
				if st.Due != nil && st.Due.Date != "" {
					b.WriteString(" " + dimStyle.Render("("+st.Due.Date+")"))
				}
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	switch m.mode {
	case modeDetailEditContent:
		b.WriteString(titleStyle.Render("  Edit content: ") + m.detailField.view() + "\n")
	case modeDetailEditDesc:
		b.WriteString(titleStyle.Render("  Edit description: ") + m.detailField.view() + "\n")
	case modeDetailReschedule:
		b.WriteString(titleStyle.Render("  Reschedule to: ") + m.detailField.view() + "\n")
	case modeDetailMove:
		ghost := ""
		lower := strings.ToLower(m.detailField.val())
		if lower != "" {
			for _, p := range m.projects {
				if strings.HasPrefix(strings.ToLower(p.Name), lower) {
					ghost = dimStyle.Render(p.Name[len(m.detailField.val()):])
					break
				}
			}
		}
		b.WriteString(titleStyle.Render("  Move to project: ") + m.detailField.view() + ghost + "\n")
	case modeDetailConfirmDelete:
		b.WriteString(errorStyle.Render("  Delete this task? ") + normalStyle.Render("[y] yes  [any] cancel") + "\n")
	case modeDetailAddComment:
		b.WriteString(titleStyle.Render("  Add comment: ") + m.detailField.view() + "\n")
	case modeDetailViewComments:
		b.WriteString(titleStyle.Render("  Comments:") + "\n")
		if len(m.comments) == 0 {
			b.WriteString(dimStyle.Render("    No comments") + "\n")
		} else {
			for _, cm := range m.comments {
				b.WriteString(dimStyle.Render("    • ") + normalStyle.Render(cm.Content) + "\n")
			}
		}
		b.WriteString("\n" + helpStyle.Render("  A:add comment  q/esc:back") + "\n")
	default:
		b.WriteString(helpStyle.Render("e:edit  E:desc  r:reschedule  x:complete  X:complete parent  d:delete  alt+m:move  W:web  C:comments  A:comment  S:sub-task  q/esc:back") + "\n")
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("52")).
		Padding(1, 3).
		Width(w - 4).
		Render(b.String())
}

func (m model) View() string {
	if (m.mode == modeDetail || m.mode == modeDetailEditContent || m.mode == modeDetailEditDesc ||
		m.mode == modeDetailReschedule || m.mode == modeDetailMove || m.mode == modeDetailConfirmDelete ||
		m.mode == modeDetailAddComment || m.mode == modeDetailViewComments) &&
		len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
		return m.viewDetail()
	}
	sidebarWidth := 24

	// --- Sidebar with scrolling ---
	var sidebar strings.Builder
	sidebar.WriteString("\n" + titleStyle.Render("Projects") + "\n")
	sidebar.WriteString(dimStyle.Render(strings.Repeat("─", sidebarWidth-2)) + "\n")

	itemWidth := sidebarWidth - 2
	renderItem := func(label string, selected bool) string {
		if selected {
			return selectedStyle.Width(itemWidth).Render(" " + label)
		}
		return normalStyle.Render("  " + label)
	}

	// Build list of all sidebar items: "Today" + projects
	type sideItem struct {
		label    string
		selected bool
	}
	allItems := make([]sideItem, 0, 1+len(m.projects))
	allItems = append(allItems, sideItem{"Today", m.projectCursor == 0})
	for i, p := range m.projects {
		name := p.Name
		if len(name) > sidebarWidth-4 {
			name = name[:sidebarWidth-4]
		}
		allItems = append(allItems, sideItem{name, i+1 == m.projectCursor})
	}

	sideH := m.sidebarHeight()
	projEnd := m.projScroll + sideH
	if projEnd > len(allItems) {
		projEnd = len(allItems)
	}
	scrollable := len(allItems) > sideH
	if m.projScroll > 0 {
		sidebar.WriteString(dimStyle.Render("  ▲ more") + "\n")
	} else if scrollable {
		sidebar.WriteString("\n")
	}
	for i := m.projScroll; i < projEnd; i++ {
		sidebar.WriteString(renderItem(allItems[i].label, allItems[i].selected) + "\n")
	}
	if projEnd < len(allItems) {
		sidebar.WriteString(dimStyle.Render("  ▼ more") + "\n")
	} else if scrollable {
		sidebar.WriteString("\n")
	}

	// --- Main pane with scrolling ---
	var main strings.Builder
	mainWidth := m.width - sidebarWidth - 6
	if mainWidth < 20 {
		mainWidth = 60
	}
	if m.searchQuery != "" {
		main.WriteString("\n" + titleStyle.Render("Search: "+m.searchQuery) + "  " + dimStyle.Render("(q or esc to return, J/K to browse projects)") + "\n")
		main.WriteString(dimStyle.Render(strings.Repeat("─", mainWidth)) + "\n")
	} else if m.projectCursor == 0 {
		main.WriteString("\n" + titleStyle.Render("Today") + "\n")
		main.WriteString(dimStyle.Render(strings.Repeat("─", mainWidth)) + "\n")
	} else if m.projectCursor-1 < len(m.projects) {
		main.WriteString("\n" + titleStyle.Render(m.projects[m.projectCursor-1].Name) + "\n")
		main.WriteString(dimStyle.Render(strings.Repeat("─", mainWidth)) + "\n")
	}
	if len(m.tasks) == 0 {
		main.WriteString(dimStyle.Render("  No tasks") + "\n")
	}
	if m.filterOverdue {
		main.WriteString(errorStyle.Render("  ⚠ Overdue only") + "  " + dimStyle.Render("(esc to clear)") + "\n\n")
	}

	today := time.Now().Format("2006-01-02")

	// Build all task lines (including project group headers and section headers)
	var taskLines []string
	lastProjectID := ""
	lastSectionID := ""
	for i, t := range m.tasks {
		if m.projectCursor == 0 && t.ProjectID != lastProjectID {
			lastProjectID = t.ProjectID
			header := m.projectTag(t.ProjectID)
			if header == "" {
				header = dimStyle.Render("(no project)")
			}
			taskLines = append(taskLines, "", header)
		}
		// Section headers in project view
		if m.projectCursor != 0 && t.SectionID != lastSectionID {
			lastSectionID = t.SectionID
			sectionName := ""
			for _, s := range m.sections {
				if s.ID == t.SectionID {
					sectionName = s.Name
					break
				}
			}
			if sectionName != "" {
				taskLines = append(taskLines, "", dimStyle.Render("┄ ")+lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75")).Render(sectionName))
			} else if lastSectionID == "" && i > 0 {
				// No section (ungrouped) - only show divider if there were prior tasks
			}
		}
		due := dueStr(t)
		overdue := t.Due != nil && t.Due.Date != "" && t.Due.Date < today
		dueRendered := dimStyle.Render(due)
		if overdue {
			dueRendered = errorStyle.Render(due)
		}
		tag := ""
		if m.projectCursor != 0 {
			tag = m.projectTag(t.ProjectID)
			if tag != "" {
				tag = " " + tag
			}
		}
		// Priority indicator
		priorityPrefix := ""
		switch t.Priority {
		case 4:
			priorityPrefix = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render("⚑ ")
		case 3:
			priorityPrefix = lipgloss.NewStyle().Foreground(lipgloss.Color("215")).Render("⚑ ")
		case 2:
			priorityPrefix = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("⚑ ")
		}
		// Labels
		labelStr := ""
		if len(t.Labels) > 0 {
			for _, l := range t.Labels {
				labelStr += " " + dimStyle.Render("@"+l)
			}
		}
		// Deadline indicator
		deadlineStr := ""
		if t.Deadline != nil && t.Deadline.Date != "" {
			deadlineStr = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("215")).Render("⏳"+t.Deadline.Date)
		}
		// Duration indicator
		durationStr := ""
		if t.Duration != nil {
			durationStr = " " + dimStyle.Render(fmt.Sprintf("⏱%d%s", t.Duration.Amount, t.Duration.Unit[:1]))
		}
		// Sub-task indentation
		indent := "  "
		if t.ParentID != "" {
			indent = "    "
		}
		isCursor := i == m.taskCursor
		isMarked := m.selected[i]
		var line string
		switch {
		case isCursor && isMarked:
			line = markedCursorStyle.Render(indent+"▶ "+priorityPrefix+t.Content) + dueRendered + deadlineStr + durationStr + tag + labelStr
		case isCursor:
			line = selectedStyle.Render(indent+"○ "+priorityPrefix+t.Content) + dueRendered + deadlineStr + durationStr + tag + labelStr
		case isMarked:
			line = markedStyle.Render(indent+"● "+priorityPrefix+t.Content) + dueRendered + deadlineStr + durationStr + tag + labelStr
		default:
			line = normalStyle.Render(indent+"○ "+priorityPrefix+t.Content) + dueRendered + deadlineStr + durationStr + tag + labelStr
		}
		taskLines = append(taskLines, line)
	}

	// Scroll the line list
	taskH := m.taskListHeight()
	lineEnd := m.taskScroll + taskH
	if lineEnd > len(taskLines) {
		lineEnd = len(taskLines)
	}
	scrollStart := m.taskScroll
	if scrollStart > len(taskLines) {
		scrollStart = len(taskLines)
	}

	if scrollStart > 0 {
		main.WriteString(dimStyle.Render("  ▲ more tasks") + "\n")
	} else if m.taskTotalLines() > taskH {
		main.WriteString("\n") // placeholder to keep layout stable
	}
	for _, line := range taskLines[scrollStart:lineEnd] {
		main.WriteString(line + "\n")
	}
	if lineEnd < len(taskLines) {
		main.WriteString(dimStyle.Render("  ▼ more tasks") + "\n")
	} else if m.taskTotalLines() > taskH {
		main.WriteString("\n") // placeholder to keep layout stable
	}

	promptBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		MarginTop(1)

	if m.mode == modeAdd {
		ghost := ""
		if c := m.addTaskCompletion(); c != "" {
			idx := strings.LastIndex(m.input.val(), "#")
			fragment := m.input.val()[idx+1:]
			ghost = dimStyle.Render(c[len(fragment):])
		}
		main.WriteString(promptBox.Render(titleStyle.Render("New task: ") + m.input.view() + ghost) + "\n")
	} else if m.mode == modeAddDesc {
		main.WriteString(promptBox.Render(dimStyle.Render("Task: "+m.detailField.val()) + "\n" + titleStyle.Render("Description (enter to skip): ") + m.input.view()) + "\n")
	} else if m.mode == modeSearch {
		main.WriteString(promptBox.Render(titleStyle.Render("Search: ") + m.input.view()) + "\n")
	} else if m.mode == modeGoto {
		ghost := ""
		if c := m.gotoCompletion(); c != "" && c != m.input.val() {
			ghost = dimStyle.Render(c[len(m.input.val()):])
		}
		main.WriteString(promptBox.Render(titleStyle.Render("Go to project: ") + m.input.view() + ghost) + "\n")
	} else if m.mode == modeConfirmDelete && len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
		var deletePrompt string
		if len(m.selected) > 1 {
			deletePrompt = fmt.Sprintf("Delete %d tasks? ", len(m.selected))
		} else {
			deletePrompt = "Delete \"" + m.tasks[m.taskCursor].Content + "\"? "
		}
		main.WriteString(promptBox.Render(errorStyle.Render(deletePrompt) + normalStyle.Render("[y] yes  [any] cancel")) + "\n")
	} else if m.mode == modeMoveTask && len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
		ghost := ""
		if c := m.gotoCompletion(); c != "" && c != m.input.val() {
			ghost = dimStyle.Render(c[len(m.input.val()):])
		}
		main.WriteString(promptBox.Render(titleStyle.Render("Move to project: ") + m.input.view() + ghost) + "\n")
	} else if m.mode == modeReschedule && len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
		main.WriteString(promptBox.Render(titleStyle.Render("Reschedule to: ") + m.input.view()) + "\n")
	} else if m.mode == modeAddSubTask && len(m.tasks) > 0 && m.taskCursor < len(m.tasks) {
		main.WriteString(promptBox.Render(titleStyle.Render("Sub-task for \""+m.tasks[m.taskCursor].Content+"\": ") + m.input.view()) + "\n")
	} else if m.mode == modeConfirmQuit {
		main.WriteString(promptBox.Render(errorStyle.Render("Quit? ") + normalStyle.Render("[y] yes  [any] cancel")) + "\n")
	}

	contentHeight := m.availableHeight()
	sideRendered := sidebarStyle.Width(sidebarWidth).Height(contentHeight).MaxHeight(contentHeight).Render(sidebar.String())
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
			helpStyle.Render("j/k:tasks  J/K:projects  x:complete  d:delete  a:add  S:sub-task  /:search  c:goto  alt+m:move  r:reschedule  O:overdue  alt+r:refresh  g/G:top/bottom  {/}:project group  V:visual  q/esc:back/quit"),
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
	if t.Due == nil || t.Due.Date == "" {
		return ""
	}
	now := time.Now()
	today := now.Format("2006-01-02")
	tomorrow := now.AddDate(0, 0, 1).Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	dateStr := t.Due.Date
	// Extract just the date part if it contains time
	datePart := dateStr
	if len(dateStr) > 10 {
		datePart = dateStr[:10]
	}

	// Friendly name
	var friendly string
	switch datePart {
	case today:
		friendly = "Today"
	case tomorrow:
		friendly = "Tomorrow"
	case yesterday:
		friendly = "Yesterday"
	default:
		// Use the Due.String if available and short enough
		if t.Due.String != "" && len(t.Due.String) <= 20 {
			friendly = t.Due.String
		} else {
			// Parse and format as "Mon 2 Jan"
			if parsed, err := time.Parse("2006-01-02", datePart); err == nil {
				if parsed.Year() == now.Year() {
					friendly = parsed.Format("Mon 2 Jan")
				} else {
					friendly = parsed.Format("2 Jan 2006")
				}
			} else {
				friendly = datePart
			}
		}
	}

	// Add time if present
	if len(dateStr) > 10 {
		// Format: 2006-01-02T15:04:05 or 2006-01-02T15:04:05Z
		if parsed, err := time.Parse("2006-01-02T15:04:05", dateStr[:19]); err == nil {
			friendly += " " + parsed.Format("15:04")
		} else if parsed, err := time.Parse("2006-01-02T15:04:05Z", dateStr[:20]); err == nil {
			friendly += " " + parsed.Format("15:04")
		}
	}

	// Recurring indicator
	if t.Due.IsRecurring {
		friendly += " 🔁"
	}

	return " (" + friendly + ")"
}
