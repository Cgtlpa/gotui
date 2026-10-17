package gotui

import "errors"

// ErrCancelled is returned when a component or the whole app is cancelled
// (e.g. the user presses Esc or Ctrl+C).
var ErrCancelled = errors.New("gotui: cancelled")
