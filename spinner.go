package gotui

import "time"

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner shows a loading animation while a command runs.
type Spinner struct {
	Text   string
	Cmd    Cmd
	Theme  Theme
	Layout Layout

	frame int
	done  bool
	msg   Msg
}

func NewSpinner() *Spinner {
	return &Spinner{Text: "Loading...", Theme: DefaultTheme()}
}

func (s *Spinner) WithText(t string) *Spinner   { s.Text = t; return s }
func (s *Spinner) WithCmd(c Cmd) *Spinner       { s.Cmd = c; return s }
func (s *Spinner) WithTheme(t Theme) *Spinner   { s.Theme = t; return s }
func (s *Spinner) WithLayout(l Layout) *Spinner { s.Layout = l; return s }

func (s *Spinner) theme() Theme {
	if s.Theme == (Theme{}) {
		return DefaultTheme()
	}
	return s.Theme
}

func (s *Spinner) Init() Cmd { return s.Cmd }

func (s *Spinner) Update(msg Msg) (Model, Cmd) {
	if msg == nil {
		s.frame = (s.frame + 1) % len(spinnerFrames)
		return s, tickCmd()
	}
	if _, ok := msg.(QuitMsg); ok {
		s.done = true
		return s, nil
	}
	s.msg = msg
	s.done = true
	return s, nil
}

func (s *Spinner) View() string {
	th := s.theme()
	frame := spinnerFrames[s.frame]
	result := th.Bold() + th.FgColor(th.Primary) + frame + th.Reset() +
		" " + th.FgColor(th.Foreground) + s.Text + th.Reset() + "\n"
	if s.Layout != (Layout{}) {
		result = s.Layout.Apply(result)
	}
	return result
}

func (s *Spinner) Result() interface{} { return s.msg }

func (s *Spinner) Run() (Msg, error) {
	prog := NewProgram(s)

	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				prog.Send(nil)
			case <-stop:
				return
			}
		}
	}()

	if err := prog.Run(); err != nil {
		close(stop)
		return nil, err
	}
	close(stop)
	return s.msg, nil
}

func tickCmd() Cmd {
	return func() Msg { return nil }
}
