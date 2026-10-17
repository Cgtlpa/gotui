package gotui

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
)

// Msg is any message fed into a Model's Update.
type Msg interface{}

// Cmd is a command that returns a Msg. nil means do nothing.
type Cmd func() Msg

// Batch runs several commands and delivers every message they produce.
func Batch(cmds ...Cmd) Cmd {
	return func() Msg {
		msgs := BatchMsg{}
		for _, c := range cmds {
			if c != nil {
				if m := c(); m != nil {
					msgs.Msgs = append(msgs.Msgs, m)
				}
			}
		}
		return msgs
	}
}

// BatchMsg holds multiple messages from Batched commands.
type BatchMsg struct{ Msgs []Msg }

// QuitMsg tells the program to exit.
type QuitMsg struct{}

// KeyMsg is delivered on every keypress.
type KeyMsg struct{ Key Key }

// InitMsg is delivered once when the program starts.
type InitMsg struct{}

// Model holds state, reacts to messages, and renders itself to a string.
type Model interface {
	Init() Cmd
	Update(msg Msg) (Model, Cmd)
	View() string
}

// Program runs a Model in a raw-mode terminal loop.
type Program struct {
	model     Model
	term      *Terminal
	quit      bool
	center    bool
	width     int
	height    int
	textScale int
}

// NewProgram creates a Program for the given model.
func NewProgram(model Model) *Program {
	return &Program{model: model}
}

// Run starts the program and blocks until it quits.
func (p *Program) Run() error {
	term, err := NewTerminal()
	if err != nil {
		return err
	}
	p.term = term
	defer func() { _ = p.term.Restore() }()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		<-sigChan
		p.Quit()
	}()

	// Enter alt screen, hide cursor, move to home, clear screen and scrollback.
	fmt.Fprint(p.term, ansiAltScreen+ansiCursorHide+ansiCursorHome+ansiClearScreen+ansiClearScrollback)
	defer fmt.Fprint(p.term, ansiCursorShow+ansiExitAltScreen)

	if cmd := p.model.Init(); cmd != nil {
		p.tick(cmd())
	}

	for !p.quit {
		p.render()

		key, err := p.term.ReadKey()
		if err != nil {
			if isClosed(err) || isEOF(err) {
				p.Quit()
				continue
			}
			return err
		}
		if key.Type == KeyEsc || key.Type == KeyCtrlC {
			return ErrCancelled
		}

		p.tick(KeyMsg{Key: key})
	}

	return nil
}

// tick processes a message and any commands it triggers.
func (p *Program) tick(msg Msg) {
	for {
		if _, ok := msg.(QuitMsg); ok {
			p.quit = true
			return
		}
		newModel, cmd := p.model.Update(msg)
		if newModel != nil {
			p.model = newModel
		}
		if cmd == nil || p.quit {
			return
		}
		msg = cmd()
		if msg == nil {
			return
		}
	}
}

// Send feeds a message into the program from outside.
func (p *Program) Send(msg Msg) {
	p.tick(msg)
}

// Quit stops the program after the current frame.
func (p *Program) Quit() {
	p.quit = true
}

// SetLayout configures centering and padding for rendering.
func (p *Program) SetLayout(center bool, width, height int) {
	p.center = center
	p.width = width
	p.height = height
}

// SetTextScale configures horizontal text scaling for rendering.
func (p *Program) SetTextScale(scale int) {
	p.textScale = scale
}

// render moves to home, clears the screen, and prints the current view.
// In raw mode, \n alone does NOT return to column 0 — it only moves down.
// We convert every \n to \r\n so each line starts at column 0 (no diagonal text).
// Center=true centers each line horizontally within the terminal width.
func (p *Program) render() {
	view := p.model.View()
	lines := strings.Split(view, "\n")

	// Apply text scale: repeat each visible character n times.
	if p.textScale > 1 {
		scaled := make([]string, len(lines))
		for i, line := range lines {
			scaled[i] = scaleLine(line, p.textScale)
		}
		lines = scaled
	}

	// Detect terminal width for centering.
	termWidth := TerminalWidth()
	if p.width > 0 {
		termWidth = p.width
	}

	// Vertical centering: count blank lines to add at top.
	topBlankLines := 0
	if p.height > 0 && p.center {
		contentLines := len(lines)
		if contentLines < p.height {
			topBlankLines = (p.height - contentLines) / 2
		}
	}

	var b strings.Builder

	// Move to home, clear screen.
	b.WriteString("\r" + ansiCursorHome + ansiClearScreen + ansiClearScrollback)

	// Add vertical centering blank lines.
	for i := 0; i < topBlankLines; i++ {
		b.WriteString("\r\n")
	}

	// Print each line, centered horizontally if requested.
	for i, line := range lines {
		if i > 0 || topBlankLines > 0 {
			b.WriteString("\r\n")
		}

		if p.center && termWidth > 0 {
			lineLen := visibleLen(line)
			if lineLen < termWidth {
				left := (termWidth - lineLen) / 2
				b.WriteString(strings.Repeat(" ", left))
			}
		}

		b.WriteString(line)
	}

	fmt.Fprint(p.term, b.String())
}

// Quit is a command that exits the program.
var Quit Cmd = func() Msg { return QuitMsg{} }

func isClosed(err error) bool {
	return err != nil && err.Error() == "file already closed"
}

func isEOF(err error) bool {
	return err != nil && err.Error() == "EOF"
}
