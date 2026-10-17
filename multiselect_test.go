package gotui

import (
	"fmt"
	"strings"
	"testing"
)

func TestMultiSelectToggleAndConfirm(t *testing.T) {
	ms := NewMultiSelect().WithOptions("A", "B", "C")

	// Check A, B and C, then uncheck C again.
	for _, k := range []func() KeyMsg{
		func() KeyMsg { return keyRune(' ') },
		keyDown,
		func() KeyMsg { return keyRune(' ') },
		keyDown,
		func() KeyMsg { return keyRune(' ') },
		func() KeyMsg { return keyRune(' ') },
	} {
		m, _ := ms.Update(k())
		ms = m.(*MultiSelect)
	}

	m, cmd := ms.Update(keyEnter())
	ms = m.(*MultiSelect)
	if cmd == nil {
		t.Fatal("expected Quit command on Enter")
	}
	if _, ok := cmd().(QuitMsg); !ok {
		t.Fatal("expected QuitMsg")
	}
	got := ms.Result().([]string)
	if strings.Join(got, ",") != "A,B" {
		t.Fatalf("result = %v, want [A B]", got)
	}
}

func TestMultiSelectEmptyEnterIsNoop(t *testing.T) {
	ms := NewMultiSelect()
	m, cmd := ms.Update(keyEnter())
	if cmd != nil {
		t.Fatal("Enter on an empty checklist should not quit")
	}
	_ = m
}

func TestMultiSelectPreselect(t *testing.T) {
	ms := NewMultiSelect().WithOptions("A", "B", "C").WithChecked(0, 2)
	if !strings.Contains(stripANSI(ms.View()), "[x] A") {
		t.Fatal("A should render as checked")
	}

	m, cmd := ms.Update(keyEnter())
	ms = m.(*MultiSelect)
	if cmd == nil {
		t.Fatal("expected Quit command")
	}
	got := ms.Result().([]string)
	if strings.Join(got, ",") != "A,C" {
		t.Fatalf("result = %v, want [A C]", got)
	}
}

func TestMultiSelectSections(t *testing.T) {
	ms := NewMultiSelect().
		Section("Terminals", "kitty", "alacritty").
		Section("Shell", "zsh", "fish")

	view := stripANSI(ms.View())
	for _, want := range []string{"Terminals", "Shell", "kitty", "fish"} {
		if !strings.Contains(view, want) {
			t.Fatalf("view missing %q:\n%s", want, view)
		}
	}

	// Cursor starts on kitty; two downs reach zsh (headers are skipped).
	m, _ := ms.Update(keyDown())
	ms = m.(*MultiSelect)
	m, _ = ms.Update(keyDown())
	ms = m.(*MultiSelect)
	if ms.cursor != 2 {
		t.Fatalf("cursor = %d, want 2 (zsh)", ms.cursor)
	}

	m, _ = ms.Update(keyRune(' '))
	ms = m.(*MultiSelect)
	m, _ = ms.Update(keyEnter())
	ms = m.(*MultiSelect)
	if got := ms.Result().([]string); strings.Join(got, ",") != "zsh" {
		t.Fatalf("result = %v, want [zsh]", got)
	}
}

func TestMultiSelectScrolling(t *testing.T) {
	var opts []string
	for i := 0; i < 20; i++ {
		opts = append(opts, fmt.Sprintf("opt%02d", i))
	}
	ms := NewMultiSelect().WithOptions(opts...).WithRows(5)

	view := stripANSI(ms.View())
	if !strings.Contains(view, "more below") {
		t.Fatal("expected 'more below' at the top of a long list")
	}
	if strings.Contains(view, "more above") {
		t.Fatal("did not expect 'more above' at the top of the list")
	}
	if n := strings.Count(view, "["); n != 5 {
		t.Fatalf("rendered %d option checkboxes, want 5 (viewport)", n)
	}

	for i := 0; i < len(opts)-1; i++ {
		m, _ := ms.Update(keyDown())
		ms = m.(*MultiSelect)
	}
	view = stripANSI(ms.View())
	if !strings.Contains(view, "more above") {
		t.Fatal("expected 'more above' once scrolled to the bottom")
	}
	if strings.Contains(view, "more below") {
		t.Fatal("did not expect 'more below' at the bottom of the list")
	}
	if !strings.Contains(view, "opt19") {
		t.Fatal("cursor option opt19 should be visible")
	}
}

func TestMultiSelectViewStraightDown(t *testing.T) {
	ms := NewMultiSelect().WithTitle("Pick extras").Section("Group", "A", "B").WithOptions("C")
	indented := 0
	for i, line := range strings.Split(stripANSI(ms.View()), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lc := leadingCols(line)
		if lc == 0 {
			continue // title and footer are left-aligned
		}
		if lc != 4 {
			t.Fatalf("line %d (%q) starts at column %d, want 4", i, line, lc)
		}
		indented++
	}
	if indented < 4 {
		t.Fatalf("expected at least 4 indented lines, got %d", indented)
	}
}

func TestAppMultiSelect(t *testing.T) {
	app := New()
	app.MultiSelect("extras", "Choose extras", "A", "B")
	if len(app.steps) != 1 || app.steps[0].key != "extras" {
		t.Fatalf("step not registered correctly: %+v", app.steps)
	}
	ms, ok := app.steps[0].model.(*MultiSelect)
	if !ok {
		t.Fatal("step model is not *MultiSelect")
	}

	// Simulate the run: check A, confirm, and store the result like
	// App.Run does.
	m, _ := ms.Update(keyRune(' '))
	ms = m.(*MultiSelect)
	m, cmd := ms.Update(keyEnter())
	ms = m.(*MultiSelect)
	if _, ok := cmd().(QuitMsg); !ok {
		t.Fatal("expected Quit on Enter")
	}
	app.values["extras"] = ms.Result()

	got, ok := app.Value("extras").([]string)
	if !ok {
		t.Fatalf("app.Value(extras) is %T, want []string", app.Value("extras"))
	}
	if strings.Join(got, ",") != "A" {
		t.Fatalf("extras = %v, want [A]", got)
	}
}
