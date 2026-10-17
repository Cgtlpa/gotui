package gotui

import (
	"fmt"
	"strings"
)

// Theme holds the color palette used by components.
type Theme struct {
	Primary    int
	Secondary  int
	Success    int
	Warning    int
	Danger     int
	Muted      int
	Background int
	Foreground int
}

// FgColor returns the ANSI foreground sequence for a 256-color code.
func (t Theme) FgColor(color int) string { return fmt.Sprintf("\x1b[38;5;%dm", color) }

// BgColor returns the ANSI background sequence for a 256-color code.
func (t Theme) BgColor(color int) string { return fmt.Sprintf("\x1b[48;5;%dm", color) }

// Reset returns the ANSI reset sequence.
func (t Theme) Reset() string { return "\x1b[0m" }

// Bold returns the ANSI bold sequence.
func (t Theme) Bold() string { return "\x1b[1m" }

// Reverse returns the ANSI reverse-video sequence (used for cursors).
func (t Theme) Reverse() string { return "\x1b[7m" }

// DefaultTheme is a balanced palette.
func DefaultTheme() Theme {
	return Theme{
		Primary:    ColorCyan,
		Secondary:  ColorBlue,
		Success:    ColorGreen,
		Warning:    ColorYellow,
		Danger:     ColorRed,
		Muted:      ColorGray,
		Background: ColorBlack,
		Foreground: ColorWhite,
	}
}

// DarkTheme works well on dark terminals.
func DarkTheme() Theme {
	return Theme{
		Primary:    ColorCyan,
		Secondary:  ColorMagenta,
		Success:    ColorGreen,
		Warning:    ColorYellow,
		Danger:     ColorRed,
		Muted:      ColorGray,
		Background: ColorBlack,
		Foreground: ColorWhite,
	}
}

// NeonTheme is high-contrast magenta/cyan.
func NeonTheme() Theme {
	return Theme{
		Primary:    ColorMagenta,
		Secondary:  ColorCyan,
		Success:    ColorGreen,
		Warning:    ColorYellow,
		Danger:     ColorRed,
		Muted:      ColorGray,
		Background: ColorBlack,
		Foreground: ColorWhite,
	}
}

const (
	ColorBlack   = 0
	ColorRed     = 1
	ColorGreen   = 2
	ColorYellow  = 3
	ColorBlue    = 4
	ColorMagenta = 5
	ColorCyan    = 6
	ColorWhite   = 7
	ColorGray    = 8
)

// Style is a beginner-friendly wrapper around Theme's ANSI helpers.
type Style struct{}

func (s Style) FgColor(color int) string { return fmt.Sprintf("\x1b[38;5;%dm", color) }
func (s Style) BgColor(color int) string { return fmt.Sprintf("\x1b[48;5;%dm", color) }
func (s Style) Rgb(r, g, b uint8) string { return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b) }
func (s Style) Bold() string             { return "\x1b[1m" }
func (s Style) Reverse() string          { return "\x1b[7m" }
func (s Style) Reset() string            { return "\x1b[0m" }

// Layout controls how a component's view is positioned and scaled.
type Layout struct {
	// Width pads each line to this many characters (0 = no padding).
	Width int
	// Height pads the view to this many lines (0 = no padding).
	Height int
	// Center centers each line within Width.
	Center bool
	// TextScale repeats each character horizontally to simulate larger text.
	TextScale int
}

// Apply pads and scales a view string according to the layout.
func (l Layout) Apply(view string) string {
	lines := splitLines(view)

	// Scale horizontally (TextScale). Only visible characters are repeated;
	// ANSI escape sequences are preserved intact.
	if l.TextScale > 1 {
		scaled := make([]string, len(lines))
		for i, line := range lines {
			scaled[i] = scaleLine(line, l.TextScale)
		}
		lines = scaled
	}

	// Pad each line to Width, centering if requested.
	if l.Width > 0 {
		for i, line := range lines {
			lines[i] = padLine(line, l.Width, l.Center)
		}
	}

	// Pad vertically to Height.
	if l.Height > 0 {
		for len(lines) < l.Height {
			lines = append(lines, strings.Repeat(" ", maxLineWidth(lines)))
		}
	}

	return strings.Join(lines, "\n")
}

// scaleLine repeats each visible character n times while leaving ANSI escape
// sequences untouched.
func scaleLine(line string, n int) string {
	var b strings.Builder
	i := 0
	for i < len(line) {
		if line[i] == '\x1b' && i+1 < len(line) && line[i+1] == '[' {
			// Copy the entire escape sequence as-is.
			b.WriteByte(line[i])
			b.WriteByte(line[i+1])
			i += 2
			for i < len(line) && line[i] != '[' && (line[i] < '@' || line[i] > '~') {
				b.WriteByte(line[i])
				i++
			}
			if i < len(line) {
				b.WriteByte(line[i])
				i++
			}
		} else {
			for j := 0; j < n; j++ {
				b.WriteByte(line[i])
			}
			i++
		}
	}
	return b.String()
}

// splitLines splits a string into lines, preserving empty lines.
func splitLines(s string) []string {
	if s == "" {
		return []string{""}
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

// visibleLen returns the number of visible characters in a string,
// stripping ANSI escape sequences.
func visibleLen(s string) int {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			i += 2
			for i < len(s) && s[i] != '[' && (s[i] < '@' || s[i] > '~') {
				i++
			}
			if i < len(s) {
				i++ // skip final byte
			}
		} else {
			b.WriteByte(s[i])
			i++
		}
	}
	return b.Len()
}

// padLine pads a line to width, centering if requested.
// Uses visible length so ANSI codes don't throw off alignment.
func padLine(line string, width int, center bool) string {
	if visibleLen(line) >= width {
		return line
	}
	pad := width - visibleLen(line)
	if center {
		left := pad / 2
		right := pad - left
		return strings.Repeat(" ", left) + line + strings.Repeat(" ", right)
	}
	return line + strings.Repeat(" ", pad)
}

// maxLineWidth returns the visible width of the widest line.
func maxLineWidth(lines []string) int {
	w := 0
	for _, line := range lines {
		if v := visibleLen(line); v > w {
			w = v
		}
	}
	return w
}
