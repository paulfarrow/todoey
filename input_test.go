package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTextInput_SetAndVal(t *testing.T) {
	var ti textInput
	ti.set("hello")
	if ti.val() != "hello" {
		t.Fatalf("expected 'hello', got %q", ti.val())
	}
	if ti.pos != 5 {
		t.Fatalf("expected pos=5, got %d", ti.pos)
	}
}

func TestTextInput_Clear(t *testing.T) {
	var ti textInput
	ti.set("hello")
	ti.clear()
	if ti.val() != "" || ti.pos != 0 {
		t.Fatalf("expected empty after clear, got %q pos=%d", ti.val(), ti.pos)
	}
}

func TestTextInput_Insert(t *testing.T) {
	var ti textInput
	ti.set("helo")
	ti.pos = 3
	ti.insert("l")
	if ti.val() != "hello" {
		t.Fatalf("expected 'hello', got %q", ti.val())
	}
	if ti.pos != 4 {
		t.Fatalf("expected pos=4, got %d", ti.pos)
	}
}

func TestTextInput_InsertAtStart(t *testing.T) {
	var ti textInput
	ti.set("world")
	ti.pos = 0
	ti.insert("hello ")
	if ti.val() != "hello world" {
		t.Fatalf("expected 'hello world', got %q", ti.val())
	}
}

func TestTextInput_Backspace(t *testing.T) {
	var ti textInput
	ti.set("hello")
	ti.backspace()
	if ti.val() != "hell" || ti.pos != 4 {
		t.Fatalf("expected 'hell' pos=4, got %q pos=%d", ti.val(), ti.pos)
	}
}

func TestTextInput_BackspaceAtStart(t *testing.T) {
	var ti textInput
	ti.set("hello")
	ti.pos = 0
	ti.backspace()
	if ti.val() != "hello" || ti.pos != 0 {
		t.Fatalf("backspace at start should be no-op, got %q pos=%d", ti.val(), ti.pos)
	}
}

func TestTextInput_BackspaceMiddle(t *testing.T) {
	var ti textInput
	ti.set("hello")
	ti.pos = 3
	ti.backspace()
	if ti.val() != "helo" || ti.pos != 2 {
		t.Fatalf("expected 'helo' pos=2, got %q pos=%d", ti.val(), ti.pos)
	}
}

func TestTextInput_DeleteForward(t *testing.T) {
	var ti textInput
	ti.set("hello")
	ti.pos = 0
	ti.deleteForward()
	if ti.val() != "ello" || ti.pos != 0 {
		t.Fatalf("expected 'ello' pos=0, got %q pos=%d", ti.val(), ti.pos)
	}
}

func TestTextInput_DeleteForwardAtEnd(t *testing.T) {
	var ti textInput
	ti.set("hello")
	ti.deleteForward()
	if ti.val() != "hello" {
		t.Fatalf("deleteForward at end should be no-op, got %q", ti.val())
	}
}

func TestTextInput_MoveLeftRight(t *testing.T) {
	var ti textInput
	ti.set("hello")
	ti.moveLeft()
	if ti.pos != 4 {
		t.Fatalf("expected pos=4 after moveLeft, got %d", ti.pos)
	}
	ti.moveRight()
	if ti.pos != 5 {
		t.Fatalf("expected pos=5 after moveRight, got %d", ti.pos)
	}
}

func TestTextInput_MoveHomeEnd(t *testing.T) {
	var ti textInput
	ti.set("hello world")
	ti.moveHome()
	if ti.pos != 0 {
		t.Fatalf("expected pos=0 after moveHome, got %d", ti.pos)
	}
	ti.moveEnd()
	if ti.pos != 11 {
		t.Fatalf("expected pos=11 after moveEnd, got %d", ti.pos)
	}
}

func TestTextInput_WordLeft(t *testing.T) {
	var ti textInput
	ti.set("hello world foo")
	// cursor at end
	ti.wordLeft()
	if ti.pos != 12 {
		t.Fatalf("expected pos=12, got %d", ti.pos)
	}
	ti.wordLeft()
	if ti.pos != 6 {
		t.Fatalf("expected pos=6, got %d", ti.pos)
	}
	ti.wordLeft()
	if ti.pos != 0 {
		t.Fatalf("expected pos=0, got %d", ti.pos)
	}
}

func TestTextInput_WordRight(t *testing.T) {
	var ti textInput
	ti.set("hello world foo")
	ti.pos = 0
	ti.wordRight()
	if ti.pos != 6 {
		t.Fatalf("expected pos=6, got %d", ti.pos)
	}
	ti.wordRight()
	if ti.pos != 12 {
		t.Fatalf("expected pos=12, got %d", ti.pos)
	}
	ti.wordRight()
	if ti.pos != 15 {
		t.Fatalf("expected pos=15, got %d", ti.pos)
	}
}

func TestTextInput_View(t *testing.T) {
	var ti textInput
	ti.set("ab")
	ti.pos = 1
	v := ti.view()
	// cursor should be on 'b' with reverse video
	if v == "" {
		t.Fatal("view should not be empty")
	}
	// 'a' should appear before cursor
	if v[0] != 'a' {
		t.Fatalf("expected 'a' at start, got %q", v)
	}
}

func TestTextInput_ViewAtEnd(t *testing.T) {
	var ti textInput
	ti.set("ab")
	v := ti.view()
	// Should contain block cursor at end
	if len(v) == 0 {
		t.Fatal("view should not be empty")
	}
	if !contains(v, "█") {
		t.Fatalf("expected block cursor at end, got %q", v)
	}
}

func TestTextInput_HandleLeft(t *testing.T) {
	var ti textInput
	ti.set("abc")
	consumed := ti.handle(tea.KeyMsg{Type: tea.KeyLeft})
	if !consumed {
		t.Fatal("left should be consumed")
	}
	if ti.pos != 2 {
		t.Fatalf("expected pos=2, got %d", ti.pos)
	}
}

func TestTextInput_HandleRight(t *testing.T) {
	var ti textInput
	ti.set("abc")
	ti.pos = 1
	consumed := ti.handle(tea.KeyMsg{Type: tea.KeyRight})
	if !consumed {
		t.Fatal("right should be consumed")
	}
	if ti.pos != 2 {
		t.Fatalf("expected pos=2, got %d", ti.pos)
	}
}

func TestTextInput_HandleHome(t *testing.T) {
	var ti textInput
	ti.set("abc")
	ti.handle(tea.KeyMsg{Type: tea.KeyHome})
	if ti.pos != 0 {
		t.Fatalf("expected pos=0, got %d", ti.pos)
	}
}

func TestTextInput_HandleEnd(t *testing.T) {
	var ti textInput
	ti.set("abc")
	ti.pos = 0
	ti.handle(tea.KeyMsg{Type: tea.KeyEnd})
	if ti.pos != 3 {
		t.Fatalf("expected pos=3, got %d", ti.pos)
	}
}

func TestTextInput_HandleBackspace(t *testing.T) {
	var ti textInput
	ti.set("abc")
	ti.handle(tea.KeyMsg{Type: tea.KeyBackspace})
	if ti.val() != "ab" {
		t.Fatalf("expected 'ab', got %q", ti.val())
	}
}

func TestTextInput_HandleDelete(t *testing.T) {
	var ti textInput
	ti.set("abc")
	ti.pos = 0
	ti.handle(tea.KeyMsg{Type: tea.KeyDelete})
	if ti.val() != "bc" {
		t.Fatalf("expected 'bc', got %q", ti.val())
	}
}

func TestTextInput_HandleCharInput(t *testing.T) {
	var ti textInput
	ti.set("")
	ti.handle(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if ti.val() != "x" {
		t.Fatalf("expected 'x', got %q", ti.val())
	}
}

func TestTextInput_HandleSpace(t *testing.T) {
	var ti textInput
	ti.set("ab")
	ti.handle(tea.KeyMsg{Type: tea.KeySpace})
	if ti.val() != "ab " {
		t.Fatalf("expected 'ab ', got %q", ti.val())
	}
}

func TestTextInput_HandleUnknownKeyNotConsumed(t *testing.T) {
	var ti textInput
	ti.set("abc")
	consumed := ti.handle(tea.KeyMsg{Type: tea.KeyF1})
	if consumed {
		t.Fatal("F1 should not be consumed")
	}
}

func TestLastRune(t *testing.T) {
	_, sz := lastRune("hello")
	if sz != 1 {
		t.Fatalf("expected size 1, got %d", sz)
	}
	_, sz = lastRune("")
	if sz != 0 {
		t.Fatalf("expected size 0 for empty, got %d", sz)
	}
}

func TestFirstRune(t *testing.T) {
	_, sz := firstRune("hello")
	if sz != 1 {
		t.Fatalf("expected size 1, got %d", sz)
	}
	_, sz = firstRune("")
	if sz != 0 {
		t.Fatalf("expected size 0 for empty, got %d", sz)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
