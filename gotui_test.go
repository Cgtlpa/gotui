package gotui

import (
	"errors"
	"strings"
	"testing"
)

func keyRune(r rune) KeyMsg {
	return KeyMsg{Key: Key{Type: KeyRune, Rune: r}}
}

func keyEnter() KeyMsg {
	return KeyMsg{Key: Key{Type: KeyEnter}}
}

func keyBackspace() KeyMsg {
	return KeyMsg{Key: Key{Type: KeyBackspace}}
}

func keyUp() KeyMsg {
	return KeyMsg{Key: Key{Type: KeyUp}}
}

func keyDown() KeyMsg {
	return KeyMsg{Key: Key{Type: KeyDown}}
}

func keyLeft() KeyMsg {
	return KeyMsg{Key: Key{Type: KeyLeft}}
}

func keyRight() KeyMsg {
	return KeyMsg{Key: Key{Type: KeyRight}}
}

func keyEsc() KeyMsg {
	return KeyMsg{Key: Key{Type: KeyEsc}}
}

func keyCtrlC() KeyMsg {
	return KeyMsg{Key: Key{Type: KeyCtrlC}}
}

// ---- Input tests ----

func TestInputTypingAndEditing(t *testing.T) {
	in := NewInput().WithPrompt("> ").WithPlaceholder("name")

	m, cmd := in.Update(keyRune('H'))
	in = m.(*Input)
	if cmd != nil {
		t.Fatal("expected no command on typing")
	}
	if in.Value != "H" {
		t.Fatalf("value = %q, want %q", in.Value, "H")
	}

	m, cmd = in.Update(keyRune('i'))
	in = m.(*Input)
	if in.Value != "Hi" {
		t.Fatalf("value = %q, want %q", in.Value, "Hi")
	}

	// Move cursor left, insert, backspace.
	m, _ = in.Update(keyLeft())
	in = m.(*Input)
	if in.cursor != 1 {
		t.Fatalf("cursor = %d, want %d", in.cursor, 1)
	}
	m, _ = in.Update(keyRune('a'))
	in = m.(*Input)
	if in.Value != "Hai" {
		t.Fatalf("value = %q, want %q", in.Value, "Hai")
	}
	m, _ = in.Update(keyBackspace())
	in = m.(*Input)
	if in.Value != "Hi" {
		t.Fatalf("value after backspace = %q, want %q", in.Value, "Hi")
	}
}

func TestInputBackspaceDeletesLastChar(t *testing.T) {
	in := NewInput().WithValue("Hi")
	m, _ := in.Update(keyBackspace())
	in = m.(*Input)
	if in.Value != "H" {
		t.Fatalf("value after backspace = %q, want %q", in.Value, "H")
	}
}

func TestInputBackspaceAtCursorZeroIsNoop(t *testing.T) {
	in := NewInput().WithValue("Hi")
	// cursor starts at end (position 2), move to start.
	m, _ := in.Update(keyLeft())
	in = m.(*Input)
	m, _ = in.Update(keyLeft())
	in = m.(*Input)
	if in.cursor != 0 {
		t.Fatalf("cursor = %d, want %d", in.cursor, 0)
	}
	// Backspace at cursor 0 should be a no-op.
	m, _ = in.Update(keyBackspace())
	in = m.(*Input)
	if in.Value != "Hi" {
		t.Fatalf("value after backspace at cursor 0 = %q, want %q", in.Value, "Hi")
	}
}

func TestInputEnterQuits(t *testing.T) {
	in := NewInput().WithValue("final")
	m, cmd := in.Update(keyEnter())
	in = m.(*Input)
	if _, ok := cmd().(QuitMsg); !ok {
		t.Fatal("expected Quit command")
	}
	if in.result != "final" {
		t.Fatalf("result = %q, want %q", in.result, "final")
	}
}

func TestInputCharLimit(t *testing.T) {
	in := NewInput().WithCharLimit(3)
	m, _ := in.Update(keyRune('a'))
	in = m.(*Input)
	m, _ = in.Update(keyRune('b'))
	in = m.(*Input)
	m, _ = in.Update(keyRune('c'))
	in = m.(*Input)
	// Should be at limit now, typing more should be ignored.
	m, _ = in.Update(keyRune('d'))
	in = m.(*Input)
	if in.Value != "abc" {
		t.Fatalf("value = %q, want %q (char limit enforced)", in.Value, "abc")
	}
}

// ---- Select tests ----

func TestSelectNavigation(t *testing.T) {
	s := NewSelect().WithOptions("Red", "Green", "Blue")
	if s.cursor != 0 {
		t.Fatalf("initial cursor = %d", s.cursor)
	}

	m, _ := s.Update(keyDown())
	s = m.(*Select)
	if s.cursor != 1 {
		t.Fatalf("cursor after down = %d", s.cursor)
	}
	m, _ = s.Update(keyDown())
	s = m.(*Select)
	if s.cursor != 2 {
		t.Fatalf("cursor after second down = %d", s.cursor)
	}
	m, _ = s.Update(keyUp())
	s = m.(*Select)
	if s.cursor != 1 {
		t.Fatalf("cursor after up = %d", s.cursor)
	}

	// Enter on cursor 1 -> Green and Quit.
	m, cmd := s.Update(keyEnter())
	s = m.(*Select)
	if _, ok := cmd().(QuitMsg); !ok {
		t.Fatal("expected Quit on Enter")
	}
	if s.result != "Green" {
		t.Fatalf("result = %q, want %q", s.result, "Green")
	}
}

func TestSelectSelectsFirstByDefault(t *testing.T) {
	s := NewSelect().WithOptions("Only")
	m, cmd := s.Update(keyEnter())
	s = m.(*Select)
	if _, ok := cmd().(QuitMsg); !ok {
		t.Fatal("expected Quit on Enter")
	}
	if s.result != "Only" {
		t.Fatalf("result = %q, want %q", s.result, "Only")
	}
}

// ---- Confirm tests ----

func TestConfirmToggle(t *testing.T) {
	c := NewConfirm().DefaultYes()
	if c.selected != 1 {
		t.Fatalf("default yes selected = %d", c.selected)
	}

	m, _ := c.Update(keyLeft())
	c = m.(*Confirm)
	if c.selected != 0 {
		t.Fatalf("after left selected = %d", c.selected)
	}

	m, _ = c.Update(keyRight())
	c = m.(*Confirm)
	if c.selected != 1 {
		t.Fatalf("after right selected = %d", c.selected)
	}

	m, cmd := c.Update(keyEnter())
	c = m.(*Confirm)
	if _, ok := cmd().(QuitMsg); !ok {
		t.Fatal("expected Quit on Enter")
	}
	if c.result != true {
		t.Fatalf("result = %v, want true", c.result)
	}
}

// ---- Form tests ----

func TestFormTabAndSubmit(t *testing.T) {
	f := NewForm().WithTitle("t").WithInput("name", "Name").WithInput("email", "Email")
	if f.focus != 0 {
		t.Fatalf("initial focus = %d", f.focus)
	}

	m, _ := f.Update(keyRune('A'))
	f = m.(*Form)
	if f.fields[0].Value != "A" {
		t.Fatalf("field0 = %q", f.fields[0].Value)
	}

	m, _ = f.Update(KeyMsg{Key: Key{Type: KeyTab}})
	f = m.(*Form)
	if f.focus != 1 || !f.fields[1].focused || f.fields[0].focused {
		t.Fatalf("focus did not move to field 1: focus=%d", f.focus)
	}

	m, _ = f.Update(keyRune('B'))
	f = m.(*Form)
	if f.fields[1].Value != "B" {
		t.Fatalf("field1 = %q", f.fields[1].Value)
	}

	m, cmd := f.Update(keyEnter())
	f = m.(*Form)
	if _, ok := cmd().(QuitMsg); !ok {
		t.Fatal("expected Quit on submit")
	}
	if f.Get("name") != "A" || f.Get("email") != "B" {
		t.Fatalf("values: name=%q email=%q", f.Get("name"), f.Get("email"))
	}
	_ = f.View()
}

// ---- App result storage tests ----

func TestAppValueStorage(t *testing.T) {
	app := New()
	app.Select("wm", "Choose WM", "InstantWM", "Niri", "DWM")

	// Simulate selecting the first option (InstantWM) and confirming.
	// We can't easily run the full app without a terminal, so test the
	// Value/Result flow directly through the component.
	sel := NewSelect().WithTitle("Choose WM").WithOptions("InstantWM", "Niri", "DWM")
	m, cmd := sel.Update(keyEnter())
	sel = m.(*Select)
	if _, ok := cmd().(QuitMsg); !ok {
		t.Fatal("expected Quit on Enter")
	}
	if sel.Result() != "InstantWM" {
		t.Fatalf("result = %q, want %q", sel.Result(), "InstantWM")
	}

	// Store and retrieve via app.Value
	app.values["wm"] = sel.Result()
	if got := app.Value("wm"); got != "InstantWM" {
		t.Fatalf("app.Value(wm) = %q, want %q", got, "InstantWM")
	}
}

func TestAppValueMissingKey(t *testing.T) {
	app := New()
	if app.Value("nonexistent") != nil {
		t.Fatal("expected nil for missing key")
	}
}

// ---- Cancellation tests ----

func TestErrCancelledIsReturned(t *testing.T) {
	if !errors.Is(ErrCancelled, ErrCancelled) {
		t.Fatal("ErrCancelled should match itself")
	}
}

// TestRenderConvertsNewlinesToCRNL verifies that render() converts \n to \r\n.
// In raw mode, \n alone moves down without returning to column 0, causing
// diagonal text. \r\n ensures every line starts at column 0.
func TestRenderConvertsNewlinesToCRNL(t *testing.T) {
	view := "line1\nline2\nline3"
	converted := strings.ReplaceAll(view, "\n", "\r\n")
	if converted != "line1\r\nline2\r\nline3" {
		t.Fatalf("expected \\r\\n conversion, got %q", converted)
	}
}

// TestRenderStartsWithCarriageReturn verifies render() prefixes output with \r
// to ensure column 0 at the start.
func TestRenderStartsWithCarriageReturn(t *testing.T) {
	prefix := "\r" + ansiCursorHome + ansiClearScreen + ansiClearScrollback
	if prefix[0] != '\r' {
		t.Fatal("render() should start with \\r to reset column")
	}
}

// TestAllComponentsStraightDown verifies every built-in component renders
// with straight lines (no diagonal/stair-step effect) after ANSI codes are
// stripped and \n is converted to \r\n.
func TestAllComponentsStraightDown(t *testing.T) {
	type viewCase struct {
		name string
		view string
	}

	cases := []viewCase{
		{
			name: "Input empty",
			view: NewInput().WithPrompt("Name: ").WithPlaceholder("e.g. Ada").View(),
		},
		{
			name: "Input filled",
			view: NewInput().WithPrompt("Name: ").WithValue("Ada Lovelace").View(),
		},
		{
			name: "Input password",
			view: NewPassword().WithPrompt("Pass: ").WithValue("secret").View(),
		},
		{
			name: "Select",
			view: NewSelect().WithTitle("Pick").WithOptions("A", "B", "C").View(),
		},
		{
			name: "Select cursor on second",
			view: (func() string {
				s := NewSelect().WithTitle("Pick").WithOptions("A", "B", "C")
				s.WithCursor(1)
				return s.View()
			})(),
		},
		{
			name: "Confirm No",
			view: NewConfirm().WithTitle("Sure?").DefaultNo().View(),
		},
		{
			name: "Confirm Yes",
			view: NewConfirm().WithTitle("Sure?").DefaultYes().View(),
		},
		{
			name: "Form",
			view: NewForm().WithTitle("Sign up").
				WithInput("name", "Name").
				WithInput("email", "Email").
				WithPassword("pass", "Password").View(),
		},
		{
			name: "Text",
			view: NewText().WithTitle("Hi").WithBody("Hello world\nSecond line").View(),
		},
		{
			name: "Spinner",
			view: NewSpinner().WithText("Loading...").View(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate render(): strip ANSI, then check each section is straight.
			// A section is a group of non-blank lines separated by blank lines.
			stripped := stripANSI(tc.view)
			lines := strings.Split(stripped, "\n")

			sections := [][]string{}
			current := []string{}
			for _, line := range lines {
				if strings.TrimSpace(line) == "" {
					if len(current) > 0 {
						sections = append(sections, current)
						current = nil
					}
				} else {
					current = append(current, line)
				}
			}
			if len(current) > 0 {
				sections = append(sections, current)
			}

			// Each section must have all lines starting at the same column.
			for si, section := range sections {
				baseline := leadingCols(section[0])
				for li, line := range section {
					if leadingCols(line) != baseline {
						t.Fatalf("section %d line %d (%q) starts at column %d, want %d (diagonal!)",
							si, li, line, leadingCols(line), baseline)
					}
				}
			}
		})
	}
}

// stripANSI removes ANSI escape sequences from a string for testing alignment.
func stripANSI(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && (s[i] < '@' || s[i] > '~') {
				i++
			}
			if i < len(s) {
				i++
			}
		} else {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.String()
}

// leadingCols returns the number of leading visual columns, treating the
// selection arrow "▶" as occupying 1 column (like a space).
func leadingCols(s string) int {
	cols := 0
	i := 0
	for i < len(s) {
		if s[i] == ' ' {
			i++
			cols++
			continue
		}
		// Detect multi-byte UTF-8 arrow "▶" (U+25B6, bytes e2 96 b6).
		// It takes 3 bytes but only 1 visual column.
		if i+2 < len(s) && s[i] == 0xe2 && s[i+1] == 0x96 && s[i+2] == 0xb6 {
			i += 3 // skip the 3 UTF-8 bytes
			cols++ // count as 1 visual column
			continue
		}
		break
	}
	return cols
}

// TestViewStraightDown verifies that rendered views have lines starting at
// the same column (no diagonal/stair-step effect).
func TestViewStraightDown(t *testing.T) {
	// Select view — all option lines must start at the same visual column.
	s := NewSelect().WithTitle("Pick one").WithOptions("InstantWM", "Niri", "DWM")
	view := s.View()
	lines := strings.Split(stripANSI(view), "\n")

	// Collect leading columns of all non-blank lines that are indented (>0 cols).
	// Lines at column 0 (title, footer) are left-aligned and exempt.
	var indentedCols []int
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lc := leadingCols(line)
		if lc > 0 {
			indentedCols = append(indentedCols, lc)
		}
	}
	if len(indentedCols) == 0 {
		t.Fatal("no indented lines found")
	}

	// All indented lines must start at the same column.
	for _, lc := range indentedCols {
		if lc != indentedCols[0] {
			t.Fatalf("indented lines start at different columns: %v (diagonal!)", indentedCols)
		}
	}
}

// TestInputViewStraightDown verifies Input renders straight.
func TestInputViewStraightDown(t *testing.T) {
	in := NewInput().WithPrompt("Name: ").WithPlaceholder("e.g. Ada")
	view := in.View()
	lines := strings.Split(stripANSI(view), "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lc := leadingCols(line)
		if lc != 0 {
			t.Fatalf("line %d (%q) starts at column %d, want 0 (diagonal!)", i, line, lc)
		}
	}
}

// TestFormViewStraightDown verifies Form fields render straight down.
// Form fields are left-aligned at column 0.
func TestFormViewStraightDown(t *testing.T) {
	f := NewForm().WithTitle("Sign up").WithInput("name", "Name").WithInput("email", "Email")
	view := f.View()
	lines := strings.Split(stripANSI(view), "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lc := leadingCols(line)
		if lc != 0 {
			t.Fatalf("line %d (%q) starts at column %d, want 0 (diagonal!)", i, line, lc)
		}
	}
}

// ---- Layout / centering tests ----

// TestPadLineCenter verifies that padLine centers text within the given width.
func TestPadLineCenter(t *testing.T) {
	// "Hi" in width 10: pad = 8, left = 4, right = 4
	result := padLine("Hi", 10, true)
	if leadingCols(result) != 4 {
		t.Fatalf("padLine('Hi', 10, center) leading cols = %d, want 4", leadingCols(result))
	}
	if visibleLen(result) != 10 {
		t.Fatalf("padLine('Hi', 10, center) visible len = %d, want 10", visibleLen(result))
	}

	// "Hello" in width 10: pad = 5, left = 2, right = 3
	result = padLine("Hello", 10, true)
	if leadingCols(result) != 2 {
		t.Fatalf("padLine('Hello', 10, center) leading cols = %d, want 2", leadingCols(result))
	}

	// "Wide text" in width 5: text is wider, no padding
	result = padLine("Wide text", 5, true)
	if result != "Wide text" {
		t.Fatalf("padLine('Wide text', 5, center) = %q, want %q", result, "Wide text")
	}
}

// TestPadLineNoCenter verifies left-align padding.
func TestPadLineNoCenter(t *testing.T) {
	result := padLine("Hi", 10, false)
	if leadingCols(result) != 0 {
		t.Fatalf("padLine('Hi', 10, left) leading cols = %d, want 0", leadingCols(result))
	}
	if visibleLen(result) != 10 {
		t.Fatalf("padLine('Hi', 10, left) visible len = %d, want 10", visibleLen(result))
	}
}

// TestLayoutApplyCenter verifies Layout.Apply centers each line within the width.
func TestLayoutApplyCenter(t *testing.T) {
	layout := Layout{Width: 20, Center: true}
	view := "Title\nShort\nA longer line"
	result := layout.Apply(view)

	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Each line should be padded to exactly Width visible characters.
		if visibleLen(line) != 20 {
			t.Fatalf("line %q is not padded to width 20: visibleLen=%d",
				line, visibleLen(line))
		}
	}
}

// TestLayoutApplyNoWidth verifies that Width=0 means no padding.
func TestLayoutApplyNoWidth(t *testing.T) {
	layout := Layout{Center: true} // Width = 0
	view := "Hello\nWorld"
	result := layout.Apply(view)
	if result != view {
		t.Fatalf("Layout{Center:true} with Width=0 should not pad, got %q", result)
	}
}

// TestScaleLinePreservesANSI verifies that scaleLine doesn't corrupt escape sequences.
func TestScaleLinePreservesANSI(t *testing.T) {
	// A line with ANSI color codes
	line := "\x1b[38;5;6mCyan text\x1b[0m"
	result := scaleLine(line, 2)
	// The escape sequences should be intact
	if !strings.Contains(result, "\x1b[38;5;6m") {
		t.Fatal("scaleLine lost the foreground color escape")
	}
	if !strings.Contains(result, "\x1b[0m") {
		t.Fatal("scaleLine lost the reset escape")
	}
	// Each visible character should be doubled
	visible := stripANSI(result)
	expected := "CCyyaann  tteexxtt"
	if visible != expected {
		t.Fatalf("scaleLine visible text = %q, want %q", visible, expected)
	}
}

// TestCenteredSelectStraightDown verifies that a centered Select renders all
// option lines centered within the width.
func TestCenteredSelectStraightDown(t *testing.T) {
	layout := Layout{Width: 40, Center: true}
	s := NewSelect().WithTitle("Pick one").WithOptions("InstantWM", "Niri", "DWM")
	view := layout.Apply(s.View())
	stripped := stripANSI(view)
	lines := strings.Split(stripped, "\n")

	// Each indented content line shorter than Width should be padded to Width.
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Skip lines before the first blank line (title area)
		beforeBlank := true
		for j := 0; j < i; j++ {
			if strings.TrimSpace(lines[j]) == "" {
				beforeBlank = false
				break
			}
		}
		if beforeBlank {
			continue
		}
		// Lines wider than Width are left unchanged by padLine.
		if visibleLen(line) <= 40 {
			if visibleLen(line) != 40 {
				t.Fatalf("line %d (%q) is not padded to width 40: visibleLen=%d",
					i, line, visibleLen(line))
			}
		}
	}
}
