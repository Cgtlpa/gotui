package gotui

import (
	"strings"
	"testing"
)

func TestMultiSelectToggleAndConfirm(t *testing.T) {
	ms := NewMultiSelect().WithOptions("A", "B", "C")

	// Check A, move down twice to C, check C, toggle it off and back on.
	m, _ := ms.Update(keyRune(' '))
	ms = m.(*MultiSelect)
	m, _ = ms.Update(keyDown())
	ms = m.(*MultiSelect)
	m, _ = ms.Update(keyDown())
	ms = m.(*MultiSelect)
	for i := 0; i < 3; i++ {
		m, cmd := ms.Update(keyRune(' '))
		ms = m.(*MultiSelect)
		if cmd != nil {
			t.Fatal("space should not quit")
		}
	}

	if view := stripANSI(ms.View()); !strings.Contains(view, "[x] A") || !strings.Contains(view, "[x] C") {
		t.Fatalf("A and C should render as checked:\n%s", view)
	}

	m, cmd := ms.Update(keyEnter())
	ms = m.(*MultiSelect)
	if cmd == nil {
		t.Fatal("expected Quit command on Enter")
	}
	if _, ok := cmd().(QuitMsg); !ok {
		t.Fatal("expected QuitMsg")
	}
	if got := ms.Result().([]string); strings.Join(got, ",") != "A,C" {
		t.Fatalf("result = %v, want [A C]", got)
	}
}

func TestMultiSelectEnterWithNothingChecked(t *testing.T) {
	ms := NewMultiSelect().WithOptions("A", "B")
	m, cmd := ms.Update(keyEnter())
	ms = m.(*MultiSelect)
	if _, ok := cmd().(QuitMsg); !ok {
		t.Fatal("expected Quit on Enter")
	}
	if got := ms.Result().([]string); len(got) != 0 {
		t.Fatalf("result = %v, want empty", got)
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

func TestMultiSelectViewStraightDown(t *testing.T) {
	ms := NewMultiSelect().WithTitle("Pick extras").WithOptions("A", "B", "C")
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
	if indented != 3 {
		t.Fatalf("expected 3 indented option lines, got %d", indented)
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
