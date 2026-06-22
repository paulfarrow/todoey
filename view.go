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
	default:
		b.WriteString(helpStyle.Render("e:edit content  E:edit desc  r:reschedule  x:complete  d:delete  alt+m:move  W:open in browser  q/esc:back") + "\n")
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
		m.mode == modeDetailReschedule || m.mode == modeDetailMove || m.mode == modeDetailConfirmDelete) &&
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

	// Build all task lines (including project group headers)
	var taskLines []string
	lastProjectID := ""
	for i, t := range m.tasks {
		if m.projectCursor == 0 && t.ProjectID != lastProjectID {
			lastProjectID = t.ProjectID
			header := m.projectTag(t.ProjectID)
			if header == "" {
				header = dimStyle.Render("(no project)")
			}
			taskLines = append(taskLines, "", header)
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
			helpStyle.Render("j/k:tasks  J/K:projects  x:complete  d:delete  a:add  /:search  c:goto  alt+m:move  r:reschedule  O:overdue  alt+r:refresh  g/G:top/bottom  {/}:hop_project V:visual  q/esc:back/quit"),
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
