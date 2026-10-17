# gotui — Complete Tutorial & Guide

A Go TUI (Terminal User Interface) framework with the **customizability of Bubble Tea** (charm.sh) and the **simplicity of Huh** (charm.sh).

Built for beginners, powerful enough for experts. Zero external dependencies — pure Go standard library.

```
Module path: github.com/Cgtlpa/gotui
Install:     go get github.com/Cgtlpa/gotui
Go version:  1.16+
```

---

## Table of Contents

1. [What is gotui?](#what-is-gotui)
2. [Installation](#installation)
3. [Your First App (60 seconds)](#your-first-app)
4. [The App API in Detail](#the-app-api-in-detail)
5. [Standalone Components](#standalone-components)
6. [Results & Error Handling](#results--error-handling)
7. [Cancellation](#cancellation)
8. [Layout System](#layout-system)
9. [Colors & Themes](#colors--themes)
10. [Custom Models (Bubble Tea-style)](#custom-models-bubble-tea-style)
11. [Terminal Restoration](#terminal-restoration)
12. [Keyboard Reference](#keyboard-reference)
13. [Complete Walkthrough: Setup Wizard](#complete-walkthrough-setup-wizard)
14. [Examples](#examples)
15. [Known Limitations](#known-limitations)
16. [FAQ](#faq)

---

## What is gotui?

gotui gives you **two APIs in one framework**:

| Style | Best for | Example |
|-------|----------|---------|
| **Simple App API** | Setup wizards, forms, quick CLI tools | `app := gotui.New(); app.Text(...); app.Run()` |
| **Custom Model API** | Games, dashboards, async UI, full control | `gotui.NewProgram(myModel).Run()` |

**Simple App API** — Chainable methods, sequential screens, automatic result storage. Perfect for the 80% case (forms, inputs, confirms, selects).

**Custom Model API** — Implement `Init()`/`Update()`/`View()` yourself. Total control over rendering, input, and asynchronous commands. Perfect for the 20% case where you need custom behavior.

---

## Installation

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

## Your First App (60 seconds)

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

**What happens:**

1. `gotui.New()` creates a new App with the default theme.
2. `app.Title("Welcome")` sets the heading shown above every screen.
3. `app.Text("Hello, world!")` adds a read-only screen (press any key to continue).
4. `app.Select("option", "Pick one", "A", "B", "C")` adds a selection list. The user picks one. The result is stored under the key `"option"`.
5. `app.Run()` shows each screen in order, one after another.
6. `app.Value("option")` retrieves the user's choice after `Run()` completes.

---

## The App API in Detail

### `gotui.New()` — Create a new App

Returns a configured `*App` with the default theme and an empty steps list.

```go
app := gotui.New()
```

### `app.Title(text)` — Set the heading

Shown above every screen. Renders in bold with your theme's primary color.

```go
app.Title("Silen Setup")
```

Chainable: returns `*App`.

### `app.Center(bool)` — Center text horizontally

`true` centers each line within `Width`. `false` left-aligns (default).

```go
app.Center(true)
```

Chainable: returns `*App`.

### `app.Width(n)` — Set line width

Pads each line to n characters. `0` = auto (no padding).

```go
app.Width(50)
```

Chainable: returns `*App`.

### `app.Height(n)` — Set view height

Pads the view to n lines. `0` = auto (no padding).

```go
app.Height(15)
```

Chainable: returns `*App`.

### `app.TextScale(n)` — Scale text size

Repeats each character n× horizontally. `1` = normal, `2` = double, etc.

This is a visual effect (terminals can't change actual font size), but it's great for titles and headers.

```go
app.TextScale(2)
```

Chainable: returns `*App`.

### `app.Theme(th)` — Set color palette

Use a built-in theme or build your own.

```go
app.Theme(gotui.DarkTheme())
```

Chainable: returns `*App`.

### `app.Value(key)` — Retrieve a result

Returns the stored result for a key after `Run()` completes. Returns `nil` if the key wasn't set.

```go
wm := app.Value("wm")        // "InstantWM"
name := app.Value("name")    // "Ada Lovelace"
confirmed := app.Value("confirm") // true/false
```

### `app.Run() error` — Execute all steps

Shows each screen in order, one after another. Blocks until all steps complete or the user cancels.

Returns:
- `nil` on success
- `gotui.ErrCancelled` if the user pressed Esc or Ctrl+C
- Other errors if the terminal could not be started

```go
if err := app.Run(); err != nil {
    if errors.Is(err, gotui.ErrCancelled) {
        fmt.Println("Cancelled.")
        return
    }
    log.Fatal(err)
}
```

---

## Adding Screens (Steps)

Each method adds a screen. `Run()` executes them sequentially.

### `app.Text(text)` — Read-only text screen

Displays text. User presses any key to continue. No result stored.

```go
app.Text("Welcome to Silen Setup!\n\nThis will configure your window manager.")
```

Supports multiple lines via `\n`.

### `app.Select(key, title, options...)` — Single-selection list

User picks one option with `↑`/`↓` (or `j`/`k`), `Enter` confirms. Result stored under `key`.

```go
app.Select("wm", "Choose a window manager", "InstantWM", "Niri", "DWM")
```

The selected option is stored as a string under `"wm"`.

### `app.Input(key, prompt, placeholder)` — Single-line text input

User types text. `Enter` submits. Result stored under `key`.

```go
app.Input("name", "Your name: ", "e.g. Ada Lovelace")
```

Default char limit: 64 characters. Use standalone `WithCharLimit(n)` to change (see Standalone Components below).

### `app.Password(key, prompt)` — Masked text input

Same as Input, but typed characters are hidden with `*`. Result stored under `key`.

```go
app.Password("pass", "Password: ")
```

### `app.Confirm(key, title)` — Yes/no prompt

User toggles with `←`/`→`, `Enter` confirms. Result stored under `key` (`true` = Yes, `false` = No).

```go
app.Confirm("confirm", "Continue with these settings?")
```

### `app.Form(key, label, ...)` — Multi-field form

Fields are given as alternating `key, label` pairs. User tabs between fields with `Tab`, `Enter` on the last field submits. All results stored under their keys.

```go
app.Form("email", "Email", "bio", "Bio")
```

### `app.Custom(model)` — Fully custom Model

For when built-in components aren't enough. Implement the `Model` interface yourself (see Custom Models section below).

```go
app.Custom(myModel{})
```

---

## Standalone Components

Every component can also be used independently, returning its value directly without the App wrapper.

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

### Component Reference

| Component | Constructor | `Run()` returns |
|-----------|-------------|-----------------|
| `NewInput()` | `*Input` | `(string, error)` |
| `NewPassword()` | `*Password` | `(string, error)` |
| `NewSelect()` | `*Select` | `(string, error)` |
| `NewConfirm()` | `*Confirm` | `(bool, error)` |
| `NewForm()` | `*Form` | `error` (use `Get(key)`) |
| `NewText()` | `*Text` | `error` |
| `NewSpinner()` | `*Spinner` | `(Msg, error)` |

### Standalone Configuration Methods

**Input:**
```go
gotui.NewInput().
    WithPrompt("Name: ").
    WithPlaceholder("e.g. Ada").
    WithValue("Pre-filled").
    WithCharLimit(32).
    WithTheme(gotui.DarkTheme()).
    WithLayout(gotui.Layout{Width: 40, Center: true}).
    Run()
```

**Password:**
```go
gotui.NewPassword().
    WithPrompt("Password: ").
    WithTheme(gotui.DefaultTheme()).
    Run()
```

**Select:**
```go
gotui.NewSelect().
    WithTitle("Pick a color").
    WithOptions("Red", "Green", "Blue").
    WithCursor(1).
    WithTheme(gotui.NeonTheme()).
    WithLayout(gotui.Layout{Width: 30}).
    Run()
```

**Confirm:**
```go
gotui.NewConfirm().
    WithTitle("Delete?").
    DefaultYes().
    WithTheme(gotui.DefaultTheme()).
    Run()
```

**Form:**
```go
form := gotui.NewForm().
    WithTitle("Sign up").
    WithInput("name", "Full name").
    WithInput("email", "Email").
    WithPassword("pass", "Password").
    WithTheme(gotui.DefaultTheme())

err := form.Run()
if err != nil {
    log.Fatal(err)
}
fmt.Println(form.Get("name"))
fmt.Println(form.Get("pass"))
```

**Text:**
```go
gotui.NewText().
    WithTitle("Welcome").
    WithBody("Hello, world!\nThis is gotui.").
    WithTheme(gotui.DarkTheme()).
    Run()
```

**Spinner:**
```go
spinner := gotui.NewSpinner().
    WithText("Loading...").
    WithCmd(func() gotui.Msg {
        // do work...
        return gotui.Quit
    })

msg, err := spinner.Run()
```

---

## Results & Error Handling

Every `Run()` returns an `error`. Always check it.

### App-level errors

```go
if err := app.Run(); err != nil {
    if errors.Is(err, gotui.ErrCancelled) {
        // User pressed Esc or Ctrl+C
        fmt.Println("Setup cancelled by user.")
        return
    }
    // Terminal failure, etc.
    log.Fatal(err)
}
```

### Standalone component errors

```go
name, err := gotui.NewInput().WithPrompt("Name: ").Run()
if err != nil {
    if errors.Is(err, gotui.ErrCancelled) {
        fmt.Println("Input cancelled.")
    }
    log.Fatal(err)
}
fmt.Println(name)
```

### Retrieving results

After `app.Run()` succeeds:

```go
wm := app.Value("wm")           // string: "InstantWM"
name := app.Value("name")       // string: "Ada Lovelace"
confirmed := app.Value("confirm") // bool: true/false
```

For forms:

```go
form := gotui.NewForm().
    WithInput("name", "Name").
    WithInput("email", "Email")
form.Run()
fmt.Println(form.Get("name"))
fmt.Println(form.Get("email"))
```

---

## Cancellation

Behavior is consistent across **every** component:

| Key | Action |
|-----|--------|
| `Enter` | Confirm current choice / submit form |
| `Esc` | Cancel current step → returns `ErrCancelled` |
| `Ctrl+C` | Cancel whole app → returns `ErrCancelled` |
| `Backspace` | Delete character before cursor |
| `←`/`→` | Move cursor (input) / toggle Yes/No (confirm) |
| `↑`/`↓` / `j`/`k` | Move selection (select) |
| `Tab` | Next field (form) |

Example:

```go
if err := app.Run(); err != nil {
    if errors.Is(err, gotui.ErrCancelled) {
        fmt.Println("User cancelled the setup.")
        return
    }
    log.Fatal(err)
}
```

---

## Layout System

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
| `Width(n)` | Pad each line to n characters (0 = auto) |
| `Height(n)` | Pad view to n lines (0 = auto) |
| `Center(bool)` | Center each line within Width |
| `TextScale(n)` | Repeat each character n× horizontally |

Long text is **not clipped or wrapped** — it renders as-is. If width exceeds the terminal, the terminal wraps naturally.

### Example: Centered input

```go
app := gotui.New()
app.Title("Login")
app.Center(true)
app.Width(40)
app.Input("username", "Username: ", "e.g. ada")
app.Run()
```

### Example: Big title

```go
app := gotui.New()
app.Center(true)
app.Width(60)
app.TextScale(2)
app.Text("SILEN SETUP")
app.Run()
```

---

## Colors & Themes

Every component supports `WithTheme(t)`. The App supports `app.Theme(t)`.

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
t.Warning = 226   // bright yellow
t.Muted = 244     // light gray

app := gotui.New()
app.Theme(t)
```

### Color constants

```
ColorBlack   = 0
ColorRed     = 1
ColorGreen   = 2
ColorYellow  = 3
ColorBlue    = 4
ColorMagenta = 5
ColorCyan    = 6
ColorWhite   = 7
ColorGray    = 8
```

### Theme methods

| Method | Returns |
|--------|---------|
| `t.FgColor(c)` | ANSI foreground sequence for 256-color `c` |
| `t.BgColor(c)` | ANSI background sequence for 256-color `c` |
| `t.Bold()` | Bold sequence |
| `t.Reverse()` | Reverse video (used for cursors) |
| `t.Reset()` | Reset sequence |

### Style wrapper (beginner-friendly)

```go
style := gotui.Style{}
red := style.FgColor(gotui.ColorRed) + "error" + style.Reset()
greenBold := style.Bold() + style.FgColor(gotui.ColorGreen) + "ok" + style.Reset()
trueColor := style.Rgb(255, 100, 0) + "orange" + style.Reset()
```

---

## Custom Models (Bubble Tea-style)

When built-in components aren't enough, implement the `Model` interface:

```go
type Model interface {
    Init() Cmd
    Update(msg Msg) (Model, Cmd)
    View() string
}
```

### Example: A counter

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

- Returning `gotui.Quit` to exit
- Returning `gotui.Batch(cmd1, cmd2)` to run several commands
- Async work (goroutines, timers, API calls) — send messages back with `program.Send(msg)`

### Sending messages from outside

```go
program.Send(gotui.KeyMsg{Key: gotui.Key{Type: gotui.KeyRune, Rune: 'q'}})
```

---

## Terminal Restoration

The terminal is **always** restored to its original state when `Run()` returns — even on error or cancellation. This is handled internally via `defer`. Your shell will work normally after exit.

Test it: press Ctrl+C during a component, then type in your shell — it should work normally.

---

## Keyboard Reference

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

## Complete Walkthrough: Setup Wizard

Let's build a complete "Silen Setup" wizard that demonstrates every feature.

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/Cgtlpa/gotui"
)

func main() {
	app := gotui.New()
	app.Title("Silen Setup")
	app.Center(true)
	app.Width(50)
	app.Theme(gotui.NeonTheme())

	// Welcome screen
	app.Text("Welcome to Silen Setup!\n\n" +
		"This wizard will configure your window manager and user account.\n" +
		"Press any key to continue.")

	// Select a window manager (result stored under "wm")
	app.Select("wm", "Choose a window manager",
		"InstantWM",
		"Niri",
		"DWM",
	)

	// Enter your name (result stored under "name")
	app.Input("name", "Your name: ", "e.g. Ada Lovelace")

	// Confirm (result stored under "confirm")
	app.Confirm("confirm", "Continue with these settings?")

	// Multi-field form (results stored under "email" and "bio")
	app.Form("email", "Email", "bio", "Bio")

	// Run the wizard
	if err := app.Run(); err != nil {
		if errors.Is(err, gotui.ErrCancelled) {
			fmt.Println("\nSetup cancelled by user.")
			return
		}
		log.Fatal(err)
	}

	// Display results
	fmt.Println("\n═══ Setup Complete ═══")
	fmt.Printf("  Window Manager: %s\n", app.Value("wm"))
	fmt.Printf("  Name:           %s\n", app.Value("name"))
	fmt.Printf("  Email:          %s\n", app.Value("email"))
	fmt.Printf("  Bio:            %s\n", app.Value("bio"))
	fmt.Println("══════════════════════")
}
```

**What the user sees:**

1. A neon-colored welcome screen (press any key)
2. A selection list for window manager (↑/↓ to move, Enter to confirm)
3. A text input for their name (type, Enter to submit)
4. A yes/no confirm (←/→ to toggle, Enter to confirm)
5. A multi-field form with email and bio (Tab to move, Enter on last field to submit)
6. A summary printed to the terminal after the wizard exits

---

## Examples

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

## Known Limitations

- **Windows:** Current raw-mode implementation targets Unix-like systems (Linux/macOS) via `syscall` termios. Windows support is a contribution opportunity.
- **Terminal resize:** Not handled mid-session. Resize before starting.
- **TextScale:** `TextScale(n)` repeats characters horizontally — it's a visual effect, not real font scaling (terminals can't change font size).
- **No wrapping:** Long text is not wrapped or clipped — it renders as-is.
- **Module proxy cache:** After pushing a new commit, `go get ...@latest` may take minutes to an hour to pick it up. Use `go get ...@<commit-hash>` for immediate access to a specific version.

---

## FAQ

**Q: Do I need to install anything besides Go?**
A: No. gotui is pure standard library — zero external dependencies.

**Q: Can I use gotui components inside Bubble Tea?**
A: gotui implements the same `Model`/`Update`/`View` pattern as Bubble Tea. If you're already using `charmbracelet/bubbletea`, you can port a gotui `Model` directly — the signatures match. They are separate modules, so you'd copy the model logic rather than import both.

**Q: Can I mix simple and custom APIs?**
A: Yes. Every simple component is also a `Model`, so you can put `NewInput()` inside a custom `Program`. Likewise, a custom `Model` can be run via `gotui.Run()` if it also implements `Component.Result()`.

**Q: Does it support Windows?**
A: Not yet — the raw-mode implementation targets Unix-like systems (Linux/macOS). Windows support via `golang.org/x/term` is a contribution opportunity.

**Q: How do I get results from the App?**
A: Use `app.Value("key")` after `Run()` returns. The key is the first argument you passed to `Select`, `Input`, `Password`, or `Confirm`.

**Q: What happens if the user presses Ctrl+C?**
A: `Run()` returns `gotui.ErrCancelled`. Check with `errors.Is(err, gotui.ErrCancelled)`.

**Q: Is the terminal restored after exit?**
A: Yes, always — even on error or cancellation. This is handled via `defer` in `Run()`.

---

## License

MIT — same spirit as the charm.sh tools that inspired it.

Happy terminal building! 🍵