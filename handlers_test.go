package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// --- handleNormal tests ---

func TestHandleNormal_JDown(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	rm := result.(model)
	if rm.taskCursor != 1 {
		t.Fatalf("expected cursor=1, got %d", rm.taskCursor)
	}
}

func TestHandleNormal_JAtBottom(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.taskCursor = 2
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	rm := result.(model)
	if rm.taskCursor != 2 {
		t.Fatalf("expected cursor to stay at 2, got %d", rm.taskCursor)
	}
}

func TestHandleNormal_KUp(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.taskCursor = 2
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	rm := result.(model)
	if rm.taskCursor != 1 {
		t.Fatalf("expected cursor=1, got %d", rm.taskCursor)
	}
}

func TestHandleNormal_KAtTop(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	rm := result.(model)
	if rm.taskCursor != 0 {
		t.Fatalf("expected cursor to stay at 0, got %d", rm.taskCursor)
	}
}

func TestHandleNormal_GTop(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.taskCursor = 2
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	rm := result.(model)
	if rm.taskCursor != 0 {
		t.Fatalf("expected cursor=0, got %d", rm.taskCursor)
	}
}

func TestHandleNormal_GBottom(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	rm := result.(model)
	if rm.taskCursor != 2 {
		t.Fatalf("expected cursor=2, got %d", rm.taskCursor)
	}
}

func TestHandleNormal_JNextProject(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.projectCursor = 0
	result, cmd := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})
	rm := result.(model)
	if rm.projectCursor != 1 {
		t.Fatalf("expected projectCursor=1, got %d", rm.projectCursor)
	}
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
}

func TestHandleNormal_JAtLastProject(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.projectCursor = 2 // already at last
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})
	rm := result.(model)
	if rm.projectCursor != 2 {
		t.Fatalf("expected projectCursor to stay at 2, got %d", rm.projectCursor)
	}
}

func TestHandleNormal_KPrevProject(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.projectCursor = 1
	result, cmd := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})
	rm := result.(model)
	if rm.projectCursor != 0 {
		t.Fatalf("expected projectCursor=0, got %d", rm.projectCursor)
	}
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
}

func TestHandleNormal_T(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.projectCursor = 2
	result, cmd := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'T'}})
	rm := result.(model)
	if rm.projectCursor != 0 {
		t.Fatalf("expected projectCursor=0, got %d", rm.projectCursor)
	}
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
}

func TestHandleNormal_EnterDetail(t *testing.T) {
	m := testModel()
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyEnter})
	rm := result.(model)
	if rm.mode != modeDetail {
		t.Fatalf("expected modeDetail, got %d", rm.mode)
	}
}

func TestHandleNormal_EnterNoTasks(t *testing.T) {
	m := testModel()
	m.tasks = nil
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyEnter})
	rm := result.(model)
	if rm.mode != modeNormal {
		t.Fatalf("expected mode to stay modeNormal with no tasks, got %d", rm.mode)
	}
}

func TestHandleNormal_A(t *testing.T) {
	m := testModel()
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	rm := result.(model)
	if rm.mode != modeAdd {
		t.Fatalf("expected modeAdd, got %d", rm.mode)
	}
}

func TestHandleNormal_Slash(t *testing.T) {
	m := testModel()
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	rm := result.(model)
	if rm.mode != modeSearch {
		t.Fatalf("expected modeSearch, got %d", rm.mode)
	}
}

func TestHandleNormal_C(t *testing.T) {
	m := testModel()
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	rm := result.(model)
	if rm.mode != modeGoto {
		t.Fatalf("expected modeGoto, got %d", rm.mode)
	}
}

func TestHandleNormal_D(t *testing.T) {
	m := testModel()
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	rm := result.(model)
	if rm.mode != modeConfirmDelete {
		t.Fatalf("expected modeConfirmDelete, got %d", rm.mode)
	}
}

func TestHandleNormal_DNoTasks(t *testing.T) {
	m := testModel()
	m.tasks = nil
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	rm := result.(model)
	if rm.mode != modeNormal {
		t.Fatalf("expected mode to stay modeNormal with no tasks, got %d", rm.mode)
	}
}

func TestHandleNormal_X(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	_, cmd := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if cmd == nil {
		t.Fatal("expected close command")
	}
}

func TestHandleNormal_XNoTasks(t *testing.T) {
	m := testModel()
	m.tasks = nil
	_, cmd := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if cmd != nil {
		t.Fatal("expected no command with no tasks")
	}
}

func TestHandleNormal_R(t *testing.T) {
	m := testModel()
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	rm := result.(model)
	if rm.mode != modeReschedule {
		t.Fatalf("expected modeReschedule, got %d", rm.mode)
	}
}

func TestHandleNormal_Q(t *testing.T) {
	m := testModel()
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	rm := result.(model)
	if rm.mode != modeConfirmQuit {
		t.Fatalf("expected modeConfirmQuit, got %d", rm.mode)
	}
}

func TestHandleNormal_Q_WithSearch(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.searchQuery = "something"
	result, cmd := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	rm := result.(model)
	// Should behave like esc - clear search
	if rm.searchQuery != "" {
		t.Fatalf("expected search cleared, got %q", rm.searchQuery)
	}
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
}

func TestHandleNormal_V_EnterVisual(t *testing.T) {
	m := testModel()
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	rm := result.(model)
	if rm.mode != modeVisual {
		t.Fatalf("expected modeVisual, got %d", rm.mode)
	}
	if !rm.selected[0] {
		t.Fatal("expected current task selected")
	}
}

func TestHandleNormal_V_ExitVisual(t *testing.T) {
	m := testModel()
	m.mode = modeVisual
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'V'}})
	rm := result.(model)
	if rm.mode != modeNormal {
		t.Fatalf("expected modeNormal after V toggle, got %d", rm.mode)
	}
}

func TestHandleNormal_VisualJ(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeVisual
	m.visualAnchor = 0
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	rm := result.(model)
	if rm.taskCursor != 1 {
		t.Fatalf("expected cursor=1, got %d", rm.taskCursor)
	}
	if !rm.selected[0] || !rm.selected[1] {
		t.Fatal("expected visual selection to include 0 and 1")
	}
}

func TestHandleNormal_SpaceToggle(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeySpace})
	rm := result.(model)
	// After space, cursor advances and task 0 should have been toggled on then cursor moved
	if rm.taskCursor != 1 {
		t.Fatalf("expected cursor=1 after space, got %d", rm.taskCursor)
	}
}

func TestHandleNormal_O_ToggleOverdue(t *testing.T) {
	m := testModel()
	m.tasks = []task{
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
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'O'}})
	rm := result.(model)
	if !rm.filterOverdue {
		t.Fatal("expected filterOverdue=true")
	}
	if len(rm.tasks) != 1 {
		t.Fatalf("expected 1 overdue task, got %d", len(rm.tasks))
	}
	// Toggle back
	result, _ = rm.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'O'}})
	rm = result.(model)
	if rm.filterOverdue {
		t.Fatal("expected filterOverdue=false")
	}
	if len(rm.tasks) != 2 {
		t.Fatalf("expected 2 tasks restored, got %d", len(rm.tasks))
	}
}

func TestHandleNormal_Esc_ClearsOverdue(t *testing.T) {
	m := testModel()
	m.filterOverdue = true
	m.allTasks = m.tasks
	m.tasks = m.tasks[:1]
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyEsc})
	rm := result.(model)
	if rm.filterOverdue {
		t.Fatal("expected overdue cleared")
	}
	if len(rm.tasks) != 3 {
		t.Fatalf("expected tasks restored, got %d", len(rm.tasks))
	}
}

func TestHandleNormal_Esc_ClearsSearch(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.searchQuery = "hello"
	result, cmd := m.handleNormal(tea.KeyMsg{Type: tea.KeyEsc})
	rm := result.(model)
	if rm.searchQuery != "" {
		t.Fatalf("expected search cleared, got %q", rm.searchQuery)
	}
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
}

func TestHandleNormal_Esc_ClearsSelection(t *testing.T) {
	m := testModel()
	m.selected = map[int]bool{0: true, 1: true}
	result, _ := m.handleNormal(tea.KeyMsg{Type: tea.KeyEsc})
	rm := result.(model)
	if rm.selected != nil {
		t.Fatal("expected selection cleared")
	}
}

func TestHandleNormal_AltR(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	result, cmd := m.handleNormal(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}, Alt: true})
	rm := result.(model)
	if rm.status != "Refreshing..." {
		t.Fatalf("expected 'Refreshing...', got %q", rm.status)
	}
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
}

// --- handleInput tests ---

func TestHandleInput_Esc(t *testing.T) {
	m := testModel()
	m.mode = modeSearch
	m.input.set("hello")
	result, _ := m.handleInput(tea.KeyMsg{Type: tea.KeyEsc})
	rm := result.(model)
	if rm.mode != modeNormal {
		t.Fatalf("expected modeNormal, got %d", rm.mode)
	}
	if rm.input.val() != "" {
		t.Fatal("expected input cleared")
	}
}

func TestHandleInput_Search_Enter(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeSearch
	m.input.set("find me")
	result, cmd := m.handleInput(tea.KeyMsg{Type: tea.KeyEnter})
	rm := result.(model)
	if rm.searchQuery != "find me" {
		t.Fatalf("expected searchQuery='find me', got %q", rm.searchQuery)
	}
	if cmd == nil {
		t.Fatal("expected search command")
	}
}

func TestHandleInput_Search_EnterEmpty(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeSearch
	m.input.set("  ")
	result, cmd := m.handleInput(tea.KeyMsg{Type: tea.KeyEnter})
	rm := result.(model)
	if rm.searchQuery != "" {
		t.Fatalf("expected empty searchQuery, got %q", rm.searchQuery)
	}
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
}

func TestHandleInput_Goto_Today(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeGoto
	m.projectCursor = 2
	m.input.set("today")
	result, cmd := m.handleInput(tea.KeyMsg{Type: tea.KeyEnter})
	rm := result.(model)
	if rm.projectCursor != 0 {
		t.Fatalf("expected projectCursor=0, got %d", rm.projectCursor)
	}
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
}

func TestHandleInput_Goto_Project(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeGoto
	m.input.set("Work")
	result, cmd := m.handleInput(tea.KeyMsg{Type: tea.KeyEnter})
	rm := result.(model)
	if rm.projectCursor != 1 {
		t.Fatalf("expected projectCursor=1, got %d", rm.projectCursor)
	}
	if cmd == nil {
		t.Fatal("expected refresh command")
	}
}

func TestHandleInput_Tab_GotoCompletion(t *testing.T) {
	m := testModel()
	m.mode = modeGoto
	m.input.set("wo")
	result, _ := m.handleInput(tea.KeyMsg{Type: tea.KeyTab})
	rm := result.(model)
	if rm.input.val() != "Work" {
		t.Fatalf("expected 'Work', got %q", rm.input.val())
	}
}

func TestHandleInput_Tab_AddCompletion(t *testing.T) {
	m := testModel()
	m.mode = modeAdd
	m.input.set("Buy milk #wo")
	result, _ := m.handleInput(tea.KeyMsg{Type: tea.KeyTab})
	rm := result.(model)
	if rm.input.val() != "Buy milk #Work" {
		t.Fatalf("expected 'Buy milk #Work', got %q", rm.input.val())
	}
}

func TestHandleInput_MoveTask(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeMoveTask
	m.input.set("Personal")
	_, cmd := m.handleInput(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected move command")
	}
}

func TestHandleInput_Reschedule(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeReschedule
	m.input.set("tomorrow")
	_, cmd := m.handleInput(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected reschedule command")
	}
}

func TestHandleInput_Add_Enter(t *testing.T) {
	m := testModel()
	m.mode = modeAdd
	m.input.set("New task")
	result, _ := m.handleInput(tea.KeyMsg{Type: tea.KeyEnter})
	rm := result.(model)
	if rm.mode != modeAddDesc {
		t.Fatalf("expected modeAddDesc, got %d", rm.mode)
	}
	if rm.detailField.val() != "New task" {
		t.Fatalf("expected detailField='New task', got %q", rm.detailField.val())
	}
}

func TestHandleInput_AddDesc_Enter(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeAddDesc
	m.detailField.set("New task")
	m.input.set("A description")
	_, cmd := m.handleInput(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected create command")
	}
}

func TestHandleInput_AddDesc_EnterEmpty(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeAddDesc
	m.detailField.set("New task")
	m.input.set("")
	_, cmd := m.handleInput(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected create command even with empty desc")
	}
}

func TestHandleInput_CharPassthrough(t *testing.T) {
	m := testModel()
	m.mode = modeSearch
	m.input.clear()
	result, _ := m.handleInput(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	rm := result.(model)
	if rm.input.val() != "a" {
		t.Fatalf("expected 'a' in input, got %q", rm.input.val())
	}
}

// --- handleDetail tests ---

func TestHandleDetail_Esc(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	result, _ := m.handleDetail(tea.KeyMsg{Type: tea.KeyEsc})
	rm := result.(model)
	if rm.mode != modeNormal {
		t.Fatalf("expected modeNormal, got %d", rm.mode)
	}
}

func TestHandleDetail_Q(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	result, _ := m.handleDetail(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	rm := result.(model)
	if rm.mode != modeNormal {
		t.Fatalf("expected modeNormal, got %d", rm.mode)
	}
}

func TestHandleDetail_E(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	result, _ := m.handleDetail(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	rm := result.(model)
	if rm.mode != modeDetailEditContent {
		t.Fatalf("expected modeDetailEditContent, got %d", rm.mode)
	}
	if rm.detailField.val() != "Task 1" {
		t.Fatalf("expected detailField='Task 1', got %q", rm.detailField.val())
	}
}

func TestHandleDetail_BigE(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	m.tasks[0].Description = "My desc"
	result, _ := m.handleDetail(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'E'}})
	rm := result.(model)
	if rm.mode != modeDetailEditDesc {
		t.Fatalf("expected modeDetailEditDesc, got %d", rm.mode)
	}
	if rm.detailField.val() != "My desc" {
		t.Fatalf("expected 'My desc', got %q", rm.detailField.val())
	}
}

func TestHandleDetail_R(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	result, _ := m.handleDetail(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	rm := result.(model)
	if rm.mode != modeDetailReschedule {
		t.Fatalf("expected modeDetailReschedule, got %d", rm.mode)
	}
}

func TestHandleDetail_X(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeDetail
	result, cmd := m.handleDetail(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	rm := result.(model)
	if rm.mode != modeNormal {
		t.Fatalf("expected modeNormal, got %d", rm.mode)
	}
	if cmd == nil {
		t.Fatal("expected close command")
	}
}

func TestHandleDetail_D(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	result, _ := m.handleDetail(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	rm := result.(model)
	if rm.mode != modeDetailConfirmDelete {
		t.Fatalf("expected modeDetailConfirmDelete, got %d", rm.mode)
	}
}

func TestHandleDetail_ConfirmDelete_Y(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeDetailConfirmDelete
	result, cmd := m.handleDetail(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	rm := result.(model)
	if rm.mode != modeNormal {
		t.Fatalf("expected modeNormal, got %d", rm.mode)
	}
	if cmd == nil {
		t.Fatal("expected delete command")
	}
}

func TestHandleDetail_ConfirmDelete_Cancel(t *testing.T) {
	m := testModel()
	m.mode = modeDetailConfirmDelete
	result, _ := m.handleDetail(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	rm := result.(model)
	if rm.mode != modeDetail {
		t.Fatalf("expected modeDetail after cancel, got %d", rm.mode)
	}
}

func TestHandleDetail_EditContent_Enter(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeDetailEditContent
	m.detailField.set("Updated content")
	result, cmd := m.handleDetail(tea.KeyMsg{Type: tea.KeyEnter})
	rm := result.(model)
	if rm.mode != modeDetail {
		t.Fatalf("expected modeDetail after edit, got %d", rm.mode)
	}
	if cmd == nil {
		t.Fatal("expected update command")
	}
}

func TestHandleDetail_EditContent_EnterEmpty(t *testing.T) {
	m := testModel()
	m.mode = modeDetailEditContent
	m.detailField.set("  ")
	result, cmd := m.handleDetail(tea.KeyMsg{Type: tea.KeyEnter})
	rm := result.(model)
	if rm.mode != modeDetail {
		t.Fatalf("expected modeDetail, got %d", rm.mode)
	}
	if cmd != nil {
		t.Fatal("expected no command for empty content")
	}
}

func TestHandleDetail_EditContent_Esc(t *testing.T) {
	m := testModel()
	m.mode = modeDetailEditContent
	m.detailField.set("something")
	result, _ := m.handleDetail(tea.KeyMsg{Type: tea.KeyEsc})
	rm := result.(model)
	if rm.mode != modeDetail {
		t.Fatalf("expected modeDetail, got %d", rm.mode)
	}
	if rm.detailField.val() != "" {
		t.Fatal("expected detailField cleared on esc")
	}
}

func TestHandleDetail_EditDesc_Enter(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeDetailEditDesc
	m.detailField.set("new description")
	_, cmd := m.handleDetail(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected update command for description")
	}
}

func TestHandleDetail_Reschedule_Enter(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeDetailReschedule
	m.detailField.set("next week")
	_, cmd := m.handleDetail(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected reschedule command")
	}
}

func TestHandleDetail_Move_Enter(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	m := testModel()
	m.mode = modeDetailMove
	m.detailField.set("Personal")
	_, cmd := m.handleDetail(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected move command")
	}
}

func TestHandleDetail_Move_Tab(t *testing.T) {
	m := testModel()
	m.mode = modeDetailMove
	m.detailField.set("wo")
	result, _ := m.handleDetail(tea.KeyMsg{Type: tea.KeyTab})
	rm := result.(model)
	if rm.detailField.val() != "Work" {
		t.Fatalf("expected 'Work', got %q", rm.detailField.val())
	}
}

func TestHandleDetail_Move_EnterNoMatch(t *testing.T) {
	m := testModel()
	m.mode = modeDetailMove
	m.detailField.set("Nonexistent")
	result, cmd := m.handleDetail(tea.KeyMsg{Type: tea.KeyEnter})
	rm := result.(model)
	if rm.mode != modeDetail {
		t.Fatalf("expected modeDetail, got %d", rm.mode)
	}
	if cmd != nil {
		t.Fatal("expected no command for non-matching project")
	}
}

func TestHandleDetail_EditContent_CharPassthrough(t *testing.T) {
	m := testModel()
	m.mode = modeDetailEditContent
	m.detailField.clear()
	result, _ := m.handleDetail(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	rm := result.(model)
	if rm.detailField.val() != "x" {
		t.Fatalf("expected 'x', got %q", rm.detailField.val())
	}
}

func TestHandleDetail_AltM(t *testing.T) {
	m := testModel()
	m.mode = modeDetail
	result, _ := m.handleDetail(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}, Alt: true})
	rm := result.(model)
	if rm.mode != modeDetailMove {
		t.Fatalf("expected modeDetailMove, got %d", rm.mode)
	}
}
