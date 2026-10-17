package gotui

import (
	"fmt"
	"strings"
)

// MultiSelect is a checkbox list: ↑/↓ (or j/k) moves the cursor, space
// toggles the option under it, and Enter confirms.
//
//	chosen, err := gotui.NewMultiSelect().
//	    WithTitle("Choose extras").
//	    WithOptions("Kitty", "Rofi", "Firefox").
//	    Run()
type MultiSelect struct {
	Title   string
	Options []string
	Theme   Theme
	Layout  Layout

	checked []bool
	cursor  int
	result  []string
}

// NewMultiSelect creates an empty checklist.
func NewMultiSelect() *MultiSelect {
	return &MultiSelect{Theme: DefaultTheme()}
}

func (m *MultiSelect) WithTitle(t string) *MultiSelect { m.Title = t; return m }

// WithOptions sets the selectable options. They all start unchecked.
func (m *MultiSelect) WithOptions(opts ...string) *MultiSelect {
	m.Options = opts
	m.checked = make([]bool, len(opts))
	return m
}

func (m *MultiSelect) WithTheme(t Theme) *MultiSelect   { m.Theme = t; return m }
func (m *MultiSelect) WithLayout(l Layout) *MultiSelect { m.Layout = l; return m }

func (m *MultiSelect) theme() Theme {
	if m.Theme == (Theme{}) {
		return DefaultTheme()
	}
	return m.Theme
}

// normalize keeps the checkbox state in step with Options, so that direct
// field assignment still renders and navigates correctly.
func (m *MultiSelect) normalize() {
	if len(m.checked) < len(m.Options) {
		m.checked = append(m.checked, make([]bool, len(m.Options)-len(m.checked))...)
	}
	switch {
	case len(m.Options) == 0:
		m.cursor = 0
	case m.cursor < 0:
		m.cursor = 0
	case m.cursor >= len(m.Options):
		m.cursor = len(m.Options) - 1
	}
}

func (m *MultiSelect) Init() Cmd { return nil }

func (m *MultiSelect) Update(msg Msg) (Model, Cmd) {
	km, ok := msg.(KeyMsg)
	if !ok {
		return m, nil
	}
	m.normalize()

	switch km.Key.Type {
	case KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case KeyDown:
		if m.cursor < len(m.Options)-1 {
			m.cursor++
		}
	case KeyRune:
		switch km.Key.Rune {
		case 'k':
			if m.cursor > 0 {
				m.cursor--
			}
		case 'j':
			if m.cursor < len(m.Options)-1 {
				m.cursor++
			}
		case ' ':
			if len(m.Options) > 0 {
				m.checked[m.cursor] = !m.checked[m.cursor]
			}
		}
	case KeyEnter:
		if len(m.Options) == 0 {
			return m, nil
		}
		m.result = make([]string, 0, len(m.Options))
		for i, o := range m.Options {
			if m.checked[i] {
				m.result = append(m.result, o)
			}
		}
		return m, Quit
	}
	return m, nil
}

// Result returns the checked options once the user pressed Enter.
func (m *MultiSelect) Result() interface{} { return m.result }

func (m *MultiSelect) count() int {
	n := 0
	for _, c := range m.checked {
		if c {
			n++
		}
	}
	return n
}

func (m *MultiSelect) View() string {
	m.normalize()
	th := m.theme()
	var b strings.Builder

	if m.Title != "" {
		b.WriteString(th.Bold())
		b.WriteString(th.FgColor(th.Primary))
		b.WriteString(m.Title)
		b.WriteString(th.Reset())
		b.WriteString("\n\n")
	}

	if len(m.Options) == 0 {
		b.WriteString(th.FgColor(th.Muted))
		b.WriteString("(no options)")
		b.WriteString(th.Reset())
		b.WriteString("\n\n")
	} else {
		for i, opt := range m.Options {
			mark := "[ ]"
			if m.checked[i] {
				mark = "[x]"
			}
			switch {
			case i == m.cursor:
				b.WriteString(th.Bold())
				b.WriteString(th.Reverse())
				b.WriteString(th.FgColor(th.Background))
				b.WriteString(th.BgColor(th.Primary))
				b.WriteString("  ▶ " + mark + " " + opt)
				b.WriteString(th.Reset())
			case m.checked[i]:
				b.WriteString(th.FgColor(th.Success))
				b.WriteString("    " + mark + " " + opt)
				b.WriteString(th.Reset())
			default:
				b.WriteString(th.FgColor(th.Foreground))
				b.WriteString("    " + mark + " " + opt)
				b.WriteString(th.Reset())
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(th.FgColor(th.Muted))
	b.WriteString(fmt.Sprintf("%d selected  •  ↑/↓ or j/k to move  •  space to toggle  •  Enter to confirm", m.count()))
	b.WriteString(th.Reset())

	result := b.String()
	if m.Layout != (Layout{}) {
		result = m.Layout.Apply(result)
	}
	return result
}

// Run runs the checklist on its own and returns the checked options.
func (m *MultiSelect) Run() ([]string, error) {
	v, err := Run(m)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	return v.([]string), nil
}
