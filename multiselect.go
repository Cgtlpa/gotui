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
//	    Section("Terminals", "Kitty", "Alacritty").
//	    Section("Shell", "Zsh", "Fish").
//	    WithRows(10).
//	    Run()
type MultiSelect struct {
	Title   string
	Options []string
	Theme   Theme
	Layout  Layout

	rows     []msRow
	checked  []bool
	cursor   int
	offset   int
	viewport int
	result   []string
}

// msRow is one rendered line of the list: a section header or an option.
type msRow struct {
	header bool
	text   string
	opt    int // index into Options (options only)
}

// NewMultiSelect creates an empty checklist.
func NewMultiSelect() *MultiSelect {
	return &MultiSelect{Theme: DefaultTheme()}
}

// WithTitle sets the heading shown above the list.
func (m *MultiSelect) WithTitle(t string) *MultiSelect { m.Title = t; return m }

// WithOptions appends selectable options.
func (m *MultiSelect) WithOptions(opts ...string) *MultiSelect {
	for _, o := range opts {
		m.rows = append(m.rows, msRow{text: o, opt: len(m.Options)})
		m.Options = append(m.Options, o)
	}
	if n := len(m.Options) - len(m.checked); n > 0 {
		m.checked = append(m.checked, make([]bool, n)...)
	}
	return m
}

// Section appends a non-selectable heading followed by its options.
func (m *MultiSelect) Section(title string, opts ...string) *MultiSelect {
	m.rows = append(m.rows, msRow{header: true, text: title})
	return m.WithOptions(opts...)
}

// WithChecked pre-checks the options at the given indices.
func (m *MultiSelect) WithChecked(indices ...int) *MultiSelect {
	for _, i := range indices {
		if i < 0 {
			continue
		}
		if i >= len(m.checked) {
			m.checked = append(m.checked, make([]bool, i+1-len(m.checked))...)
		}
		m.checked[i] = true
	}
	return m
}

// WithCursor moves the cursor to the option at index n.
func (m *MultiSelect) WithCursor(n int) *MultiSelect { m.cursor = n; return m }

// WithTheme sets the color palette.
func (m *MultiSelect) WithTheme(t Theme) *MultiSelect { m.Theme = t; return m }

// WithLayout sets standalone layout options.
func (m *MultiSelect) WithLayout(l Layout) *MultiSelect { m.Layout = l; return m }

// WithRows limits how many list lines are visible at once (0 = all).
// Longer lists scroll to keep the cursor in view.
func (m *MultiSelect) WithRows(n int) *MultiSelect { m.viewport = n; return m }

func (m *MultiSelect) theme() Theme {
	if m.Theme == (Theme{}) {
		return DefaultTheme()
	}
	return m.Theme
}

// sync repairs internal state so that direct assignment to the public
// fields (Options in particular) still renders and navigates correctly.
func (m *MultiSelect) sync() {
	if len(m.rows) == 0 && len(m.Options) > 0 {
		for i, o := range m.Options {
			m.rows = append(m.rows, msRow{text: o, opt: i})
		}
	}
	if n := len(m.Options) - len(m.checked); n > 0 {
		m.checked = append(m.checked, make([]bool, n)...)
	} else if n < 0 {
		m.checked = m.checked[:len(m.Options)]
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

// Init implements Model.
func (m *MultiSelect) Init() Cmd { return nil }

// Update implements Model.
func (m *MultiSelect) Update(msg Msg) (Model, Cmd) {
	km, ok := msg.(KeyMsg)
	if !ok {
		return m, nil
	}
	m.sync()

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
			if i < len(m.checked) && m.checked[i] {
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

// window returns the range of rows to render, scrolling so the cursor stays
// visible when WithRows limits the height.
func (m *MultiSelect) window() (int, int) {
	if m.viewport <= 0 || m.viewport >= len(m.rows) {
		return 0, len(m.rows)
	}
	cur := 0
	for i, r := range m.rows {
		if !r.header && r.opt == m.cursor {
			cur = i
			break
		}
	}
	start := m.offset
	if cur < start {
		start = cur
	}
	if cur >= start+m.viewport {
		start = cur - m.viewport + 1
	}
	if max := len(m.rows) - m.viewport; start > max {
		start = max
	}
	if start < 0 {
		start = 0
	}
	m.offset = start
	end := start + m.viewport
	if end > len(m.rows) {
		end = len(m.rows)
	}
	return start, end
}

// View implements Model.
func (m *MultiSelect) View() string {
	m.sync()
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
		start, end := m.window()
		if start > 0 {
			b.WriteString(th.FgColor(th.Muted) + "    … more above" + th.Reset() + "\n")
		}
		for _, row := range m.rows[start:end] {
			if row.header {
				b.WriteString(th.Bold())
				b.WriteString(th.FgColor(th.Secondary))
				b.WriteString("    " + row.text)
				b.WriteString(th.Reset())
				b.WriteString("\n")
				continue
			}
			mark := "[ ]"
			if m.checked[row.opt] {
				mark = "[x]"
			}
			switch {
			case row.opt == m.cursor:
				b.WriteString(th.Bold())
				b.WriteString(th.Reverse())
				b.WriteString(th.FgColor(th.Background))
				b.WriteString(th.BgColor(th.Primary))
				b.WriteString("  ▶ " + mark + " " + row.text)
				b.WriteString(th.Reset())
			case m.checked[row.opt]:
				b.WriteString(th.FgColor(th.Success))
				b.WriteString("    " + mark + " " + row.text)
				b.WriteString(th.Reset())
			default:
				b.WriteString(th.FgColor(th.Foreground))
				b.WriteString("    " + mark + " " + row.text)
				b.WriteString(th.Reset())
			}
			b.WriteString("\n")
		}
		if end < len(m.rows) {
			b.WriteString(th.FgColor(th.Muted) + "    … more below" + th.Reset() + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(th.FgColor(th.Muted))
	b.WriteString(fmt.Sprintf("%d selected  •  ↑/↓ or j/k to move  •  space toggle  •  Enter to confirm", m.count()))
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
