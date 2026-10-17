package gotui

import "strings"

// Form is a multi-field form.
//
//	form := gotui.NewForm().
//	    WithTitle("Sign up").
//	    WithInput("name", "Full name").
//	    WithInput("email", "Email").
//	    WithPassword("pass", "Password")
//
//	err := form.Run()
//	fmt.Println(form.Get("name"))
type Form struct {
	Title  string
	Theme  Theme
	Layout Layout

	fields []*formField
	focus  int
	done   bool
	values map[string]string
}

type formField struct {
	Key         string
	Label       string
	Value       string
	Placeholder string
	Password    bool
	cursor      int
	focused     bool
}

func NewForm() *Form {
	return &Form{
		Theme:  DefaultTheme(),
		Layout: Layout{},
		values: make(map[string]string),
	}
}

func (f *Form) WithTitle(t string) *Form  { f.Title = t; return f }
func (f *Form) WithTheme(t Theme) *Form   { f.Theme = t; return f }
func (f *Form) WithLayout(l Layout) *Form { f.Layout = l; return f }

func (f *Form) WithInput(key, label string) *Form {
	fd := &formField{Key: key, Label: label}
	f.fields = append(f.fields, fd)
	if len(f.fields) == 1 {
		fd.focused = true
	}
	return f
}

func (f *Form) WithPassword(key, label string) *Form {
	fd := &formField{Key: key, Label: label, Password: true}
	f.fields = append(f.fields, fd)
	if len(f.fields) == 1 {
		fd.focused = true
	}
	return f
}

func (f *Form) theme() Theme {
	if f.Theme == (Theme{}) {
		return DefaultTheme()
	}
	return f.Theme
}

func (f *Form) Init() Cmd { return nil }

func (f *Form) Update(msg Msg) (Model, Cmd) {
	km, ok := msg.(KeyMsg)
	if !ok {
		return f, nil
	}

	switch km.Key.Type {
	case KeyTab:
		if len(f.fields) > 0 {
			f.fields[f.focus].focused = false
			f.focus = (f.focus + 1) % len(f.fields)
			f.fields[f.focus].focused = true
		}
		return f, nil
	case KeyEnter:
		if f.focus == len(f.fields)-1 {
			f.done = true
			for _, fd := range f.fields {
				f.values[fd.Key] = fd.Value
			}
			return f, Quit
		}
		if len(f.fields) > 0 {
			f.fields[f.focus].focused = false
			f.focus = (f.focus + 1) % len(f.fields)
			f.fields[f.focus].focused = true
		}
		return f, nil
	}

	if len(f.fields) == 0 {
		return f, nil
	}
	fd := f.fields[f.focus]

	switch km.Key.Type {
	case KeyBackspace:
		if fd.cursor > 0 {
			newCursor := runePrev(fd.Value, fd.cursor)
			fd.Value = delBefore(fd.Value, fd.cursor)
			fd.cursor = newCursor
		}
	case KeyDelete:
		if fd.cursor < len(fd.Value) {
			fd.Value = delAfter(fd.Value, fd.cursor)
		}
	case KeyLeft:
		if fd.cursor > 0 {
			fd.cursor = runePrev(fd.Value, fd.cursor)
		}
	case KeyRight:
		if fd.cursor < len(fd.Value) {
			fd.cursor = runeNext(fd.Value, fd.cursor)
		}
	case KeyHome:
		fd.cursor = 0
	case KeyEnd:
		fd.cursor = len(fd.Value)
	case KeyRune:
		before := fd.Value[:fd.cursor]
		after := fd.Value[fd.cursor:]
		fd.Value = before + string(km.Key.Rune) + after
		fd.cursor += len(string(km.Key.Rune))
	default:
		return f, nil
	}

	return f, nil
}

func (f *Form) View() string {
	th := f.theme()
	var b strings.Builder

	if f.Title != "" {
		b.WriteString(th.Bold())
		b.WriteString(th.FgColor(th.Primary))
		b.WriteString(f.Title)
		b.WriteString(th.Reset())
		b.WriteString("\n\n")
	}

	for i, fd := range f.fields {
		if i > 0 {
			b.WriteString("\n")
		}

		b.WriteString(th.FgColor(th.Secondary))
		b.WriteString(fd.Label)
		b.WriteString(th.Reset())
		b.WriteString(": ")

		display := fd.Value
		if fd.Password {
			display = strings.Repeat("*", len(fd.Value))
		}

		if display == "" && fd.Placeholder != "" {
			b.WriteString(th.FgColor(th.Muted))
			b.WriteString(fd.Placeholder)
			b.WriteString(th.Reset())
		} else {
			b.WriteString(th.FgColor(th.Foreground))
			b.WriteString(display)
			b.WriteString(th.Reset())
		}

		if fd.focused {
			// Show a simple block cursor at the end of the field text.
			b.WriteString(th.Bold())
			b.WriteString(th.Reverse())
			b.WriteString(th.FgColor(th.Background))
			b.WriteString(th.BgColor(th.Primary))
			if fd.cursor < len(display) {
				b.WriteString(string(display[fd.cursor]))
			} else {
				b.WriteString(" ")
			}
			b.WriteString(th.Reset())
		}
	}

	b.WriteString("\n\n")
	b.WriteString(th.FgColor(th.Muted))
	if len(f.fields) > 1 {
		b.WriteString("Tab to move  •  Enter to submit")
	} else {
		b.WriteString("Enter to submit")
	}
	b.WriteString(th.Reset())

	result := b.String()
	if f.Layout != (Layout{}) {
		result = f.Layout.Apply(result)
	}
	return result
}

func (f *Form) Result() interface{} { return f.values }

func (f *Form) Get(key string) string {
	return f.values[key]
}

func (f *Form) Run() error {
	_, err := Run(f)
	return err
}
