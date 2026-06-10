package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestScrollRendering_SmallWindow(t *testing.T) {
	m := model{
		cfg:      config{AutoRefresh: false},
		projects: []project{{ID: "p1", Name: "Work"}, {ID: "p2", Name: "Personal"}, {ID: "p3", Name: "Shopping"}, {ID: "p4", Name: "Fitness"}, {ID: "p5", Name: "Reading"}},
		mode:     modeNormal,
		width:    60,
		height:   15,
	}
	for i := 0; i < 20; i++ {
		m.tasks = append(m.tasks, task{ID: fmt.Sprintf("t%d", i), Content: fmt.Sprintf("Task number %d", i), ProjectID: "p1"})
	}
	m.taskCursor = 5
	m.ensureTaskVisible()
	m.ensureProjVisible()

	v := m.View()

	if !strings.Contains(v, "Projects") {
		t.Error("'Projects' header missing from sidebar")
	}
	if !strings.Contains(v, "Today") {
		t.Error("'Today' missing from sidebar")
	}
	if !strings.Contains(v, "Task number 5") {
		t.Error("cursor task (Task number 5) not visible")
	}
	if !strings.Contains(v, "▼ more") {
		t.Error("expected ▼ scroll indicator")
	}

	// Verify the output doesn't contain tasks that should be scrolled off
	taskH := m.taskListHeight()
	for i := m.taskScroll + taskH; i < len(m.tasks); i++ {
		if strings.Contains(v, fmt.Sprintf("Task number %d", i)) {
			t.Errorf("Task number %d should be scrolled off but is visible", i)
		}
	}
}

func TestScrollRendering_CursorAtBottom(t *testing.T) {
	m := model{
		cfg:      config{AutoRefresh: false},
		projects: []project{{ID: "p1", Name: "Work"}},
		mode:     modeNormal,
		width:    60,
		height:   15,
	}
	for i := 0; i < 30; i++ {
		m.tasks = append(m.tasks, task{ID: fmt.Sprintf("t%d", i), Content: fmt.Sprintf("Task %d", i), ProjectID: "p1"})
	}
	m.taskCursor = 29
	m.ensureTaskVisible()

	v := m.View()
	if !strings.Contains(v, "Task 29") {
		t.Error("last task (Task 29) not visible when cursor is at bottom")
	}
	if !strings.Contains(v, "▲ more") {
		t.Error("expected ▲ scroll indicator when scrolled down")
	}
}

func TestScrollRendering_TodayMultipleProjects(t *testing.T) {
	m := model{
		cfg:      config{AutoRefresh: false},
		projects: []project{{ID: "p1", Name: "Work"}, {ID: "p2", Name: "Personal"}, {ID: "p3", Name: "Shopping"}},
		mode:     modeNormal,
		width:    80,
		height:   12,
	}
	// 3 tasks per project = 9 tasks + 6 header lines (2 per group) = 15 total lines
	for i := 0; i < 3; i++ {
		m.tasks = append(m.tasks, task{ID: fmt.Sprintf("w%d", i), Content: fmt.Sprintf("Work task %d", i), ProjectID: "p1"})
	}
	for i := 0; i < 3; i++ {
		m.tasks = append(m.tasks, task{ID: fmt.Sprintf("p%d", i), Content: fmt.Sprintf("Personal task %d", i), ProjectID: "p2"})
	}
	for i := 0; i < 3; i++ {
		m.tasks = append(m.tasks, task{ID: fmt.Sprintf("s%d", i), Content: fmt.Sprintf("Shopping task %d", i), ProjectID: "p3"})
	}

	// Cursor on last task of last project
	m.taskCursor = 8 // "Shopping task 2"
	m.ensureTaskVisible()

	v := m.View()
	if !strings.Contains(v, "Shopping task 2") {
		t.Error("cursor task 'Shopping task 2' not visible")
	}

	// Now cursor at first task, scroll should go back to top
	m.taskCursor = 0
	m.ensureTaskVisible()
	v = m.View()
	if !strings.Contains(v, "Work task 0") {
		t.Error("first task 'Work task 0' not visible after scrolling up")
	}
	if !strings.Contains(v, "Work") {
		t.Error("Work project header not visible")
	}
	// ▲ should NOT appear when scrolled to top
	if strings.Contains(v, "▲ more tasks") {
		t.Error("'▲ more tasks' should not appear when at top of list")
	}
}

func TestScrollRendering_HeadersStayFixed(t *testing.T) {
	m := model{
		cfg:      config{AutoRefresh: false},
		projects: []project{{ID: "p1", Name: "Work"}},
		mode:     modeNormal,
		width:    80,
		height:   15,
	}
	for i := 0; i < 30; i++ {
		m.tasks = append(m.tasks, task{ID: fmt.Sprintf("t%d", i), Content: fmt.Sprintf("Task %d", i), ProjectID: "p1"})
	}

	// At the top: no ▲, has ▼
	m.taskCursor = 0
	m.ensureTaskVisible()
	v1 := m.View()

	// In the middle: has both ▲ and ▼
	m.taskCursor = 15
	m.ensureTaskVisible()
	v2 := m.View()

	// At the bottom: has ▲, no ▼
	m.taskCursor = 29
	m.ensureTaskVisible()
	v3 := m.View()

	// Find "Today" header line position in each view - should be at the same position
	lines1 := strings.Split(v1, "\n")
	lines2 := strings.Split(v2, "\n")
	lines3 := strings.Split(v3, "\n")

	findHeader := func(lines []string) int {
		for i, l := range lines {
			if strings.Contains(l, "Today") && !strings.Contains(l, "sidebar") {
				return i
			}
		}
		return -1
	}

	h1 := findHeader(lines1)
	h2 := findHeader(lines2)
	h3 := findHeader(lines3)

	if h1 != h2 || h2 != h3 {
		t.Errorf("'Today' header moved: top=%d, middle=%d, bottom=%d", h1, h2, h3)
	}
}

func TestScrollRendering_TodayOverdueWithHeaders(t *testing.T) {
	m := model{
		cfg:           config{AutoRefresh: false},
		projects:      []project{{ID: "p1", Name: "Work"}, {ID: "p2", Name: "Personal"}},
		mode:          modeNormal,
		width:         80,
		height:        12,
		filterOverdue: true,
	}
	// Overdue tasks across 2 projects
	for i := 0; i < 5; i++ {
		m.tasks = append(m.tasks, task{
			ID: fmt.Sprintf("w%d", i), Content: fmt.Sprintf("Overdue work %d", i), ProjectID: "p1",
			Due: &struct {
				Date        string `json:"date"`
				String      string `json:"string"`
				IsRecurring bool   `json:"is_recurring"`
			}{Date: "2020-01-01"},
		})
	}
	for i := 0; i < 5; i++ {
		m.tasks = append(m.tasks, task{
			ID: fmt.Sprintf("p%d", i), Content: fmt.Sprintf("Overdue personal %d", i), ProjectID: "p2",
			Due: &struct {
				Date        string `json:"date"`
				String      string `json:"string"`
				IsRecurring bool   `json:"is_recurring"`
			}{Date: "2020-01-01"},
		})
	}
	m.taskCursor = 9 // last task
	m.ensureTaskVisible()

	v := m.View()
	if !strings.Contains(v, "Overdue personal 4") {
		t.Error("last overdue task not visible")
	}
}

func TestScrollRendering_SidebarScrolled(t *testing.T) {
	m := model{
		cfg:    config{AutoRefresh: false},
		mode:   modeNormal,
		width:  60,
		height: 12,
	}
	for i := 0; i < 20; i++ {
		m.projects = append(m.projects, project{ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("Project %d", i)})
	}
	m.tasks = []task{{ID: "t1", Content: "Test", ProjectID: "p0"}}
	m.projectCursor = 15
	m.ensureProjVisible()

	v := m.View()
	if !strings.Contains(v, "Projects") {
		t.Error("'Projects' header must always be visible")
	}
	if !strings.Contains(v, "Project 14") {
		t.Error("selected project not visible")
	}
}
