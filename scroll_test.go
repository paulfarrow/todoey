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
		t.Logf("projScroll=%d, sidebarHeight=%d, projectCursor=%d, availableHeight=%d", m.projScroll, m.sidebarHeight(), m.projectCursor, m.availableHeight())
		t.Logf("total sidebar items: %d", 1+len(m.projects))
		t.Error("selected project not visible")
	}
}
