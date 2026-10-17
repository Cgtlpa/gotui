package gotui

// KeyType identifies the kind of key that was pressed.
type KeyType int

const (
	KeyUnknown KeyType = iota
	KeyRune
	KeyEnter
	KeyBackspace
	KeyTab
	KeyEsc
	KeyCtrlC
	KeyCtrlD
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyDelete
	KeyInsert
	KeyPageUp
	KeyPageDown
)

// Key describes a single keypress.
type Key struct {
	Type KeyType
	Rune rune
	Code string
}

func (k Key) String() string {
	switch k.Type {
	case KeyRune:
		return string(k.Rune)
	case KeyEnter:
		return "enter"
	case KeyBackspace:
		return "backspace"
	case KeyTab:
		return "tab"
	case KeyEsc:
		return "esc"
	case KeyCtrlC:
		return "ctrl+c"
	case KeyCtrlD:
		return "ctrl+d"
	case KeyUp:
		return "up"
	case KeyDown:
		return "down"
	case KeyLeft:
		return "left"
	case KeyRight:
		return "right"
	case KeyHome:
		return "home"
	case KeyEnd:
		return "end"
	case KeyDelete:
		return "delete"
	case KeyInsert:
		return "insert"
	case KeyPageUp:
		return "page up"
	case KeyPageDown:
		return "page down"
	default:
		return "unknown"
	}
}
