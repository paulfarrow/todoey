package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// mockAPI implements TodoistAPI for testing.
type mockAPI struct {
	projects       []project
	tasks          []task
	err            error
	deletedIDs     []string
	closedIDs      []string
	movedTasks     []mockMove
	rescheduled    []mockReschedule
	createdTexts   []string
	updatedTasks   []mockUpdate
	searchResults  []task
}

type mockMove struct {
	TaskID, ProjectID string
}

type mockReschedule struct {
	TaskID, DueString string
}

type mockUpdate struct {
	TaskID string
	Fields map[string]string
}

func (m *mockAPI) GetProjects() ([]project, error) { return m.projects, m.err }
func (m *mockAPI) GetTodayTasks() ([]task, error)   { return m.tasks, m.err }
func (m *mockAPI) GetTasksByProject(projectID string) ([]task, error) {
	var out []task
	for _, t := range m.tasks {
		if t.ProjectID == projectID {
			out = append(out, t)
		}
	}
	return out, m.err
}
func (m *mockAPI) DeleteTask(taskID string) error {
	m.deletedIDs = append(m.deletedIDs, taskID)
	return m.err
}
func (m *mockAPI) CloseTask(taskID string) error {
	m.closedIDs = append(m.closedIDs, taskID)
	return m.err
}
func (m *mockAPI) MoveTask(taskID, projectID string) error {
	m.movedTasks = append(m.movedTasks, mockMove{taskID, projectID})
	return m.err
}
func (m *mockAPI) SearchTasks(query string) ([]task, error) {
	if m.searchResults != nil {
		return m.searchResults, m.err
	}
	return m.tasks, m.err
}
func (m *mockAPI) RescheduleTask(taskID, dueString string) error {
	m.rescheduled = append(m.rescheduled, mockReschedule{taskID, dueString})
	return m.err
}
func (m *mockAPI) CreateTask(text string) (*task, error) {
	m.createdTexts = append(m.createdTexts, text)
	t := task{ID: "new-1", Content: text, ProjectID: "p1"}
	return &t, m.err
}
func (m *mockAPI) UpdateTask(taskID string, fields map[string]string) error {
	m.updatedTasks = append(m.updatedTasks, mockUpdate{taskID, fields})
	return m.err
}

func newMock() *mockAPI {
	return &mockAPI{
		projects: []project{{ID: "p1", Name: "Work"}, {ID: "p2", Name: "Personal"}},
		tasks: []task{
			{ID: "t1", Content: "Task 1", ProjectID: "p1"},
			{ID: "t2", Content: "Task 2", ProjectID: "p2"},
			{ID: "t3", Content: "Task 3", ProjectID: "p1"},
		},
	}
}

func setMock(m *mockAPI) func() {
	old := api
	api = m
	return func() { api = old }
}

func testModel() model {
	return model{
		cfg:      config{AutoRefresh: false},
		projects: []project{{ID: "p1", Name: "Work"}, {ID: "p2", Name: "Personal"}},
		tasks: []task{
			{ID: "t1", Content: "Task 1", ProjectID: "p1"},
			{ID: "t2", Content: "Task 2", ProjectID: "p2"},
			{ID: "t3", Content: "Task 3", ProjectID: "p1"},
		},
		mode:   modeNormal,
		status: "3 tasks",
	}
}

func TestVisualRange(t *testing.T) {
	r := visualRange(1, 3)
	if len(r) != 3 || !r[1] || !r[2] || !r[3] {
		t.Fatalf("expected {1,2,3}, got %v", r)
	}
	r = visualRange(3, 1)
	if len(r) != 3 || !r[1] || !r[2] || !r[3] {
		t.Fatalf("reversed range should also give {1,2,3}, got %v", r)
	}
}

func TestSelectedTasks_NoSelection(t *testing.T) {
	m := testModel()
	m.taskCursor = 1
	tasks := m.selectedTasks()
	if len(tasks) != 1 || tasks[0].ID != "t2" {
		t.Fatalf("expected task at cursor, got %v", tasks)
	}
}

func TestSelectedTasks_WithSelection(t *testing.T) {
	m := testModel()
	m.selected = map[int]bool{0: true, 2: true}
	tasks := m.selectedTasks()
	if len(tasks) != 2 || tasks[0].ID != "t1" || tasks[1].ID != "t3" {
		t.Fatalf("expected t1 and t3, got %v", tasks)
	}
}

func TestSelectedTasks_EmptyTasks(t *testing.T) {
	m := testModel()
	m.tasks = nil
	tasks := m.selectedTasks()
	if tasks != nil {
		t.Fatalf("expected nil, got %v", tasks)
	}
}

func TestOverdueOnly(t *testing.T) {
	m := testModel()
	m.tasks = []task{
		{ID: "t1", Due: &struct {
			Date        string `json:"date"`
			String      string `json:"string"`
			IsRecurring bool   `json:"is_recurring"`
		}{Date: "2020-01-01"}},
		{ID: "t2", Due: &struct {
			Date        string `json:"date"`
			String      string `json:"string"`
			IsRecurring bool   `json:"is_recurring"`
		}{Date: "2099-01-01"}},
		{ID: "t3"},
	}
	overdue := m.overdueOnly(m.tasks)
	if len(overdue) != 1 || overdue[0].ID != "t1" {
		t.Fatalf("expected only t1 overdue, got %v", overdue)
	}
}

func TestGotoCompletion(t *testing.T) {
	m := testModel()
	m.input.set("wo")
	if c := m.gotoCompletion(); c != "Work" {
		t.Fatalf("expected 'Work', got %q", c)
	}
	m.input.set("to")
	if c := m.gotoCompletion(); c != "Today" {
		t.Fatalf("expected 'Today', got %q", c)
	}
	m.input.set("")
	if c := m.gotoCompletion(); c != "" {
		t.Fatalf("expected empty, got %q", c)
	}
	m.input.set("xyz")
	if c := m.gotoCompletion(); c != "" {
		t.Fatalf("expected empty for no match, got %q", c)
	}
}

func TestAddTaskCompletion(t *testing.T) {
	m := testModel()
	m.input.set("Buy milk #wo")
	if c := m.addTaskCompletion(); c != "Work" {
		t.Fatalf("expected 'Work', got %q", c)
	}
	m.input.set("Buy milk")
	if c := m.addTaskCompletion(); c != "" {
		t.Fatalf("expected empty without #, got %q", c)
	}
	m.input.set("#")
	if c := m.addTaskCompletion(); c != "" {
		t.Fatalf("expected empty for bare #, got %q", c)
	}
}

func TestUpdate_WindowSize(t *testing.T) {
	m := testModel()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	rm := result.(model)
	if rm.width != 120 || rm.height != 40 {
		t.Fatalf("expected 120x40, got %dx%d", rm.width, rm.height)
	}
}

func TestUpdate_DataMsg(t *testing.T) {
	m := testModel()
	m.taskCursor = 10 // out of range
	newTasks := []task{{ID: "x1", Content: "New", ProjectID: "p1"}}
	result, _ := m.Update(dataMsg{tasks: newTasks, projectCursor: 0})
	rm := result.(model)
	if len(rm.tasks) != 1 || rm.tasks[0].ID != "x1" {
		t.Fatalf("expected new tasks, got %v", rm.tasks)
	}
	if rm.taskCursor != 0 {
		t.Fatalf("cursor should be clamped to 0, got %d", rm.taskCursor)
	}
	if rm.status != "1 tasks" {
		t.Fatalf("expected '1 tasks', got %q", rm.status)
	}
}

func TestUpdate_DataMsg_StaleDiscard(t *testing.T) {
	m := testModel()
	m.projectCursor = 1
	result, _ := m.Update(dataMsg{tasks: []task{{ID: "x"}}, projectCursor: 0})
	rm := result.(model)
	// Should discard stale data - tasks unchanged
	if len(rm.tasks) != 3 {
		t.Fatalf("stale data should be discarded, got %d tasks", len(rm.tasks))
	}
}

func TestUpdate_DataMsg_ProjectsUpdate(t *testing.T) {
	m := testModel()
	newProjects := []project{{ID: "p9", Name: "New Project"}}
	result, _ := m.Update(dataMsg{projects: newProjects, tasks: m.tasks, projectCursor: 0})
	rm := result.(model)
	if len(rm.projects) != 1 || rm.projects[0].Name != "New Project" {
		t.Fatalf("expected updated projects, got %v", rm.projects)
	}
}

func TestUpdate_ErrMsg(t *testing.T) {
	m := testModel()
	result, _ := m.Update(errMsg("something went wrong"))
	rm := result.(model)
	if rm.status != "something went wrong" {
		t.Fatalf("expected error in status, got %q", rm.status)
	}
}

func TestUpdate_StatusMsg(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	result, cmd := m.Update(statusMsg("Completed: Task 1"))
	rm := result.(model)
	if rm.status != "Completed: Task 1" {
		t.Fatalf("expected status update, got %q", rm.status)
	}
	if cmd == nil {
		t.Fatal("expected a refresh command")
	}
}

func TestUpdate_TickMsg_Current(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.cfg.AutoRefresh = true
	m.cfg.RefreshInterval = 60
	m.refreshGen = 5
	result, cmd := m.Update(tickMsg{gen: 5})
	rm := result.(model)
	if rm.refreshGen != 6 {
		t.Fatalf("expected refreshGen incremented to 6, got %d", rm.refreshGen)
	}
	if cmd == nil {
		t.Fatal("expected refresh command from tick")
	}
}

func TestUpdate_TickMsg_Stale(t *testing.T) {
	m := testModel()
	m.refreshGen = 5
	_, cmd := m.Update(tickMsg{gen: 3})
	if cmd != nil {
		t.Fatal("stale tick should not produce a command")
	}
}

func TestUpdate_ConfirmDelete_Yes(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeConfirmDelete
	m.taskCursor = 0
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	rm := result.(model)
	if rm.mode != modeNormal {
		t.Fatalf("expected modeNormal, got %d", rm.mode)
	}
	if cmd == nil {
		t.Fatal("expected delete command")
	}
}

func TestUpdate_ConfirmDelete_Cancel(t *testing.T) {
	m := testModel()
	m.mode = modeConfirmDelete
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	rm := result.(model)
	if rm.mode != modeNormal {
		t.Fatalf("expected modeNormal after cancel, got %d", rm.mode)
	}
}

func TestUpdate_ConfirmQuit_Yes(t *testing.T) {
	m := testModel()
	m.mode = modeConfirmQuit
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
}

func TestUpdate_ConfirmQuit_Cancel(t *testing.T) {
	m := testModel()
	m.mode = modeConfirmQuit
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	rm := result.(model)
	if rm.mode != modeNormal {
		t.Fatalf("expected modeNormal after quit cancel, got %d", rm.mode)
	}
}

func TestUpdate_DataMsg_SortsOnTodayView(t *testing.T) {
	m := testModel()
	m.projectCursor = 0
	// Tasks in reverse project order
	tasks := []task{
		{ID: "t1", Content: "A", ProjectID: "p2"},
		{ID: "t2", Content: "B", ProjectID: "p1"},
	}
	result, _ := m.Update(dataMsg{tasks: tasks, projectCursor: 0})
	rm := result.(model)
	// p1 (Work) should come before p2 (Personal) based on project order
	if rm.tasks[0].ProjectID != "p1" || rm.tasks[1].ProjectID != "p2" {
		t.Fatalf("expected sorted by project order, got %v, %v", rm.tasks[0].ProjectID, rm.tasks[1].ProjectID)
	}
}

func TestUpdate_DataMsg_FilterOverdue(t *testing.T) {
	m := testModel()
	m.filterOverdue = true
	tasks := []task{
		{ID: "t1", Due: &struct {
			Date        string `json:"date"`
			String      string `json:"string"`
			IsRecurring bool   `json:"is_recurring"`
		}{Date: "2020-01-01"}, ProjectID: "p1"},
		{ID: "t2", Due: &struct {
			Date        string `json:"date"`
			String      string `json:"string"`
			IsRecurring bool   `json:"is_recurring"`
		}{Date: "2099-12-31"}, ProjectID: "p1"},
	}
	result, _ := m.Update(dataMsg{tasks: tasks, projectCursor: 0})
	rm := result.(model)
	if len(rm.tasks) != 1 || rm.tasks[0].ID != "t1" {
		t.Fatalf("expected only overdue task, got %v", rm.tasks)
	}
	if len(rm.allTasks) != 2 {
		t.Fatalf("expected allTasks to have full set, got %d", len(rm.allTasks))
	}
}

func TestInit_WithAutoRefresh(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := model{cfg: config{AutoRefresh: true, RefreshInterval: 60}}
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected batch command with fetch and tick")
	}
}

func TestInit_WithoutAutoRefresh(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := model{cfg: config{AutoRefresh: false}}
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected fetchInitialData command")
	}
}

func TestRefreshTasks_TodayView(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.projectCursor = 0
	cmd := m.refreshTasks()
	if cmd == nil {
		t.Fatal("expected command")
	}
	// Execute the command to verify it works
	msg := cmd()
	if _, ok := msg.(dataMsg); !ok {
		if _, ok := msg.(errMsg); !ok {
			t.Fatalf("expected dataMsg or errMsg, got %T", msg)
		}
	}
}

func TestRefreshTasks_ProjectView(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.projectCursor = 1
	cmd := m.refreshTasks()
	if cmd == nil {
		t.Fatal("expected command")
	}
}

func TestRefreshTasks_SearchActive(t *testing.T) {
	m := testModel()
	m.cfg.AutoRefresh = true
	m.cfg.RefreshInterval = 60
	m.searchQuery = "something"
	cmd := m.refreshTasks()
	if cmd == nil {
		t.Fatal("expected tick command when search is active")
	}
	// Should return a tickMsg (deferred tick), not data
	// We can't easily inspect tea.Tick, but verify it's non-nil
}

func TestProjectTag(t *testing.T) {
	m := testModel()
	tag := m.projectTag("p1")
	if tag == "" {
		t.Fatal("expected non-empty project tag for p1")
	}
	tag = m.projectTag("nonexistent")
	if tag != "" {
		t.Fatalf("expected empty for unknown project, got %q", tag)
	}
}

func TestDueStr(t *testing.T) {
	tk := task{ID: "t1"}
	if dueStr(tk) != "" {
		t.Fatal("expected empty for no due")
	}
	tk.Due = &struct {
		Date        string `json:"date"`
		String      string `json:"string"`
		IsRecurring bool   `json:"is_recurring"`
	}{Date: "2024-06-01"}
	if dueStr(tk) != " (2024-06-01)" {
		t.Fatalf("expected ' (2024-06-01)', got %q", dueStr(tk))
	}
}

func TestUpdate_KeyMsg_RoutesToDetail(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	rm := result.(model)
	if rm.mode != modeNormal {
		t.Fatalf("expected q in detail to go to modeNormal, got %d", rm.mode)
	}
}

func TestUpdate_KeyMsg_RoutesToInput(t *testing.T) {
	m := testModel()
	m.mode = modeSearch
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	rm := result.(model)
	if rm.mode != modeNormal {
		t.Fatalf("expected esc in search to go to modeNormal, got %d", rm.mode)
	}
}

func TestUpdate_KeyMsg_RoutesToNormal(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeNormal
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	rm := result.(model)
	if rm.taskCursor != 1 {
		t.Fatalf("expected cursor to move to 1, got %d", rm.taskCursor)
	}
}

func TestUpdate_VisualMode_RoutesToNormal(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeVisual
	m.visualAnchor = 0
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	rm := result.(model)
	if rm.taskCursor != 1 {
		t.Fatalf("expected cursor to move in visual mode, got %d", rm.taskCursor)
	}
	if !rm.selected[0] || !rm.selected[1] {
		t.Fatal("expected visual selection to extend")
	}
}

// Ensure the mock satisfies the interface at compile time.
var _ TodoistAPI = (*mockAPI)(nil)
