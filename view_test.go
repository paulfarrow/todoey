package main

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestView_ContainsTaskContent(t *testing.T) {
	m := testModel()
	m.width = 100
	m.height = 40
	v := m.View()
	if !strings.Contains(v, "Task 1") {
		t.Fatal("expected View to contain 'Task 1'")
	}
	if !strings.Contains(v, "Task 2") {
		t.Fatal("expected View to contain 'Task 2'")
	}
}

func TestView_ContainsProjects(t *testing.T) {
	m := testModel()
	m.width = 100
	m.height = 40
	v := m.View()
	if !strings.Contains(v, "Projects") {
		t.Fatal("expected View to contain 'Projects'")
	}
	if !strings.Contains(v, "Today") {
		t.Fatal("expected View to contain 'Today'")
	}
	if !strings.Contains(v, "Work") {
		t.Fatal("expected View to contain 'Work'")
	}
	if !strings.Contains(v, "Personal") {
		t.Fatal("expected View to contain 'Personal'")
	}
}

func TestView_EmptyTasks(t *testing.T) {
	m := testModel()
	m.tasks = nil
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "No tasks") {
		t.Fatal("expected 'No tasks' when task list is empty")
	}
}

func TestView_SearchHeader(t *testing.T) {
	m := testModel()
	m.searchQuery = "hello"
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "Search: hello") {
		t.Fatal("expected search header")
	}
}

func TestView_ProjectHeader(t *testing.T) {
	m := testModel()
	m.projectCursor = 1
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "Work") {
		t.Fatal("expected project name in header")
	}
}

func TestView_OverdueFilter(t *testing.T) {
	m := testModel()
	m.filterOverdue = true
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "Overdue only") {
		t.Fatal("expected overdue indicator")
	}
}

func TestView_ModeAdd(t *testing.T) {
	m := testModel()
	m.mode = modeAdd
	m.input.set("Buy")
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "New task:") {
		t.Fatal("expected 'New task:' prompt")
	}
}

func TestView_ModeAddDesc(t *testing.T) {
	m := testModel()
	m.mode = modeAddDesc
	m.detailField.set("My task")
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "Description") {
		t.Fatal("expected description prompt")
	}
}

func TestView_ModeSearch(t *testing.T) {
	m := testModel()
	m.mode = modeSearch
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "Search:") {
		t.Fatal("expected 'Search:' prompt")
	}
}

func TestView_ModeGoto(t *testing.T) {
	m := testModel()
	m.mode = modeGoto
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "Go to project:") {
		t.Fatal("expected 'Go to project:' prompt")
	}
}

func TestView_ModeConfirmDelete(t *testing.T) {
	m := testModel()
	m.mode = modeConfirmDelete
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "Delete") {
		t.Fatal("expected delete confirmation prompt")
	}
}

func TestView_ModeMoveTask(t *testing.T) {
	m := testModel()
	m.mode = modeMoveTask
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "Move to project:") {
		t.Fatal("expected 'Move to project:' prompt")
	}
}

func TestView_ModeReschedule(t *testing.T) {
	m := testModel()
	m.mode = modeReschedule
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "Reschedule to:") {
		t.Fatal("expected 'Reschedule to:' prompt")
	}
}

func TestView_ModeConfirmQuit(t *testing.T) {
	m := testModel()
	m.mode = modeConfirmQuit
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "Quit?") {
		t.Fatal("expected quit confirmation")
	}
}

func TestView_VisualFooter(t *testing.T) {
	m := testModel()
	m.mode = modeVisual
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "VISUAL") {
		t.Fatal("expected VISUAL indicator in footer")
	}
}

func TestView_FooterHelp(t *testing.T) {
	m := testModel()
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "j/k:tasks") {
		t.Fatal("expected help text in footer")
	}
}

func TestView_StatusError(t *testing.T) {
	m := testModel()
	m.status = "Error: something"
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "Error: something") {
		t.Fatal("expected error status in view")
	}
}

func TestViewDetail_Basic(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	m.width = 80
	v := m.View()
	if !strings.Contains(v, "Task 1") {
		t.Fatal("expected task content in detail view")
	}
	if !strings.Contains(v, "e:edit") {
		t.Fatal("expected help line in detail view")
	}
}

func TestViewDetail_WithDue(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	m.tasks[0].Due = &struct {
		Date        string `json:"date"`
		String      string `json:"string"`
		IsRecurring bool   `json:"is_recurring"`
	}{Date: "2024-06-01", String: "Jun 1"}
	m.width = 80
	v := m.View()
	if !strings.Contains(v, "Jun 1") {
		t.Fatal("expected due string in detail")
	}
	if !strings.Contains(v, "2024-06-01") {
		t.Fatal("expected due date in detail")
	}
}

func TestViewDetail_WithDescription(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	m.tasks[0].Description = "A detailed description"
	m.width = 80
	v := m.View()
	if !strings.Contains(v, "A detailed description") {
		t.Fatal("expected description in detail view")
	}
}

func TestViewDetail_WithLabels(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	m.tasks[0].Labels = []string{"urgent", "work"}
	m.width = 80
	v := m.View()
	if !strings.Contains(v, "urgent") || !strings.Contains(v, "work") {
		t.Fatal("expected labels in detail view")
	}
}

func TestViewDetail_WithPriority(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	m.tasks[0].Priority = 4
	m.width = 80
	v := m.View()
	if !strings.Contains(v, "Urgent") {
		t.Fatal("expected priority in detail view")
	}
}

func TestViewDetail_EditContentMode(t *testing.T) {
	m := testModel()
	m.mode = modeDetailEditContent
	m.detailField.set("editing")
	m.width = 80
	v := m.View()
	if !strings.Contains(v, "Edit content:") {
		t.Fatal("expected edit content prompt")
	}
}

func TestViewDetail_EditDescMode(t *testing.T) {
	m := testModel()
	m.mode = modeDetailEditDesc
	m.width = 80
	v := m.View()
	if !strings.Contains(v, "Edit description:") {
		t.Fatal("expected edit description prompt")
	}
}

func TestViewDetail_RescheduleMode(t *testing.T) {
	m := testModel()
	m.mode = modeDetailReschedule
	m.width = 80
	v := m.View()
	if !strings.Contains(v, "Reschedule to:") {
		t.Fatal("expected reschedule prompt")
	}
}

func TestViewDetail_MoveMode(t *testing.T) {
	m := testModel()
	m.mode = modeDetailMove
	m.detailField.set("wo")
	m.width = 80
	v := m.View()
	if !strings.Contains(v, "Move to project:") {
		t.Fatal("expected move prompt")
	}
}

func TestViewDetail_ConfirmDeleteMode(t *testing.T) {
	m := testModel()
	m.mode = modeDetailConfirmDelete
	m.width = 80
	v := m.View()
	if !strings.Contains(v, "Delete this task?") {
		t.Fatal("expected delete confirmation in detail")
	}
}

func TestViewDetail_WithDeadline(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	m.tasks[0].Deadline = &struct {
		Date string `json:"date"`
	}{Date: "2024-07-01"}
	m.width = 80
	v := m.View()
	if !strings.Contains(v, "2024-07-01") {
		t.Fatal("expected deadline in detail")
	}
}

func TestViewDetail_WithDuration(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	m.tasks[0].Duration = &struct {
		Amount int    `json:"amount"`
		Unit   string `json:"unit"`
	}{Amount: 30, Unit: "minute"}
	m.width = 80
	v := m.View()
	if !strings.Contains(v, "30 minute") {
		t.Fatal("expected duration in detail")
	}
}

func TestViewDetail_RecurringDue(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	m.tasks[0].Due = &struct {
		Date        string `json:"date"`
		String      string `json:"string"`
		IsRecurring bool   `json:"is_recurring"`
	}{Date: "2024-06-01", IsRecurring: true}
	m.width = 80
	v := m.View()
	if !strings.Contains(v, "🔁") {
		t.Fatal("expected recurring indicator")
	}
}

func TestView_ZeroWidth(t *testing.T) {
	m := testModel()
	m.width = 0
	// Should not panic
	v := m.View()
	if v == "" {
		t.Fatal("expected non-empty view even with zero width")
	}
}

func TestViewDetail_ZeroWidth(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	m.width = 0
	// Should default to 80 and not panic
	v := m.View()
	if v == "" {
		t.Fatal("expected non-empty detail view with zero width")
	}
}

func TestView_MultipleSelected(t *testing.T) {
	m := testModel()
	m.mode = modeConfirmDelete
	m.selected = map[int]bool{0: true, 1: true, 2: true}
	m.width = 100
	v := m.View()
	if !strings.Contains(v, "3 tasks") {
		t.Fatal("expected bulk delete prompt with count")
	}
}

func TestView_DueDate(t *testing.T) {
	m := testModel()
	m.tasks[0].Due = &struct {
		Date        string `json:"date"`
		String      string `json:"string"`
		IsRecurring bool   `json:"is_recurring"`
	}{Date: "2024-06-15"}
	m.width = 100
	v := m.View()
	// Now shows friendly format like "15 Jun 2024" or "Sat 15 Jun"
	if !strings.Contains(v, "Jun") {
		t.Fatal("expected friendly due date in task list")
	}
}

// --- Scroll tests ---

func TestEnsureTaskVisible_ScrollsDown(t *testing.T) {
	m := testModel()
	m.height = 15 // small window
	m.taskCursor = 0
	m.taskScroll = 0
	// Simulate many tasks
	m.tasks = make([]task, 50)
	for i := range m.tasks {
		m.tasks[i] = task{ID: fmt.Sprintf("t%d", i), Content: fmt.Sprintf("Task %d", i), ProjectID: "p1"}
	}
	m.taskCursor = 20
	m.ensureTaskVisible()
	if m.taskScroll == 0 {
		t.Fatal("expected taskScroll to increase when cursor is past visible area")
	}
	if m.taskCursor < m.taskScroll || m.taskCursor >= m.taskScroll+m.taskListHeight() {
		t.Fatalf("cursor %d not in visible range [%d, %d)", m.taskCursor, m.taskScroll, m.taskScroll+m.taskListHeight())
	}
}

func TestEnsureTaskVisible_ScrollsUp(t *testing.T) {
	m := testModel()
	m.height = 15
	m.tasks = make([]task, 50)
	for i := range m.tasks {
		m.tasks[i] = task{ID: fmt.Sprintf("t%d", i), Content: fmt.Sprintf("Task %d", i), ProjectID: "p1"}
	}
	m.taskScroll = 30
	m.taskCursor = 5
	m.ensureTaskVisible()
	cursorLine := m.taskCursorLine()
	if m.taskScroll > cursorLine {
		t.Fatalf("expected taskScroll <= cursorLine %d, got %d", cursorLine, m.taskScroll)
	}
}

func TestEnsureProjVisible_ScrollsDown(t *testing.T) {
	m := testModel()
	m.height = 12
	m.projects = make([]project, 30)
	for i := range m.projects {
		m.projects[i] = project{ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("Project %d", i)}
	}
	m.projectCursor = 25
	m.ensureProjVisible()
	if m.projScroll == 0 {
		t.Fatal("expected projScroll to increase")
	}
	if m.projectCursor < m.projScroll || m.projectCursor >= m.projScroll+m.sidebarHeight() {
		t.Fatalf("cursor %d not in visible range [%d, %d)", m.projectCursor, m.projScroll, m.projScroll+m.sidebarHeight())
	}
}

func TestEnsureProjVisible_ScrollsUp(t *testing.T) {
	m := testModel()
	m.height = 12
	m.projects = make([]project, 30)
	for i := range m.projects {
		m.projects[i] = project{ID: fmt.Sprintf("p%d", i), Name: fmt.Sprintf("Project %d", i)}
	}
	m.projScroll = 20
	m.projectCursor = 2
	m.ensureProjVisible()
	if m.projScroll > 2 {
		t.Fatalf("expected projScroll <= 2, got %d", m.projScroll)
	}
}

func TestAvailableHeight(t *testing.T) {
	m := testModel()
	m.height = 30
	h := m.availableHeight()
	if h != 26 {
		t.Fatalf("expected 26, got %d", h)
	}
	m.height = 5
	h = m.availableHeight()
	if h < 5 {
		t.Fatalf("expected minimum 5, got %d", h)
	}
}

func TestTaskListHeight(t *testing.T) {
	m := testModel()
	m.height = 30
	h := m.taskListHeight()
	if h <= 0 {
		t.Fatalf("expected positive height, got %d", h)
	}
}

func TestSidebarHeight(t *testing.T) {
	m := testModel()
	m.height = 30
	h := m.sidebarHeight()
	if h <= 0 {
		t.Fatalf("expected positive height, got %d", h)
	}
}

func TestView_ScrollIndicators(t *testing.T) {
	m := testModel()
	m.height = 12
	m.width = 100
	m.tasks = make([]task, 50)
	for i := range m.tasks {
		m.tasks[i] = task{ID: fmt.Sprintf("t%d", i), Content: fmt.Sprintf("Task %d", i), ProjectID: "p1"}
	}
	m.taskScroll = 5
	m.taskCursor = 5
	v := m.View()
	if !strings.Contains(v, "▲ more") {
		t.Fatal("expected '▲ more' scroll indicator at top")
	}
	if !strings.Contains(v, "▼ more") {
		t.Fatal("expected '▼ more' scroll indicator at bottom")
	}
}

func TestWindowResize_AdjustsScroll(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.height = 30
	m.tasks = make([]task, 50)
	for i := range m.tasks {
		m.tasks[i] = task{ID: fmt.Sprintf("t%d", i), Content: fmt.Sprintf("Task %d", i), ProjectID: "p1"}
	}
	m.taskCursor = 40
	m.taskScroll = 35
	// Resize to smaller
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 15})
	rm := result.(model)
	if rm.taskCursor < rm.taskScroll || rm.taskCursor >= rm.taskScroll+rm.taskListHeight() {
		t.Fatalf("after resize cursor %d not visible in [%d, %d)", rm.taskCursor, rm.taskScroll, rm.taskScroll+rm.taskListHeight())
	}
}
