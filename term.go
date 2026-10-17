package gotui

import (
	"fmt"
	"io"
	"os"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

// ANSI escape sequences used for drawing and cursor control.
const (
	ansiCursorHide      = "\x1b[?25l"
	ansiCursorShow      = "\x1b[?25h"
	ansiClearScreen     = "\x1b[2J"
	ansiCursorHome      = "\x1b[H"
	ansiReset           = "\x1b[0m"
	ansiBold            = "\x1b[1m"
	ansiReverse         = "\x1b[7m"
	ansiAltScreen       = "\x1b[?1049h"
	ansiExitAltScreen   = "\x1b[?1049l"
	ansiClearScrollback = "\x1b[3J"
)

// escTimeout is how long ReadKey waits for the rest of an escape sequence
// after receiving a lone ESC byte.
const escTimeout = 50 * time.Millisecond

// TerminalWidth returns the current terminal width in columns, or 0 if unknown.
func TerminalWidth() int {
	return terminalWidth()
}

// Terminal wraps stdin/stdout in raw mode and reads keypresses one at a time.
// Works on Unix (Linux/macOS) and Windows (Windows 10+).
// Input is read synchronously: the process polls stdin for readability, which
// is what lets ReadKey disambiguate a lone ESC from an escape sequence with a
// timer (terminal files do not support read deadlines). Any bytes that arrive
// together with a keypress are buffered internally, so no keypress is ever
// skipped or duplicated.
type Terminal struct {
	in    *os.File
	out   io.Writer
	state *term.State
	fd    uintptr
	buf   []byte // pending bytes not yet consumed
}

// NewTerminal opens stdin/stdout and puts stdin into raw mode.
func NewTerminal() (*Terminal, error) {
	in := os.Stdin
	out := os.Stdout

	fd := int(in.Fd())
	if !term.IsTerminal(fd) {
		return nil, fmt.Errorf("stdin is not a terminal")
	}

	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("failed to enter raw mode: %w", err)
	}

	return &Terminal{in: in, out: out, state: state, fd: uintptr(fd)}, nil
}

// Restore restores the original terminal state and shows the cursor again.
func (t *Terminal) Restore() error {
	if t.state == nil {
		return nil
	}
	err := term.Restore(int(t.fd), t.state)
	fmt.Fprint(t.out, ansiCursorShow+ansiExitAltScreen)
	return err
}

// Write sends bytes to the terminal output.
func (t *Terminal) Write(b []byte) (int, error) {
	return t.out.Write(b)
}

// WriteString sends a string to the terminal output.
func (t *Terminal) WriteString(s string) (int, error) {
	return t.out.Write([]byte(s))
}

// Width returns the terminal width in columns, or 0 if unknown.
func (t *Terminal) Width() int {
	return terminalWidth()
}

// ReadKey reads exactly one keypress. Bytes are buffered internally, so a
// keypress is never skipped or duplicated even when several arrive in one
// read. If the first byte is ESC it waits briefly for the rest of an escape
// sequence (arrow keys, etc.); a lone ESC is reported as KeyEsc. Multi-byte
// UTF-8 characters are read completely so they are not split across
// keypresses. ReadKey returns io.EOF once stdin is closed.
func (t *Terminal) ReadKey() (Key, error) {
	b, err := t.readByte()
	if err != nil {
		return Key{}, err
	}

	if b == 0x1b {
		return t.readEsc()
	}

	// Control bytes (Enter, Tab, Backspace, Ctrl+C, ...) are single keys.
	if b < 0x20 || b == 0x7f {
		return decodeByte(b), nil
	}
	if b < 0xC2 {
		return Key{Type: KeyRune, Rune: rune(b)}, nil
	}

	// UTF-8 leading byte: read the remaining bytes of the sequence.
	seq := []byte{b}
	for i := 1; i < utf8SequenceLen(b); i++ {
		nb, err := t.readByte()
		if err != nil {
			return Key{}, err
		}
		seq = append(seq, nb)
	}
	r, _ := utf8.DecodeRune(seq)
	return Key{Type: KeyRune, Rune: r}, nil
}

// readByte returns the next input byte, blocking until one is available.
func (t *Terminal) readByte() (byte, error) {
	if len(t.buf) == 0 {
		ok, err := t.fill(-1)
		if err != nil {
			return 0, err
		}
		if !ok {
			return 0, io.EOF
		}
	}
	b := t.buf[0]
	t.buf = t.buf[1:]
	return b, nil
}

// readEsc resolves an ESC that was just consumed. While the buffered bytes
// could still be the beginning of an escape sequence (and that sequence may
// still be in flight) it waits up to escTimeout for more input. A lone ESC
// is reported as KeyEsc; otherwise the sequence is parsed and any trailing
// bytes are kept for the next ReadKey call.
func (t *Terminal) readEsc() (Key, error) {
	deadline := time.Now().Add(escTimeout)
	for escIncomplete(t.buf) {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		ok, err := t.fill(remaining)
		if err != nil {
			return Key{}, err
		}
		if !ok {
			break // timed out; treat what we have as final
		}
	}

	key, leftover := parseKey(append([]byte{0x1b}, t.buf...))
	t.buf = leftover
	return key, nil
}

// escIncomplete reports whether buf (the bytes read after an ESC) could
// still be the start of an escape sequence. A complete CSI/SS3 sequence
// ends with a byte in 0x40-0x7e; anything else after the ESC is a regular
// key, which means the ESC stood alone.
func escIncomplete(buf []byte) bool {
	if len(buf) == 0 {
		return true
	}
	if buf[0] != '[' && buf[0] != 'O' {
		return false
	}
	for _, b := range buf[1:] {
		if b >= 0x40 && b <= 0x7e {
			return false
		}
	}
	return true
}

// fill waits for input and appends at least one byte to the buffer.
// A negative timeout waits forever. It returns false when the timeout
// elapses before any input arrives.
func (t *Terminal) fill(timeout time.Duration) (bool, error) {
	for {
		ready, err := waitReadable(t.fd, timeout)
		if err != nil {
			return false, err
		}
		if !ready {
			return false, nil
		}

		var chunk [64]byte
		n, err := t.in.Read(chunk[:])
		if n > 0 {
			t.buf = append(t.buf, chunk[:n]...)
			return true, nil
		}
		if err != nil {
			return false, err
		}
	}
}

// utf8SequenceLen returns the total length in bytes of a UTF-8 sequence
// starting with the given leading byte. Returns 1 for ASCII bytes (0x00-0x7F).
func utf8SequenceLen(b byte) int {
	switch {
	case b <= 0x7F:
		return 1
	case b >= 0xC2 && b <= 0xDF:
		return 2
	case b >= 0xE0 && b <= 0xEF:
		return 3
	case b >= 0xF0 && b <= 0xF4:
		return 4
	default:
		return 1 // continuation byte or invalid, treat as single byte
	}
}

// parseKey parses a complete byte slice that starts (optionally) with ESC and
// returns the decoded Key plus any leftover bytes that belong to a later key.
func parseKey(data []byte) (Key, []byte) {
	if len(data) == 0 {
		return Key{Type: KeyUnknown}, nil
	}

	if data[0] != 0x1b {
		return decodeByte(data[0]), data[1:]
	}

	if len(data) == 1 {
		return Key{Type: KeyEsc}, nil
	}

	// ESC followed by a regular key byte is a lone ESC plus that key
	// (e.g. Esc and Enter pressed in quick succession).
	if data[1] != '[' && data[1] != 'O' {
		return Key{Type: KeyEsc}, data[1:]
	}

	// CSI ('[') and SS3 ('O') sequences end with a byte in 0x40-0x7e.
	end := -1
	for i := 2; i < len(data); i++ {
		if data[i] >= 0x40 && data[i] <= 0x7e {
			end = i
			break
		}
	}
	if end < 0 {
		// Incomplete sequence: report the ESC and keep the rest for later.
		return Key{Type: KeyEsc}, data[1:]
	}

	seq := string(data[2 : end+1])
	leftover := data[end+1:]

	switch seq[len(seq)-1] {
	case 'A':
		return Key{Type: KeyUp, Code: seq}, leftover
	case 'B':
		return Key{Type: KeyDown, Code: seq}, leftover
	case 'C':
		return Key{Type: KeyRight, Code: seq}, leftover
	case 'D':
		return Key{Type: KeyLeft, Code: seq}, leftover
	case 'H':
		return Key{Type: KeyHome, Code: seq}, leftover
	case 'F':
		return Key{Type: KeyEnd, Code: seq}, leftover
	case '~':
		// Sequences with a trailing '~' carry a numeric parameter,
		// e.g. "\x1b[3~" is Delete.
		switch seq {
		case "1~", "7~":
			return Key{Type: KeyHome, Code: seq}, leftover
		case "2~":
			return Key{Type: KeyInsert, Code: seq}, leftover
		case "3~":
			return Key{Type: KeyDelete, Code: seq}, leftover
		case "4~", "8~":
			return Key{Type: KeyEnd, Code: seq}, leftover
		case "5~":
			return Key{Type: KeyPageUp, Code: seq}, leftover
		case "6~":
			return Key{Type: KeyPageDown, Code: seq}, leftover
		}
	}

	return Key{Type: KeyUnknown}, leftover
}

// decodeByte turns a single non-ESC byte into a Key.
func decodeByte(b byte) Key {
	switch b {
	case 0x03:
		return Key{Type: KeyCtrlC}
	case 0x04:
		return Key{Type: KeyCtrlD}
	case 0x0a, 0x0d:
		return Key{Type: KeyEnter}
	case 0x09:
		return Key{Type: KeyTab}
	case 0x7f, 0x08: // DEL and BS are both backspace
		return Key{Type: KeyBackspace}
	}

	if b < 0x20 {
		return Key{Type: KeyUnknown}
	}

	return Key{Type: KeyRune, Rune: rune(b)}
}

// runePrev returns the byte index of the rune immediately before idx.
func runePrev(s string, idx int) int {
	if idx <= 0 {
		return 0
	}
	if idx > len(s) {
		idx = len(s)
	}
	if idx <= 0 {
		return 0
	}
	r := []rune(s[:idx])
	if len(r) == 0 {
		return 0
	}
	return len(string(r[:len(r)-1]))
}

// runeNext returns the byte index of the rune immediately after idx.
func runeNext(s string, idx int) int {
	if idx >= len(s) {
		return len(s)
	}
	if idx < 0 {
		idx = 0
	}
	r := []rune(s[idx:])
	if len(r) == 0 {
		return len(s)
	}
	return idx + len(string(r[:1]))
}

// delBefore removes the rune immediately before idx.
func delBefore(s string, idx int) string {
	if idx <= 0 {
		return s
	}
	if idx > len(s) {
		idx = len(s)
	}
	p := runePrev(s, idx)
	return s[:p] + s[idx:]
}

// delAfter removes the rune at/after idx.
func delAfter(s string, idx int) string {
	if idx >= len(s) {
		return s
	}
	if idx < 0 {
		idx = 0
	}
	n := runeNext(s, idx)
	return s[:idx] + s[n:]
}
