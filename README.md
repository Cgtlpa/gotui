# gotui

[![Go Reference](https://pkg.go.dev/badge/github.com/Cgtlpa/gotui.svg)](https://pkg.go.dev/github.com/Cgtlpa/gotui)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Cgtlpa/gotui)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/Cgtlpa/gotui/blob/master/LICENSE)

A Go TUI (Terminal User Interface) framework with the **customizability of Bubble Tea** (charm.sh) and the **simplicity of Huh** (charm.sh).

Built for beginners, powerful enough for experts. Zero external dependencies — pure Go standard library.

```
Module path: github.com/Cgtlpa/gotui
Install:     go get github.com/Cgtlpa/gotui
Go version:  1.16+
```

---

## 1. What is gotui?

gotui lets you build terminal interfaces in minutes. Start with one-liners, then graduate to fully custom components when you need more control.

Two APIs in one framework:

| Style | Best for | Example |
|-------|----------|---------|
| **Simple App API** | Setup wizards, forms, quick tools | `app := gotui.New(); app.Text(...); app.Run()` |
| **Custom Model API** | Games, dashboards, async UI | `gotui.NewProgram(myModel).Run()` |

---

## 2. Installation

```bash
mkdir mytui && cd mytui
go mod init mytui
go get github.com/Cgtlpa/gotui
```

Import it:

```go
import "github.com/Cgtlpa/gotui"
```

---

## 3. Your First App (10 seconds)

```go
package main

import (
	"fmt"
	"log"

	"github.com/Cgtlpa/gotui"
)

func main() {
	app := gotui.New()
	app.Title("Welcome")
	app.Text("Hello, world!")
	app.Select("option", "Pick one", "A", "B", "C")

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("You chose:", app.Value("option"))
}
```

Run it:

```bash
go run main.go
```

---

## 4. Complete Tutorial: Building a Setup Wizard

Let's build a real "Silen Setup" wizard step by step, explaining every feature along the way.

### Step 1: Create the app and set the title

```go
app := gotui.New()
app.Title("Silen Setup")
```

`Title()` sets the heading shown above every screen. It renders in bold with your theme's primary color.

### Step 2: Choose your layout

```go
app.Center(true)   // center all text horizontally
app.Width(50)      // pad each line to 50 characters
```

- `Center(true)` — centers each line within `Width`. Looks great on setup wizards.
- `Width(n)` — pads lines to n characters (0 = auto).
- `Height(n)` — pads the view to n lines (0 = auto).
- `TextScale(n)` — repeats each character n× for bigger text (1 = normal, 2 = double).

### Step 3: Add screens (steps)

Each call adds a screen. `Run()` shows them one at a time.

```go
// A welcome screen (press any key to continue)
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/Cgtlpa/gotui"
)

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	clearScreen()
	app := gotui.New()
	app.Title("Welcome")
	app.Text("Hello, world!")
	app.Select("option", "Pick one", "A", "B", "C")

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("You chose:", app.Value("option"))
}

```

### Step 4: Run and retrieve results

```go
if err := app.Run(); err != nil {
    if errors.Is(err, gotui.ErrCancelled) {
        fmt.Println("Setup cancelled by user.")
        return
    }
    log.Fatal(err)
}

fmt.Println("\n═══ Setup Complete ═══")
fmt.Printf("  WM:     %s\n", app.Value("wm"))
fmt.Printf("  Name:   %s\n", app.Value("name"))
fmt.Printf("  Email:  %s\n", app.Value("email"))
fmt.Printf("  Bio:    %s\n", app.Value("bio"))
```

### Full example

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"gotui"
)

func main() {
	app := gotui.New()
	app.Title("Silen Setup")
	app.Center(true)
	app.Width(50)

	app.Text("Welcome to Silen Setup!\n\nConfigure your system below.")
	app.Select("wm", "Choose a window manager", "InstantWM", "Niri", "DWM")
	app.Input("name", "Your name: ", "e.g. Ada Lovelace")
	app.Confirm("confirm", "Continue?")
	app.Form("email", "Email", "bio", "Bio")

	if err := app.Run(); err != nil {
		if errors.Is(err, gotui.ErrCancelled) {
			fmt.Println("Cancelled.")
			return
		}
		log.Fatal(err)
	}

	fmt.Println("\nSetup complete!")
	fmt.Printf("  WM:   %s\n", app.Value("wm"))
	fmt.Printf("  Name: %s\n", app.Value("name"))
	fmt.Printf("  Email: %s\n", app.Value("email"))
	fmt.Printf("  Bio:  %s\n", app.Value("bio"))
}
```

---

## 5. App Method Reference

### `gotui.New()` — Create a new app

Returns a configured `*App` with the default theme.

### `app.Title(text)` — Set the heading

Shown above every screen. Renders in bold primary color.

### `app.Center(bool)` — Center text

`true` centers each line within `Width`. `false` left-aligns.

### `app.Width(n)` — Set line width

Pads each line to n characters (0 = auto, no padding).

### `app.Height(n)` — Set view height

Pads the view to n lines (0 = auto, no padding).

### `app.TextScale(n)` — Scale text size

Repeats each character n× horizontally. `1` = normal, `2` = double, etc. This is a visual effect (terminals can't change font size), but it's great for titles.

### `app.Theme(th)` — Set color palette

Use `gotui.DefaultTheme()`, `gotui.DarkTheme()`, `gotui.NeonTheme()`, or build your own.

### `app.Text(text)` — Add a read-only screen

Displays text. User presses any key to continue. No result stored.

### `app.Select(key, title, options...)` — Single-selection list

User picks one option with ↑/↓ (or j/k), Enter confirms. Result stored under `key`.

### `app.Input(key, prompt, placeholder)` — Single-line text input

User types text. Result stored under `key`. Max 64 characters by default (use standalone `WithCharLimit` to change).

### `app.Password(key, prompt)` — Masked input

Same as Input, but types are hidden with `*`. Result stored under `key`.

### `app.Confirm(key, title)` — Yes/no prompt

User toggles with ←/→, Enter confirms. Result stored under `key` (true = Yes, false = No).

### `app.Form(key, label, ...)` — Multi-field form

Fields are alternating key,label pairs. User tabs between fields, Enter submits. All results stored under their keys.

### `app.Custom(model)` — Add a fully custom Model

For when built-in components aren't enough. See section 11.

### `app.Value(key)` — Retrieve a result

Returns the stored result for a key after `Run()` completes. Returns `nil` if the key wasn't set.

### `app.Run() error` — Execute all steps

Shows each screen in order. Returns `nil` on success, or `gotui.ErrCancelled` if the user pressed Esc or Ctrl+C.

---

## 6. Standalone Components

Every component can also be used independently, returning its value directly:

```go
name, err := gotui.NewInput().
    WithPrompt("Name: ").
    WithPlaceholder("e.g. Ada").
    WithCharLimit(32).
    Run()
if err != nil {
    if errors.Is(err, gotui.ErrCancelled) {
        fmt.Println("Cancelled")
    }
    log.Fatal(err)
}
fmt.Println("Hello,", name)
```

| Component | Constructor | `Run()` returns |
|-----------|-------------|-----------------|
| `NewInput()` | `*Input` | `(string, error)` |
| `NewPassword()` | `*Password` | `(string, error)` |
| `NewSelect()` | `*Select` | `(string, error)` |
| `NewConfirm()` | `*Confirm` | `(bool, error)` |
| `NewForm()` | `*Form` | `error` (use `Get(key)`) |
| `NewText()` | `*Text` | `error` |
| `NewSpinner()` | `*Spinner` | `(Msg, error)` |

### Standalone component configuration

```go
gotui.NewSelect().
    WithTitle("Pick a color").
    WithOptions("Red", "Green", "Blue").
    WithCursor(1).
    WithTheme(gotui.DarkTheme()).
    Run()

gotui.NewInput().
    WithPrompt("Hostname: ").
    WithPlaceholder("silen-pc").
    WithCharLimit(32).
    Run()

gotui.NewConfirm().
    WithTitle("Delete?").
    DefaultYes().
    Run()

gotui.NewForm().
    WithTitle("Sign up").
    WithInput("name", "Full name").
    WithInput("email", "Email").
    WithPassword("pass", "Password")
```

---

## 7. Results & Error Handling

Every `Run()` returns an `error`. Always check it:

```go
if err := app.Run(); err != nil {
    if errors.Is(err, gotui.ErrCancelled) {
        // User pressed Esc or Ctrl+C
        fmt.Println("Cancelled")
        return
    }
    // Terminal failure, etc.
    log.Fatal(err)
}
```

Retrieve results:

```go
wm := app.Value("wm")           // "InstantWM"
name := app.Value("name")       // "Ada Lovelace"
confirmed := app.Value("confirm") // true/false
```

For standalone forms:

```go
form := gotui.NewForm().
    WithInput("name", "Name").
    WithInput("email", "Email")
form.Run()
fmt.Println(form.Get("name"))
```

---

## 8. Cancellation

Behavior is consistent across every component:

| Key | Action |
|-----|--------|
| `Enter` | Confirm current choice / submit |
| `Esc` | Cancel current step → returns `ErrCancelled` |
| `Ctrl+C` | Cancel whole app → returns `ErrCancelled` |
| `Backspace` | Delete character before cursor |

All keys work the same in Input, Password, Select, Confirm, and Form.

---

## 9. Layout System

Control how each screen is positioned and scaled:

```go
app := gotui.New()
app.Width(60)
app.Height(15)
app.Center(true)
app.TextScale(2)
app.Text("Big centered title")
app.Run()
```

| Field | What it does |
|-------|--------------|
| `Width(n)` | Pad each line to n characters |
| `Height(n)` | Pad view to n lines |
| `Center(bool)` | Center each line within Width |
| `TextScale(n)` | Repeat each character n× horizontally |

Long text is not clipped — it renders as-is. If width exceeds the terminal, the terminal wraps naturally.

---

## 10. Colors & Themes

Every component supports `WithTheme(t)`.

### Built-in themes

```go
gotui.DefaultTheme()  // balanced, works everywhere
gotui.DarkTheme()     // optimized for dark terminals
gotui.NeonTheme()     // high-contrast magenta/cyan
```

### Custom theme

```go
t := gotui.DefaultTheme()
t.Primary = 200   // pink
t.Danger = 198    // deep red
t.Success = 46    // bright green

app := gotui.New()
app.Theme(t)
```

### Color constants

`ColorBlack`, `ColorRed`, `ColorGreen`, `ColorYellow`, `ColorBlue`, `ColorMagenta`, `ColorCyan`, `ColorWhite`, `ColorGray`.

### Theme methods

| Method | Returns |
|--------|---------|
| `t.FgColor(c)` | ANSI foreground sequence for 256-color `c` |
| `t.BgColor(c)` | ANSI background sequence for 256-color `c` |
| `t.Bold()` | Bold sequence |
| `t.Reverse()` | Reverse video (used for cursors) |
| `t.Reset()` | Reset sequence |

There's also `gotui.Style{}` — a beginner-friendly wrapper with `FgColor`, `BgColor`, `Bold`, `Reverse`, `Reset`, `Rgb(r,g,b)`.

---

## 11. Custom Models (Bubble Tea-style)

When built-in components aren't enough, implement the `Model` interface:

```go
type Model interface {
    Init() Cmd
    Update(msg Msg) (Model, Cmd)
    View() string
}
```

### Example: a counter

```go
type counter struct {
    count int
}

func (m counter) Init() gotui.Cmd { return nil }

func (m counter) Update(msg gotui.Msg) (gotui.Model, gotui.Cmd) {
    km, ok := msg.(gotui.KeyMsg)
    if !ok {
        return m, nil
    }

    switch km.Key.Type {
    case gotui.KeyRune:
        switch km.Key.Rune {
        case '+':
            m.count++
        case '-':
            m.count--
        case 'q':
            return m, gotui.Quit
        }
    case gotui.KeyEnter:
        return m, gotui.Quit
    }

    return m, nil
}

func (m counter) View() string {
    th := gotui.DarkTheme()
    color := th.FgColor(th.Foreground)
    if m.count > 0 {
        color = th.FgColor(gotui.ColorGreen)
    } else if m.count < 0 {
        color = th.FgColor(gotui.ColorRed)
    }
    return fmt.Sprintf(
        "%s%sCount%s: %s%d%s\n\nPress + / - / q to quit.\n",
        th.Bold(), th.FgColor(th.Primary), th.Reset(),
        color, m.count, th.Reset(),
    )
}

// Use it:
app := gotui.New()
app.Custom(counter{count: 0})
app.Run()
```

### Commands (`Cmd`)

A `Cmd` is `func() Msg`. Use them for:

- `gotui.Quit` — exit the program
- `gotui.Batch(cmd1, cmd2)` — run several commands
- Async work (goroutines, timers, API calls) — send messages back with `program.Send(msg)`

### Sending messages from outside

```go
program.Send(gotui.KeyMsg{Key: gotui.Key{Type: gotui.KeyRune, Rune: 'q'}})
```

---

## 12. Terminal Restoration

The terminal is **always** restored to its original state when `Run()` returns — even on error or cancellation. This is handled internally via `defer`. Your shell will work normally after exit.

Test it: press Ctrl+C during a component, then type in your shell — it should work normally.

---

## 13. Examples

All under `examples/`, run with `go run ./examples/<name>`:

| Example | What it shows |
|---------|---------------|
| `simple_input` | Simple App API + `app.Value()` result retrieval |
| `form` | Multi-field form with `Get(key)` |
| `custom` | Custom Model via `app.Custom()` |
| `demo` | Combined showcase with themes and results |
| `layout` | Center, Width, Height, TextScale |

Also included: `go run ./cmd/preview` — renders every component to stdout with ANSI colors (no live terminal needed).

---

## 14. Keyboard Reference

| Key | Input/Password | Select | Confirm | Form | Text |
|-----|---------------|--------|---------|------|------|
| type | insert char | — | — | insert char | — |
| `←`/`→` | move cursor | — | toggle Yes/No | move cursor | — |
| `↑`/`↓` / `j`/`k` | — | move selection | — | — | — |
| `Tab` | — | — | — | next field | — |
| `Enter` | submit | confirm | confirm | submit (last field) | continue |
| `Esc` | cancel | cancel | cancel | cancel | cancel |
| `Ctrl+C` | cancel | cancel | cancel | cancel | cancel |
| `Backspace` | delete | — | — | delete | — |

---

## 15. Known Limitations

- **Windows:** Current raw-mode implementation targets Unix-like systems (Linux/macOS) via `syscall` termios. Windows support is a contribution opportunity.
- **Terminal resize:** Not handled mid-session. Resize before starting.
- **FontSize:** `TextScale(n)` repeats characters horizontally — it's a visual effect, not real font scaling (terminals can't change font size).
- **No wrapping:** Long text is not wrapped or clipped — it renders as-is.

---

## 16. License

MIT — same spirit as the charm.sh tools that inspired it.

Happy terminal building! 🍵
