package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// textInput is a cursor-aware single-line input field.
type textInput struct {
	buf string
	pos int // byte position of cursor
}

func (t *textInput) set(s string) { t.buf = s; t.pos = len(s) }
func (t *textInput) clear()       { t.buf = ""; t.pos = 0 }
func (t *textInput) val() string  { return t.buf }

// insert inserts s at current cursor position.
func (t *textInput) insert(s string) {
	t.buf = t.buf[:t.pos] + s + t.buf[t.pos:]
	t.pos += len(s)
}

// backspace deletes the rune before the cursor.
func (t *textInput) backspace() {
	if t.pos == 0 {
		return
	}
	_, sz := lastRune(t.buf[:t.pos])
	t.buf = t.buf[:t.pos-sz] + t.buf[t.pos:]
	t.pos -= sz
}

// deleteForward deletes the rune after the cursor.
func (t *textInput) deleteForward() {
	if t.pos >= len(t.buf) {
		return
	}
	_, sz := firstRune(t.buf[t.pos:])
	t.buf = t.buf[:t.pos] + t.buf[t.pos+sz:]
}

func (t *textInput) moveLeft()  { _, sz := lastRune(t.buf[:t.pos]); t.pos -= sz }
func (t *textInput) moveRight() { _, sz := firstRune(t.buf[t.pos:]); t.pos += sz }
func (t *textInput) moveHome()  { t.pos = 0 }
func (t *textInput) moveEnd()   { t.pos = len(t.buf) }

func (t *textInput) wordLeft() {
	p := t.pos
	for p > 0 {
		_, sz := lastRune(t.buf[:p])
		if t.buf[p-sz] == ' ' && p != t.pos {
			break
		}
		p -= sz
		if t.buf[p] == ' ' {
			continue
		}
	}
	t.pos = p
}

func (t *textInput) wordRight() {
	p := t.pos
	n := len(t.buf)
	for p < n && t.buf[p] != ' ' {
		_, sz := firstRune(t.buf[p:])
		p += sz
	}
	for p < n && t.buf[p] == ' ' {
		_, sz := firstRune(t.buf[p:])
		p += sz
	}
	t.pos = p
}

// view renders the field content with a block cursor injected at pos.
func (t *textInput) view() string {
	before := t.buf[:t.pos]
	after := t.buf[t.pos:]
	var cursorChar string
	if after == "" {
		cursorChar = "█"
		after = ""
	} else {
		_, sz := firstRune(after)
		cursorChar = "\x1b[7m" + after[:sz] + "\x1b[0m"
		after = after[sz:]
	}
	return before + cursorChar + after
}

// viewWidth renders the input with text wrapping at the given width,
// showing the content across multiple lines if needed.
func (t *textInput) viewWidth(w int) string {
	if w <= 0 {
		return t.view()
	}

	// If the buffer fits in one line, just use the simple view
	if len(t.buf) < w {
		return t.view()
	}

	// Wrap the buffer into lines of at most w characters,
	// then inject the cursor at the correct position.
	var lines []string
	buf := t.buf
	for len(buf) > w {
		lines = append(lines, buf[:w])
		buf = buf[w:]
	}
	lines = append(lines, buf)

	// Find which line and offset the cursor is on
	cursorLine := 0
	cursorCol := t.pos
	for cursorCol > w {
		cursorCol -= w
		cursorLine++
	}
	if cursorCol == w && cursorLine < len(lines)-1 {
		cursorCol = 0
		cursorLine++
	}

	// Rebuild lines with cursor injected
	var result strings.Builder
	for i, l := range lines {
		if i > 0 {
			result.WriteString("\n")
		}
		if i == cursorLine {
			before := l[:cursorCol]
			after := l[cursorCol:]
			result.WriteString(before)
			if after == "" {
				result.WriteString("█")
			} else {
				_, sz := firstRune(after)
				result.WriteString("\x1b[7m" + after[:sz] + "\x1b[0m")
				result.WriteString(after[sz:])
			}
		} else {
			result.WriteString(l)
		}
	}

	// If cursor is at the very end and on a new line
	if t.pos == len(t.buf) && t.pos > 0 && t.pos%w == 0 {
		result.WriteString("\n█")
	}

	return result.String()
}

// handle processes a KeyMsg for text input; returns true if the key was consumed.
func (t *textInput) handle(msg tea.KeyMsg) bool {
	switch msg.String() {
	case "left":
		if t.pos > 0 {
			t.moveLeft()
		}
	case "right":
		if t.pos < len(t.buf) {
			t.moveRight()
		}
	case "ctrl+left", "alt+left":
		t.wordLeft()
	case "ctrl+right", "alt+right":
		t.wordRight()
	case "home", "ctrl+a":
		t.moveHome()
	case "end", "ctrl+e":
		t.moveEnd()
	case "backspace":
		t.backspace()
	case "ctrl+backspace", "alt+backspace":
		prev := t.pos
		t.wordLeft()
		t.buf = t.buf[:t.pos] + t.buf[prev:]
	case "ctrl+d", "delete":
		t.deleteForward()
	default:
		ch := msg.String()
		if ch == " " || (len(ch) == 1 && ch[0] >= 0x20) {
			t.insert(ch)
		} else if len(msg.Runes) > 1 && msg.Type == tea.KeyRunes {
			t.insert(string(msg.Runes))
		} else {
			return false
		}
	}
	return true
}

func lastRune(s string) (rune, int) {
	if s == "" {
		return 0, 0
	}
	r, sz := rune(s[len(s)-1]), 1
	_ = r
	for sz < len(s) && s[len(s)-sz]&0xC0 == 0x80 {
		sz++
	}
	return rune(s[len(s)-sz]), sz
}

func firstRune(s string) (rune, int) {
	if s == "" {
		return 0, 0
	}
	for i, r := range s {
		_ = i
		return r, len(string(r))
	}
	return 0, 0
}
