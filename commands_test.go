package main

import (
	"errors"
	"testing"
)

func TestFetchInitialData_Success(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	cmd := fetchInitialData()
	msg := cmd()
	dm, ok := msg.(dataMsg)
	if !ok {
		t.Fatalf("expected dataMsg, got %T", msg)
	}
	if len(dm.projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(dm.projects))
	}
	if len(dm.tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(dm.tasks))
	}
	if dm.projectCursor != 0 {
		t.Fatalf("expected projectCursor=0, got %d", dm.projectCursor)
	}
}

func TestFetchInitialData_ProjectsError(t *testing.T) {
	mock := &mockAPI{err: errors.New("network failure")}
	defer setMock(mock)()
	cmd := fetchInitialData()
	msg := cmd()
	em, ok := msg.(errMsg)
	if !ok {
		t.Fatalf("expected errMsg, got %T", msg)
	}
	if string(em) != "Error: network failure" {
		t.Fatalf("unexpected error: %s", em)
	}
}

func TestFetchTodayTasks_Success(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	cmd := fetchTodayTasks(0)
	msg := cmd()
	dm, ok := msg.(dataMsg)
	if !ok {
		t.Fatalf("expected dataMsg, got %T", msg)
	}
	if len(dm.tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(dm.tasks))
	}
	if dm.projectCursor != 0 {
		t.Fatalf("expected projectCursor=0, got %d", dm.projectCursor)
	}
}

func TestFetchTodayTasks_Error(t *testing.T) {
	mock := &mockAPI{err: errors.New("fail")}
	defer setMock(mock)()
	cmd := fetchTodayTasks(1)
	msg := cmd()
	if _, ok := msg.(errMsg); !ok {
		t.Fatalf("expected errMsg, got %T", msg)
	}
}

func TestFetchProjectTasks_Success(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	cmd := fetchProjectTasks("p1", 1)
	msg := cmd()
	dm, ok := msg.(dataMsg)
	if !ok {
		t.Fatalf("expected dataMsg, got %T", msg)
	}
	// mockAPI.GetTasksByProject filters by projectID
	if len(dm.tasks) != 2 {
		t.Fatalf("expected 2 tasks for p1, got %d", len(dm.tasks))
	}
	if dm.projectCursor != 1 {
		t.Fatalf("expected projectCursor=1, got %d", dm.projectCursor)
	}
}

func TestFetchProjectTasks_Error(t *testing.T) {
	mock := &mockAPI{err: errors.New("fail")}
	defer setMock(mock)()
	cmd := fetchProjectTasks("p1", 1)
	msg := cmd()
	if _, ok := msg.(errMsg); !ok {
		t.Fatalf("expected errMsg, got %T", msg)
	}
}

func TestMoveTaskCmd_Success(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	cmd := moveTask("t1", "Task 1", "p2")
	msg := cmd()
	sm, ok := msg.(statusMsg)
	if !ok {
		t.Fatalf("expected statusMsg, got %T", msg)
	}
	if string(sm) != "Moved: Task 1" {
		t.Fatalf("unexpected status: %s", sm)
	}
	if len(mock.movedTasks) != 1 || mock.movedTasks[0].TaskID != "t1" || mock.movedTasks[0].ProjectID != "p2" {
		t.Fatalf("unexpected move call: %v", mock.movedTasks)
	}
}

func TestMoveTaskCmd_Error(t *testing.T) {
	mock := &mockAPI{err: errors.New("move failed")}
	defer setMock(mock)()
	cmd := moveTask("t1", "Task 1", "p2")
	msg := cmd()
	em, ok := msg.(errMsg)
	if !ok {
		t.Fatalf("expected errMsg, got %T", msg)
	}
	if string(em) != "Error moving: move failed" {
		t.Fatalf("unexpected error: %s", em)
	}
}

func TestSearchTasksCmd_Success(t *testing.T) {
	mock := newMock()
	mock.searchResults = []task{{ID: "s1", Content: "Found"}}
	defer setMock(mock)()
	cmd := searchTasks("found")
	msg := cmd()
	dm, ok := msg.(dataMsg)
	if !ok {
		t.Fatalf("expected dataMsg, got %T", msg)
	}
	if len(dm.tasks) != 1 || dm.tasks[0].ID != "s1" {
		t.Fatalf("unexpected search results: %v", dm.tasks)
	}
	if dm.projectCursor != -1 {
		t.Fatalf("expected projectCursor=-1, got %d", dm.projectCursor)
	}
}

func TestSearchTasksCmd_Error(t *testing.T) {
	mock := &mockAPI{err: errors.New("search fail")}
	defer setMock(mock)()
	cmd := searchTasks("query")
	msg := cmd()
	if _, ok := msg.(errMsg); !ok {
		t.Fatalf("expected errMsg, got %T", msg)
	}
}

func TestDeleteTaskCmd_Success(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	cmd := deleteTask("t1", "Task 1")
	msg := cmd()
	sm, ok := msg.(statusMsg)
	if !ok {
		t.Fatalf("expected statusMsg, got %T", msg)
	}
	if string(sm) != "Deleted: Task 1" {
		t.Fatalf("unexpected status: %s", sm)
	}
	if len(mock.deletedIDs) != 1 || mock.deletedIDs[0] != "t1" {
		t.Fatalf("unexpected delete call: %v", mock.deletedIDs)
	}
}

func TestDeleteTaskCmd_Error(t *testing.T) {
	mock := &mockAPI{err: errors.New("del fail")}
	defer setMock(mock)()
	cmd := deleteTask("t1", "Task 1")
	msg := cmd()
	em, ok := msg.(errMsg)
	if !ok {
		t.Fatalf("expected errMsg, got %T", msg)
	}
	if string(em) != "Error deleting: del fail" {
		t.Fatalf("unexpected: %s", em)
	}
}

func TestCloseTaskCmd_Success(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	cmd := closeTask("t2", "Task 2")
	msg := cmd()
	sm, ok := msg.(statusMsg)
	if !ok {
		t.Fatalf("expected statusMsg, got %T", msg)
	}
	if string(sm) != "Completed: Task 2" {
		t.Fatalf("unexpected status: %s", sm)
	}
	if len(mock.closedIDs) != 1 || mock.closedIDs[0] != "t2" {
		t.Fatalf("unexpected close call: %v", mock.closedIDs)
	}
}

func TestCloseTaskCmd_Error(t *testing.T) {
	mock := &mockAPI{err: errors.New("close fail")}
	defer setMock(mock)()
	cmd := closeTask("t2", "Task 2")
	msg := cmd()
	if _, ok := msg.(errMsg); !ok {
		t.Fatalf("expected errMsg, got %T", msg)
	}
}

func TestRescheduleTaskCmd_Success(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	cmd := rescheduleTask("t1", "Task 1", "tomorrow")
	msg := cmd()
	sm, ok := msg.(statusMsg)
	if !ok {
		t.Fatalf("expected statusMsg, got %T", msg)
	}
	if string(sm) != "Rescheduled: Task 1" {
		t.Fatalf("unexpected status: %s", sm)
	}
	if len(mock.rescheduled) != 1 || mock.rescheduled[0].DueString != "tomorrow" {
		t.Fatalf("unexpected reschedule call: %v", mock.rescheduled)
	}
}

func TestRescheduleTaskCmd_Error(t *testing.T) {
	mock := &mockAPI{err: errors.New("resched fail")}
	defer setMock(mock)()
	cmd := rescheduleTask("t1", "Task 1", "tomorrow")
	msg := cmd()
	if _, ok := msg.(errMsg); !ok {
		t.Fatalf("expected errMsg, got %T", msg)
	}
}

func TestUpdateTaskCmd_Success(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	cmd := updateTask("t1", "Task 1", map[string]string{"content": "Updated"})
	msg := cmd()
	sm, ok := msg.(statusMsg)
	if !ok {
		t.Fatalf("expected statusMsg, got %T", msg)
	}
	if string(sm) != "Updated: Task 1" {
		t.Fatalf("unexpected status: %s", sm)
	}
	if len(mock.updatedTasks) != 1 || mock.updatedTasks[0].Fields["content"] != "Updated" {
		t.Fatalf("unexpected update call: %v", mock.updatedTasks)
	}
}

func TestUpdateTaskCmd_Error(t *testing.T) {
	mock := &mockAPI{err: errors.New("update fail")}
	defer setMock(mock)()
	cmd := updateTask("t1", "Task 1", map[string]string{"content": "x"})
	msg := cmd()
	if _, ok := msg.(errMsg); !ok {
		t.Fatalf("expected errMsg, got %T", msg)
	}
}

func TestCreateTaskCmd_Success(t *testing.T) {
	mock := newMock()
	defer setMock(mock)()
	cmd := createTask("New task")
	msg := cmd()
	sm, ok := msg.(statusMsg)
	if !ok {
		t.Fatalf("expected statusMsg, got %T", msg)
	}
	if string(sm) != "Added: New task" {
		t.Fatalf("unexpected status: %s", sm)
	}
	if len(mock.createdTexts) != 1 || mock.createdTexts[0] != "New task" {
		t.Fatalf("unexpected create call: %v", mock.createdTexts)
	}
}

func TestCreateTaskCmd_Error(t *testing.T) {
	mock := &mockAPI{err: errors.New("create fail")}
	defer setMock(mock)()
	cmd := createTask("New task")
	msg := cmd()
	em, ok := msg.(errMsg)
	if !ok {
		t.Fatalf("expected errMsg, got %T", msg)
	}
	if string(em) != "Error adding: create fail" {
		t.Fatalf("unexpected: %s", em)
	}
}
