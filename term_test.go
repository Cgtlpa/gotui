package gotui

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// newPipeTerminal builds a Terminal backed by an os.Pipe instead of a real
// tty, so ReadKey can be tested without raw mode or a pty.
func newPipeTerminal(t *testing.T) (*Terminal, *os.File) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	term := &Terminal{in: r, out: io.Discard, fd: r.Fd()}
	t.Cleanup(func() {
		w.Close()
		r.Close()
	})
	return term, w
}

func writeTo(t *testing.T, w *os.File, data string) {
	t.Helper()
	if _, err := w.WriteString(data); err != nil {
		t.Fatalf("write %q: %v", data, err)
	}
}

func readKeyWithin(t *testing.T, term *Terminal) (Key, error) {
	t.Helper()
	type result struct {
		key Key
		err error
	}
	ch := make(chan result, 1)
	go func() {
		k, err := term.ReadKey()
		ch <- result{k, err}
	}()
	select {
	case r := <-ch:
		return r.key, r.err
	case <-time.After(2 * time.Second):
		t.Fatal("ReadKey timed out")
		return Key{}, nil
	}
}

// TestReadKeyControlBytes is a regression test: plain control bytes used to be
// reported as KeyRune because the fast path never called decodeByte, which made
// Enter (and Tab, Backspace, Ctrl+C) unusable.
func TestReadKeyControlBytes(t *testing.T) {
	term, w := newPipeTerminal(t)
	cases := []struct {
		name string
		in   string
		want KeyType
	}{
		{"carriage return", "\r", KeyEnter},
		{"line feed", "\n", KeyEnter},
		{"tab", "\t", KeyTab},
		{"del", "\x7f", KeyBackspace},
		{"backspace", "\x08", KeyBackspace},
		{"ctrl+c", "\x03", KeyCtrlC},
		{"ctrl+d", "\x04", KeyCtrlD},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			writeTo(t, w, tc.in)
			key, err := readKeyWithin(t, term)
			if err != nil {
				t.Fatalf("ReadKey(%q) error: %v", tc.in, err)
			}
			if key.Type != tc.want {
				t.Fatalf("ReadKey(%q) = %v, want type %d", tc.in, key, tc.want)
			}
		})
	}
}

func TestReadKeyRunes(t *testing.T) {
	term, w := newPipeTerminal(t)
	for _, r := range []rune{'a', 'Z', 'é', '€'} {
		writeTo(t, w, string(r))
		key, err := readKeyWithin(t, term)
		if err != nil {
			t.Fatalf("ReadKey(%q) error: %v", string(r), err)
		}
		if key.Type != KeyRune || key.Rune != r {
			t.Fatalf("ReadKey(%q) = %v, want rune %q", string(r), key, r)
		}
	}
}

func TestReadKeyEscapeSequences(t *testing.T) {
	term, w := newPipeTerminal(t)
	cases := []struct {
		name string
		in   string
		want KeyType
	}{
		{"up", "\x1b[A", KeyUp},
		{"down", "\x1b[B", KeyDown},
		{"right", "\x1b[C", KeyRight},
		{"left", "\x1b[D", KeyLeft},
		{"home", "\x1b[H", KeyHome},
		{"end", "\x1b[F", KeyEnd},
		{"delete", "\x1b[3~", KeyDelete},
		{"insert", "\x1b[2~", KeyInsert},
		{"page up", "\x1b[5~", KeyPageUp},
		{"page down", "\x1b[6~", KeyPageDown},
		{"ss3 up", "\x1bOA", KeyUp},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			writeTo(t, w, tc.in)
			key, err := readKeyWithin(t, term)
			if err != nil {
				t.Fatalf("ReadKey(%q) error: %v", tc.in, err)
			}
			if key.Type != tc.want {
				t.Fatalf("ReadKey(%q) = %v, want type %d", tc.in, key, tc.want)
			}
		})
	}
}

// TestReadKeyLoneEsc is a regression test: ESC used to rely on
// SetReadDeadline, which terminals do not support, so a lone ESC either
// blocked or merged with the next key into KeyUnknown.
func TestReadKeyLoneEsc(t *testing.T) {
	term, w := newPipeTerminal(t)
	writeTo(t, w, "\x1b")
	key, err := readKeyWithin(t, term)
	if err != nil {
		t.Fatalf("ReadKey error: %v", err)
	}
	if key.Type != KeyEsc {
		t.Fatalf("ReadKey(ESC) = %v, want esc", key)
	}
}

func TestReadKeyEscFollowedByKey(t *testing.T) {
	term, w := newPipeTerminal(t)

	// ESC and Enter arriving in the same chunk.
	writeTo(t, w, "\x1b\r")
	key, err := readKeyWithin(t, term)
	if err != nil {
		t.Fatalf("ReadKey error: %v", err)
	}
	if key.Type != KeyEsc {
		t.Fatalf("first key = %v, want esc", key)
	}
	key, err = readKeyWithin(t, term)
	if err != nil {
		t.Fatalf("ReadKey error: %v", err)
	}
	if key.Type != KeyEnter {
		t.Fatalf("second key = %v, want enter", key)
	}

	// ESC and a rune arriving in separate chunks.
	writeTo(t, w, "\x1b")
	time.Sleep(10 * time.Millisecond)
	writeTo(t, w, "x")
	key, err = readKeyWithin(t, term)
	if err != nil {
		t.Fatalf("ReadKey error: %v", err)
	}
	if key.Type != KeyEsc {
		t.Fatalf("first key = %v, want esc", key)
	}
	key, err = readKeyWithin(t, term)
	if err != nil {
		t.Fatalf("ReadKey error: %v", err)
	}
	if key.Type != KeyRune || key.Rune != 'x' {
		t.Fatalf("second key = %v, want rune 'x'", key)
	}
}

// TestReadKeyMultipleInOneChunk verifies that buffered bytes are handed out
// one keypress at a time, in order.
func TestReadKeyMultipleInOneChunk(t *testing.T) {
	term, w := newPipeTerminal(t)
	writeTo(t, w, "ab\r")
	want := []KeyType{KeyRune, KeyRune, KeyEnter}
	for i, wt := range want {
		key, err := readKeyWithin(t, term)
		if err != nil {
			t.Fatalf("key %d error: %v", i, err)
		}
		if key.Type != wt {
			t.Fatalf("key %d = %v, want type %d", i, key, wt)
		}
	}
}

// TestReadKeyEscSequenceSplitAcrossReads is a regression test: when a chunk
// boundary fell between the ESC byte and the rest of an arrow key sequence,
// the lone ESC was reported as KeyEsc and cancelled the app instead of
// waiting for the rest of the sequence.
func TestReadKeyEscSequenceSplitAcrossReads(t *testing.T) {
	term, w := newPipeTerminal(t)

	// 21 complete arrow-downs plus a dangling ESC: exactly 64 bytes, the
	// internal read chunk size.
	writeTo(t, w, strings.Repeat("\x1b[B", 21)+"\x1b")
	time.Sleep(10 * time.Millisecond)
	writeTo(t, w, "[B")

	for i := 0; i < 22; i++ {
		key, err := readKeyWithin(t, term)
		if err != nil {
			t.Fatalf("key %d error: %v", i, err)
		}
		if key.Type != KeyDown {
			t.Fatalf("key %d = %v, want down", i, key)
		}
	}
}

func TestReadKeyEOF(t *testing.T) {
	term, w := newPipeTerminal(t)
	w.Close()
	if _, err := readKeyWithin(t, term); err != io.EOF {
		t.Fatalf("err = %v, want io.EOF", err)
	}
}
