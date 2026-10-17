package gotui

import "strings"

// Select is a single-selection list.
//
//	color, err := gotui.NewSelect().
//	    WithTitle("Pick a color").
//	    WithOptions("Red", "Green", "Blue").
//	    Run()
type Select struct {
	Title   string
	Options []string
	Theme   Theme
	Layout  Layout

	cursor int
	done   bool
	result string
}

func NewSelect() *Select {
	return &Select{Theme: DefaultTheme()}
}

func (s *Select) WithTitle(t string) *Select         { s.Title = t; return s }
func (s *Select) WithOptions(opts ...string) *Select { s.Options = opts; return s }
func (s *Select) WithCursor(n int) *Select           { s.cursor = n; return s }
func (s *Select) WithTheme(t Theme) *Select          { s.Theme = t; return s }
func (s *Select) WithLayout(l Layout) *Select        { s.Layout = l; return s }

func (s *Select) theme() Theme {
	if s.Theme == (Theme{}) {
		return DefaultTheme()
	}
	return s.Theme
}

func (s *Select) Init() Cmd { return nil }

func (s *Select) Update(msg Msg) (Model, Cmd) {
	km, ok := msg.(KeyMsg)
	if !ok {
		return s, nil
	}

	switch km.Key.Type {
	case KeyUp:
		if s.cursor > 0 {
			s.cursor--
		}
		return s, nil
	case KeyDown:
		if s.cursor < len(s.Options)-1 {
			s.cursor++
		}
		return s, nil
	case KeyRune:
		switch km.Key.Rune {
		case 'k':
			if s.cursor > 0 {
				s.cursor--
			}
		case 'j':
			if s.cursor < len(s.Options)-1 {
				s.cursor++
			}
		}
		return s, nil
	case KeyEnter:
		if len(s.Options) > 0 {
			s.done = true
			s.result = s.Options[s.cursor]
		}
		return s, Quit
	default:
		return s, nil
	}
}

func (s *Select) View() string {
	th := s.theme()
	var b strings.Builder

	if s.Title != "" {
		b.WriteString(th.Bold())
		b.WriteString(th.FgColor(th.Primary))
		b.WriteString(s.Title)
		b.WriteString(th.Reset())
		b.WriteString("\n\n")
	}

	if len(s.Options) == 0 {
		b.WriteString(th.FgColor(th.Muted))
		b.WriteString("(no options)")
		b.WriteString(th.Reset())
		b.WriteString("\n\n")
	} else {
		for i, opt := range s.Options {
			if i == s.cursor {
				b.WriteString(th.Bold())
				b.WriteString(th.Reverse())
				b.WriteString(th.FgColor(th.Background))
				b.WriteString(th.BgColor(th.Primary))
				b.WriteString("  ▶")
				b.WriteString(opt)
				b.WriteString(th.Reset())
			} else {
				b.WriteString(th.FgColor(th.Secondary))
				b.WriteString("   ")
				b.WriteString(th.FgColor(th.Foreground))
				b.WriteString(opt)
				b.WriteString(th.Reset())
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(th.FgColor(th.Muted))
	b.WriteString("↑/↓ or j/k to move  •  Enter to confirm")
	b.WriteString(th.Reset())

	result := b.String()
	if s.Layout != (Layout{}) {
		result = s.Layout.Apply(result)
	}
	return result
}

func (s *Select) Result() interface{} { return s.result }

func (s *Select) Run() (string, error) {
	v, err := Run(s)
	if err != nil {
		return "", err
	}
	if v == nil {
		return "", nil
	}
	return v.(string), nil
}
