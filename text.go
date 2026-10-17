package gotui

import "strings"

// Text displays a title and body. Press any key to continue.
type Text struct {
	Title  string
	Body   string
	Theme  Theme
	Layout Layout

	done bool
}

func NewText() *Text {
	return &Text{Theme: DefaultTheme()}
}

func (t *Text) WithTitle(ti string) *Text  { t.Title = ti; return t }
func (t *Text) WithBody(body string) *Text { t.Body = body; return t }
func (t *Text) WithTheme(th Theme) *Text   { t.Theme = th; return t }
func (t *Text) WithLayout(l Layout) *Text  { t.Layout = l; return t }
func (t *Text) setTheme(th Theme)          { t.Theme = th }

func (t *Text) theme() Theme {
	if t.Theme == (Theme{}) {
		return DefaultTheme()
	}
	return t.Theme
}

func (t *Text) Init() Cmd { return nil }

func (t *Text) Update(msg Msg) (Model, Cmd) {
	if _, ok := msg.(KeyMsg); ok {
		t.done = true
		return t, Quit
	}
	return t, nil
}

func (t *Text) View() string {
	th := t.theme()
	var b strings.Builder

	if t.Title != "" {
		b.WriteString(th.Bold())
		b.WriteString(th.FgColor(th.Primary))
		b.WriteString(t.Title)
		b.WriteString(th.Reset())
		b.WriteString("\n\n")
	}

	// Wrap body at ~78 columns.
	for _, line := range wrapText(t.Body, 78) {
		b.WriteString(th.FgColor(th.Foreground))
		b.WriteString(line)
		b.WriteString(th.Reset())
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(th.FgColor(th.Muted))
	b.WriteString("Press any key to continue")
	b.WriteString(th.Reset())

	result := b.String()
	if t.Layout != (Layout{}) {
		result = t.Layout.Apply(result)
	}
	return result
}

func (t *Text) Result() interface{} { return nil }

func (t *Text) Run() error {
	_, err := Run(t)
	return err
}

func wrapText(text string, maxChars int) []string {
	if maxChars <= 0 {
		return []string{text}
	}
	var lines []string
	for _, para := range strings.Split(text, "\n") {
		if len(para) <= maxChars {
			lines = append(lines, para)
			continue
		}
		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		var cur strings.Builder
		cur.WriteString(words[0])
		for _, w := range words[1:] {
			if cur.Len()+1+len(w) > maxChars {
				lines = append(lines, cur.String())
				cur.Reset()
				cur.WriteString(w)
			} else {
				cur.WriteString(" ")
				cur.WriteString(w)
			}
		}
		lines = append(lines, cur.String())
	}
	return lines
}
