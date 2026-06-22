package main

import tea "github.com/charmbracelet/bubbletea"

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
