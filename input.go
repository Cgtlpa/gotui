package gotui

import "strings"

// Input is a single-line text input.
//
//	name, err := gotui.NewInput().
//	    WithPrompt("Name: ").
//	    WithPlaceholder("e.g. Ada").
//	    Run()
type Input struct {
	Prompt      string
	Placeholder string
	Value       string
	CharLimit   int
	Password    bool
	Theme       Theme
	Layout      Layout

	cursor int
	done   bool
	result string
}

// NewInput creates a new Input with defaults.
func NewInput() *Input {
	return &Input{
		Prompt:    "> ",
		Theme:     DefaultTheme(),
		Layout:    Layout{},
		CharLimit: 64,
	}
}

func (i *Input) WithPrompt(p string) *Input      { i.Prompt = p; return i }
func (i *Input) WithPlaceholder(p string) *Input { i.Placeholder = p; return i }
func (i *Input) WithValue(v string) *Input       { i.Value = v; i.cursor = len(v); return i }
func (i *Input) WithCharLimit(n int) *Input      { i.CharLimit = n; return i }
func (i *Input) AsPassword() *Input              { i.Password = true; return i }
func (i *Input) WithTheme(t Theme) *Input        { i.Theme = t; return i }
func (i *Input) WithLayout(l Layout) *Input      { i.Layout = l; return i }

func (i *Input) theme() Theme {
	if i.Theme == (Theme{}) {
		return DefaultTheme()
	}
	return i.Theme
}

func (i *Input) Init() Cmd { return nil }

func (i *Input) Update(msg Msg) (Model, Cmd) {
	km, ok := msg.(KeyMsg)
	if !ok {
		return i, nil
	}

	switch km.Key.Type {
	case KeyEnter:
		i.done = true
		i.result = i.Value
		return i, Quit
	case KeyBackspace:
		if i.cursor > 0 {
			newCursor := runePrev(i.Value, i.cursor)
			i.Value = delBefore(i.Value, i.cursor)
			i.cursor = newCursor
		}
		return i, nil
	case KeyDelete:
		if i.cursor < len(i.Value) {
			i.Value = delAfter(i.Value, i.cursor)
		}
		return i, nil
	case KeyLeft:
		if i.cursor > 0 {
			i.cursor = runePrev(i.Value, i.cursor)
		}
		return i, nil
	case KeyRight:
		if i.cursor < len(i.Value) {
			i.cursor = runeNext(i.Value, i.cursor)
		}
		return i, nil
	case KeyHome:
		i.cursor = 0
		return i, nil
	case KeyEnd:
		i.cursor = len(i.Value)
		return i, nil
	case KeyRune:
		if i.CharLimit > 0 && len(i.Value) >= i.CharLimit {
			return i, nil
		}
		r := km.Key.Rune
		before := i.Value[:i.cursor]
		after := i.Value[i.cursor:]
		i.Value = before + string(r) + after
		i.cursor += len(string(r))
		return i, nil
	default:
		return i, nil
	}
}

func (i *Input) View() string {
	th := i.theme()
	var b strings.Builder

	b.WriteString(th.Bold())
	b.WriteString(th.FgColor(th.Primary))
	b.WriteString(i.Prompt)
	b.WriteString(th.Reset())

	display := i.Value
	if i.Password {
		display = strings.Repeat("*", len(i.Value))
	}

	if display == "" && i.Placeholder != "" {
		b.WriteString(th.FgColor(th.Muted))
		b.WriteString(i.Placeholder)
		b.WriteString(th.Reset())
	} else {
		b.WriteString(th.FgColor(th.Foreground))
		b.WriteString(display)
		b.WriteString(th.Reset())
	}

	// Simple block cursor at end of text (no ANSI movement codes to avoid
	// text bleeding and misalignment).
	if !i.Password && !i.done {
		b.WriteString(th.Bold())
		b.WriteString(th.Reverse())
		b.WriteString(th.FgColor(th.Background))
		b.WriteString(th.BgColor(th.Primary))
		b.WriteString(" ")
		b.WriteString(th.Reset())
	}

	result := b.String()
	if i.Layout != (Layout{}) {
		result = i.Layout.Apply(result)
	}
	return result
}

func (i *Input) Result() interface{} { return i.result }

func (i *Input) Run() (string, error) {
	v, err := Run(i)
	if err != nil {
		return "", err
	}
	if v == nil {
		return "", nil
	}
	return v.(string), nil
}
