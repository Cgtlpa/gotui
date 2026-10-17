package gotui

import "strings"

// App is the simple, cohesive entry point to gotui.
//
//	app := gotui.New()
//	app.Title("Setup Wizard")
//	app.Center(true)
//	app.Text("Welcome to the Setup Wizard")
//	app.Select("choice", "Choose an option", "InstantWM", "Niri", "DWM")
//	app.Input("name", "Your name: ", "e.g. Ada")
//	app.Confirm("confirm", "Continue?")
//
//	if err := app.Run(); err != nil {
//	    if errors.Is(err, gotui.ErrCancelled) {
//	        fmt.Println("Setup cancelled")
//	    }
//	    log.Fatal(err)
//	}
//
//	fmt.Println("Choice:", app.Value("choice"))
type App struct {
	title     string
	center    bool
	width     int
	height    int
	textScale int
	theme     Theme

	steps  []step
	values map[string]interface{}
}

// step is one registered screen in the app.
type step struct {
	model Model
	key   string // result key, empty if the step doesn't produce a value
}

// New creates a new App with defaults.
func New() *App {
	return &App{
		theme:  DefaultTheme(),
		values: make(map[string]interface{}),
	}
}

// Title sets the app title shown above each step.
func (a *App) Title(t string) *App { a.title = t; return a }

// Center turns centering on or off for every step.
func (a *App) Center(c bool) *App { a.center = c; return a }

// Width sets the line width for every step (0 = auto).
func (a *App) Width(w int) *App { a.width = w; return a }

// Height sets the line height for every step (0 = auto).
func (a *App) Height(h int) *App { a.height = h; return a }

// TextScale scales text horizontally (1 = normal, 2 = double, etc.).
func (a *App) TextScale(s int) *App { a.textScale = s; return a }

// Theme sets the color palette for every step.
func (a *App) Theme(th Theme) *App { a.theme = th; return a }

// Value returns the stored result for the given key, or nil if not set.
func (a *App) Value(key string) interface{} {
	return a.values[key]
}

// Text adds a read-only text screen. Press any key to continue.
func (a *App) Text(text string) *App {
	a.steps = append(a.steps, step{model: NewText().WithBody(text)})
	return a
}

// Select adds a single-selection list stored under key.
//
//	app.Select("wm", "Choose a window manager", "InstantWM", "Niri", "DWM")
func (a *App) Select(key, title string, options ...string) *App {
	a.steps = append(a.steps, step{
		key:   key,
		model: NewSelect().WithTitle(title).WithOptions(options...),
	})
	return a
}

// Input adds a single-line text input stored under key.
//
//	app.Input("name", "Your name: ", "e.g. Ada")
//	app.Input("name", "Your name:")          // placeholder is optional
func (a *App) Input(key, prompt string, placeholder ...string) *App {
	in := NewInput().WithPrompt(prompt)
	if len(placeholder) > 0 {
		in.WithPlaceholder(placeholder[0])
	}
	a.steps = append(a.steps, step{
		key:   key,
		model: in,
	})
	return a
}

// Password adds a masked text input stored under key.
//
//	app.Password("pass", "Password")
func (a *App) Password(key, prompt string) *App {
	a.steps = append(a.steps, step{
		key:   key,
		model: NewPassword().WithPrompt(prompt),
	})
	return a
}

// Confirm adds a yes/no prompt stored under key.
//
//	app.Confirm("confirm", "Continue?")
func (a *App) Confirm(key, title string) *App {
	a.steps = append(a.steps, step{
		key:   key,
		model: NewConfirm().WithTitle(title),
	})
	return a
}

// Form adds a multi-field form. Fields are given as alternating key,label pairs:
//
//	app.Form("name", "Full name", "email", "Email")
func (a *App) Form(fields ...string) *App {
	f := NewForm()
	for i := 0; i+1 < len(fields); i += 2 {
		f = f.WithInput(fields[i], fields[i+1])
	}
	a.steps = append(a.steps, step{model: f})
	return a
}

// Custom adds a fully custom Model (Bubble Tea-style).
func (a *App) Custom(m Model) *App {
	a.steps = append(a.steps, step{model: m})
	return a
}

// Run executes every added step in order. Returns error on terminal failure
// or ErrCancelled if the user pressed Esc or Ctrl+C.
func (a *App) Run() error {
	th := a.theme
	if th == (Theme{}) {
		th = DefaultTheme()
	}

	width := a.width
	// Auto-detect terminal width when centering is enabled but no explicit
	// width was set. This makes app.Center(true) work out of the box.
	if a.center && width == 0 {
		if w := TerminalWidth(); w > 0 {
			width = w
		}
	}

	for _, st := range a.steps {
		wrapped := &appStep{
			app:   a,
			step:  st,
			theme: th,
		}

		prog := NewProgram(wrapped)
		prog.SetLayout(a.center, width, a.height)
		prog.SetTextScale(a.textScale)
		if err := prog.Run(); err != nil {
			return err
		}

		if comp, ok := resultOrNil(st.model); ok {
			if st.key != "" {
				a.values[st.key] = comp
			}
		}
	}

	return nil
}

// resultOrNil extracts a result from a component if it implements Component.
func resultOrNil(m Model) (interface{}, bool) {
	c, ok := m.(Component)
	if !ok {
		return nil, false
	}
	return c.Result(), true
}

// appStep wraps a step's model with title/layout/theme applied at render time.
type appStep struct {
	app   *App
	step  step
	theme Theme
}

func (s *appStep) Init() Cmd { return nil }

func (s *appStep) Update(msg Msg) (Model, Cmd) {
	newModel, cmd := s.step.model.Update(msg)
	if newModel != nil {
		s.step.model = newModel
	}
	return s, cmd
}

// View renders the title followed by the component view, one line after another,
// with no cursor positioning or diagonal offsets.
// Centering and width padding are handled by Program.render().
func (s *appStep) View() string {
	th := s.theme
	inner := s.step.model.View()

	var b strings.Builder

	if s.app.title != "" {
		b.WriteString(th.Bold())
		b.WriteString(th.FgColor(th.Primary))
		b.WriteString(s.app.title)
		b.WriteString(th.Reset())
		b.WriteString("\n\n")
	}

	b.WriteString(inner)

	return b.String()
}
