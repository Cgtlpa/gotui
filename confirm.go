package gotui

import "strings"

// Confirm is a yes/no prompt.
//
//	ok, err := gotui.NewConfirm().
//	    WithTitle("Delete this file?").
//	    DefaultYes().
//	    Run()
type Confirm struct {
	Title  string
	Value  bool
	Theme  Theme
	Layout Layout

	selected int
	done     bool
	result   bool
}

func NewConfirm() *Confirm {
	return &Confirm{Title: "Are you sure?", Theme: DefaultTheme()}
}

func (c *Confirm) WithTitle(t string) *Confirm  { c.Title = t; return c }
func (c *Confirm) WithValue(v bool) *Confirm    { c.Value = v; c.selected = boolIdx(v); return c }
func (c *Confirm) DefaultYes() *Confirm         { return c.WithValue(true) }
func (c *Confirm) DefaultNo() *Confirm          { return c.WithValue(false) }
func (c *Confirm) WithTheme(t Theme) *Confirm   { c.Theme = t; return c }
func (c *Confirm) WithLayout(l Layout) *Confirm { c.Layout = l; return c }
func (c *Confirm) setTheme(t Theme)             { c.Theme = t }

func boolIdx(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (c *Confirm) theme() Theme {
	if c.Theme == (Theme{}) {
		return DefaultTheme()
	}
	return c.Theme
}

func (c *Confirm) Init() Cmd { return nil }

func (c *Confirm) Update(msg Msg) (Model, Cmd) {
	km, ok := msg.(KeyMsg)
	if !ok {
		return c, nil
	}

	switch km.Key.Type {
	case KeyLeft:
		c.selected = 0
		return c, nil
	case KeyRight:
		c.selected = 1
		return c, nil
	case KeyEnter:
		c.done = true
		c.result = c.selected == 1
		return c, Quit
	default:
		return c, nil
	}
}

func (c *Confirm) View() string {
	th := c.theme()
	var b strings.Builder

	if c.Title != "" {
		b.WriteString(th.Bold())
		b.WriteString(th.FgColor(th.Primary))
		b.WriteString(c.Title)
		b.WriteString(th.Reset())
		b.WriteString(" ")
	}

	noSel := c.selected == 0

	b.WriteString("(")
	if noSel {
		b.WriteString(th.Bold())
		b.WriteString(th.FgColor(th.Danger))
		b.WriteString("No")
	} else {
		b.WriteString(th.FgColor(th.Muted))
		b.WriteString("No")
	}
	b.WriteString(th.Reset())
	b.WriteString(") / ")

	b.WriteString("(")
	if !noSel {
		b.WriteString(th.Bold())
		b.WriteString(th.FgColor(th.Success))
		b.WriteString("Yes")
	} else {
		b.WriteString(th.FgColor(th.Muted))
		b.WriteString("Yes")
	}
	b.WriteString(th.Reset())
	b.WriteString(")")

	b.WriteString("\n\n")
	b.WriteString(th.FgColor(th.Muted))
	b.WriteString("←/→ to toggle  •  Enter to confirm")
	b.WriteString(th.Reset())

	result := b.String()
	if c.Layout != (Layout{}) {
		result = c.Layout.Apply(result)
	}
	return result
}

func (c *Confirm) Result() interface{} { return c.result }

func (c *Confirm) Run() (bool, error) {
	v, err := Run(c)
	if err != nil {
		return false, err
	}
	if v == nil {
		return false, nil
	}
	return v.(bool), nil
}
