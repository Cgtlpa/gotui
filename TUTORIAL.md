# gotui Tutorial

A step-by-step guide to building terminal apps with **gotui**, from your first screen to fully custom components.

No prior TUI experience needed. If you can write a basic Go program, you can follow along.

---

## Table of Contents

1. [What is a TUI?](#1-what-is-a-tui)
2. [Installation](#2-installation)
3. [Your first app](#3-your-first-app)
4. [How gotui thinks: screens and keys](#4-how-gotui-thinks-screens-and-keys)
5. [Building a sign-up wizard, step by step](#5-building-a-sign-up-wizard-step-by-step)
6. [Handling cancellation and errors](#6-handling-cancellation-and-errors)
7. [Layout: width, height, centering, scaling](#7-layout-width-height-centering-scaling)
8. [Colors and themes](#8-colors-and-themes)
9. [Standalone components](#9-standalone-components)
10. [Custom models](#10-custom-models)
11. [Keyboard reference](#11-keyboard-reference)
12. [Method reference](#12-method-reference)
13. [Known limitations](#13-known-limitations)

---

## 1. What is a TUI?

A **TUI** (Terminal User Interface) is a program that runs inside your terminal but looks more like an app than a plain command. Instead of typing commands and reading text, you move through menus with the arrow keys, fill in forms, and answer prompts.

Think of the difference between:

- **A normal command-line program**: like talking to someone over a walkie-talkie. You say something, they answer, and so on.
- **A TUI**: like using a touchscreen kiosk at a restaurant. There are screens, buttons, and choices.

gotui helps you build the kiosk. You describe the screens, and gotui handles the drawing, the keyboard, and cleaning up afterwards.

gotui offers two ways to build things:

| Style | Best for | What it looks like |
|-------|----------|--------------------|
| **Simple App API** | Wizards, forms, quick tools | `app := gotui.New(); app.Text(...); app.Run()` |
| **Custom Model API** | Games, dashboards, live updating UIs | `gotui.NewProgram(myModel).Run()` |

Start with the Simple App API. You can move to custom models later if you need to.

---

## 2. Installation

You need Go 1.16 or newer. Check with:

```bash
go version
```

Create a new project folder and add gotui to it:

```bash
mkdir mytui
cd mytui
go mod init mytui
go get github.com/Cgtlpa/gotui
```

What each line does:

- `mkdir mytui` / `cd mytui`: makes a new folder and steps into it.
- `go mod init mytui`: tells Go "this folder is a project called mytui". This creates a `go.mod` file.
- `go get github.com/Cgtlpa/gotui`: downloads gotui and adds it to your project.

gotui has **no other dependencies**. It only uses Go's standard library, so nothing else gets installed.

In your code, import it like this:

```go
import "github.com/Cgtlpa/gotui"
```

---

## 3. Your first app

Create a file called `main.go`:

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

You will see a "Hello, world!" screen. Press any key and a list appears. Move with ↑/↓ and press Enter to choose. After you pick, the program prints your choice.

### What just happened, line by line

| Line | Meaning |
|------|---------|
| `app := gotui.New()` | Creates an empty app. Think of it as a blank slideshow. |
| `app.Title("Welcome")` | Sets a heading that appears at the top of every screen. |
| `app.Text("Hello, world!")` | Adds slide #1: a message. The user presses any key to move on. |
| `app.Select("option", "Pick one", "A", "B", "C")` | Adds slide #2: a list of choices. The answer will be saved under the name `"option"`. |
| `app.Run()` | Plays the slideshow, one slide at a time. It returns when the last slide is done. |
| `app.Value("option")` | Looks up the answer saved under `"option"`. |

**Important:** nothing appears on screen until `Run()` is called. The lines before it only *describe* the slides.

---

## 4. How gotui thinks: screens and keys

Two ideas explain almost everything in gotui.

### Idea 1: Screens (steps)

Every call like `app.Text(...)`, `app.Select(...)`, `app.Input(...)` adds one **screen** to the end of a list. When you call `Run()`, gotui shows them in the order you added them.

Think of it like a stack of index cards. You write the cards first, then flip through them.

### Idea 2: Keys (names for answers)

Some screens collect an answer (a choice, some typed text, a yes/no). To find that answer later, you give it a **key**, which is just a name you choose.

Think of keys like labeled drawers. When the user answers a question, gotui puts the answer in the drawer with that label. After `Run()` finishes, you open the drawer with `app.Value("label")`.

```go
app.Input("name", "Your name: ", "e.g. Ada")   // the answer goes in the drawer "name"
// ... later, after Run() ...
fmt.Println(app.Value("name"))                  // open the drawer "name"
```

Rules for keys:

- Each key should be unique. Using the same key twice means the later answer replaces the earlier one.
- If a key was never filled, `app.Value("key")` returns `nil`.
- `Text` screens don't collect anything, so they don't need a key.

---

## 5. Building a sign-up wizard, step by step

Let's build something realistic: a small sign-up wizard that asks for a few things and prints a summary. We'll add one piece at a time.

### Step 1: Create the app and a title

```go
app := gotui.New()
app.Title("Sign-up Wizard")
```

`gotui.New()` gives you an app with the default theme. `Title()` sets the heading shown above every screen, in bold and in your theme's main color.

### Step 2: Add a welcome screen

```go
app.Text("Welcome!\n\nThis wizard will ask you a few questions.")
```

`Text` shows a read-only screen. `\n` means "new line", so `\n\n` leaves a blank line. The user presses any key to continue. No answer is stored.

### Step 3: Ask the user to pick from a list

```go
app.Select("plan", "Choose a plan", "Free", "Pro", "Team")
```

The arguments are:

1. `"plan"`: the **key** (the drawer label).
2. `"Choose a plan"`: the question shown on screen.
3. `"Free", "Pro", "Team"`: the options. You can list as many as you like.

The user moves with ↑/↓ (or `j`/`k`) and presses Enter. The chosen text, for example `"Pro"`, is stored under `"plan"`.

### Step 4: Ask for typed text

```go
app.Input("name", "Your name: ", "e.g. Ada Lovelace")
```

The arguments are:

1. `"name"`: the key.
2. `"Your name: "`: the prompt shown before the typing area.
3. `"e.g. Ada Lovelace"`: **placeholder** text, a gray hint that disappears once the user starts typing.

By default the input accepts up to 64 characters.

### Step 5: Ask for a secret

```go
app.Password("pass", "Choose a password: ")
```

This works like `Input`, but each character shows up as `*` so nobody can read over the user's shoulder. The real text is still saved under `"pass"`.

### Step 6: Ask a yes/no question

```go
app.Confirm("newsletter", "Subscribe to the newsletter?")
```

The user toggles between Yes and No with ←/→ and presses Enter. The stored value is `true` for Yes and `false` for No.

### Step 7: Ask several things on one screen

```go
app.Form("email", "Email", "city", "City")
```

A `Form` shows several fields on one screen. You pass them as **pairs**: a key, then a label.

- `"email"` is the key, `"Email"` is what the user sees.
- `"city"` is the key, `"City"` is what the user sees.

The user moves between fields with Tab and submits with Enter on the last field. Every field is stored under its own key.

### Step 8: Run the app

```go
if err := app.Run(); err != nil {
	if errors.Is(err, gotui.ErrCancelled) {
		fmt.Println("Cancelled.")
		return
	}
	log.Fatal(err)
}
```

`Run()` plays all your screens. It returns an error if something went wrong or if the user quit early. We'll cover that in section 6.

### Step 9: Read the answers

```go
fmt.Println("\nAll done! Here's your summary:")
fmt.Printf("  Plan:       %v\n", app.Value("plan"))
fmt.Printf("  Name:       %v\n", app.Value("name"))
fmt.Printf("  Newsletter: %v\n", app.Value("newsletter"))
fmt.Printf("  Email:      %v\n", app.Value("email"))
fmt.Printf("  City:       %v\n", app.Value("city"))
```

`app.Value(...)` returns a general-purpose value, so we print it with `%v`, which works for text and true/false alike.

### The complete program

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
	app.Title("Sign-up Wizard")
	app.Center(true)
	app.Width(50)

	app.Text("Welcome!\n\nThis wizard will ask you a few questions.")
	app.Select("plan", "Choose a plan", "Free", "Pro", "Team")
	app.Input("name", "Your name: ", "e.g. Ada Lovelace")
	app.Password("pass", "Choose a password: ")
	app.Confirm("newsletter", "Subscribe to the newsletter?")
	app.Form("email", "Email", "city", "City")

	if err := app.Run(); err != nil {
		if errors.Is(err, gotui.ErrCancelled) {
			fmt.Println("Cancelled.")
			return
		}
		log.Fatal(err)
	}

	fmt.Println("\nAll done! Here's your summary:")
	fmt.Printf("  Plan:       %v\n", app.Value("plan"))
	fmt.Printf("  Name:       %v\n", app.Value("name"))
	fmt.Printf("  Newsletter: %v\n", app.Value("newsletter"))
	fmt.Printf("  Email:      %v\n", app.Value("email"))
	fmt.Printf("  City:       %v\n", app.Value("city"))
}
```

`Center(true)` and `Width(50)` are explained in section 7.

---

## 6. Handling cancellation and errors

Users can quit at any time. gotui treats that as a normal, expected event, not a crash, so you should handle it.

`Run()` returns an `error`, and there are two kinds you care about:

1. **`gotui.ErrCancelled`**: the user pressed `Esc` or `Ctrl+C`. This is not a bug. They just changed their mind.
2. **Anything else**: a real problem, for example the terminal couldn't be set up.

The standard pattern:

```go
if err := app.Run(); err != nil {
	if errors.Is(err, gotui.ErrCancelled) {
		fmt.Println("Cancelled.")
		return
	}
	log.Fatal(err)
}
```

`errors.Is(err, gotui.ErrCancelled)` asks "is this error the 'user cancelled' kind?". Remember to import Go's `errors` package.

### Your terminal is always restored

While a gotui app runs, it temporarily takes over your terminal (it hides the cursor and reads single key presses). When `Run()` finishes, the terminal is **always** put back to normal, even if the user cancelled or an error happened. You can test this: press `Ctrl+C` in the middle of an app and then type in your shell. It should behave as usual.

---

## 7. Layout: width, height, centering, scaling

Four settings control how screens are positioned. Set them before calling `Run()`.

```go
app := gotui.New()
app.Width(60)
app.Height(15)
app.Center(true)
app.TextScale(2)
app.Text("Big centered title")
app.Run()
```

| Setting | What it does | Analogy |
|---------|--------------|---------|
| `Width(n)` | Pads each line out to `n` characters. `0` means automatic (use terminal width). | The width of a sheet of paper. |
| `Height(n)` | Pads the view out to `n` lines. `0` means automatic. | The height of the sheet. |
| `Center(true)` | Centers each line inside the width. `false` left-aligns. | Centering text in a word processor. |
| `TextScale(n)` | Repeats each character `n` times horizontally. `1` is normal, `2` is double. | A stretch effect. Letters get wider. |

### Center works without a width

`Center(true)` auto-detects your terminal width. You don't have to set `Width` at all:

```go
app := gotui.New()
app.Center(true)                    // centers using your terminal's actual width
app.Text("This will be perfectly centered")
app.Run()
```

Or combine with an explicit width:

```go
app := gotui.New()
app.Center(true)
app.Width(50)
app.Input("name", "What's your name?")   // centered within 50 columns
app.Run()
```

### Vertical centering too

`Center(true)` also centers vertically when `Height` is set:

```go
app := gotui.New()
app.Center(true)
app.Height(20)
app.Text("Vertically centered too!")
app.Run()
```

A few things to know:

- `Center(true)` without `Width` auto-detects your terminal width (works on Linux, macOS, and Windows).
- `TextScale` is a visual trick. Terminals can't change real font size, so gotui stretches the text instead. It works well for titles.
- Long text is **not** wrapped or cut off. If it's wider than your terminal, the terminal wraps it for you.

---

## 8. Colors and themes

A **theme** is a set of colors used for titles, cursors, errors, and so on. Think of it as a color scheme for the whole app.

### Built-in themes

```go
gotui.DefaultTheme()  // balanced, works in most terminals
gotui.DarkTheme()     // tuned for dark backgrounds
gotui.NeonTheme()     // bright magenta and cyan, high contrast
```

Apply one to your app:

```go
app := gotui.New()
app.Theme(gotui.NeonTheme())
```

### Making your own theme

Start from an existing theme and change the parts you want:

```go
t := gotui.DefaultTheme()
t.Primary = 200   // pink
t.Danger  = 198   // deep red
t.Success = 46    // bright green

app := gotui.New()
app.Theme(t)
```

The numbers are **256-color codes**, the standard terminal palette numbered 0–255. You can look up a "256 color chart" online to pick your own.

### Named color constants

For common colors you can use names instead of numbers:

`ColorBlack`, `ColorRed`, `ColorGreen`, `ColorYellow`, `ColorBlue`, `ColorMagenta`, `ColorCyan`, `ColorWhite`, `ColorGray`

### Theme helper methods

You mostly need these when writing custom models (section 10):

| Method | What it gives you |
|--------|-------------------|
| `t.FgColor(c)` | Text that colors the foreground (the letters) with color `c` |
| `t.BgColor(c)` | Text that colors the background with color `c` |
| `t.Bold()` | Switches on bold |
| `t.Reverse()` | Swaps foreground and background (used for cursors) |
| `t.Reset()` | Switches everything back to normal |

Always call `t.Reset()` after styled text, or the styling will "leak" into everything that follows.

There is also `gotui.Style{}`, a beginner-friendly wrapper with `FgColor`, `BgColor`, `Bold`, `Reverse`, `Reset`, and `Rgb(r, g, b)`.

---

## 9. Standalone components

You don't have to build a whole app. Every component can run on its own and hand you its answer directly, which is handy when you just need one question answered inside a bigger program.

```go
name, err := gotui.NewInput().
	WithPrompt("Name: ").
	WithPlaceholder("e.g. Ada").
	WithCharLimit(32).
	Run()
if err != nil {
	if errors.Is(err, gotui.ErrCancelled) {
		fmt.Println("Cancelled")
		return
	}
	log.Fatal(err)
}
fmt.Println("Hello,", name)
```

The `.WithSomething(...)` calls are **chained**: each one tweaks a setting and hands the component back so you can add the next. Order doesn't matter, and `Run()` goes last.

### Available components

| Component | Constructor | `Run()` returns |
|-----------|-------------|-----------------|
| Text input | `gotui.NewInput()` | `(string, error)` |
| Password input | `gotui.NewPassword()` | `(string, error)` |
| Choice list | `gotui.NewSelect()` | `(string, error)` |
| Yes/no | `gotui.NewConfirm()` | `(bool, error)` |
| Multi-field form | `gotui.NewForm()` | `error` (read answers with `Get(key)`) |
| Read-only text | `gotui.NewText()` | `error` |
| Loading spinner | `gotui.NewSpinner()` | `(Msg, error)` |

### Configuration examples

```go
// A list
color, err := gotui.NewSelect().
	WithTitle("Pick a color").
	WithOptions("Red", "Green", "Blue").
	WithCursor(1).                    // start with the 2nd option highlighted
	WithTheme(gotui.DarkTheme()).
	Run()

// A yes/no that starts on "Yes"
ok, err := gotui.NewConfirm().
	WithTitle("Delete this file?").
	DefaultYes().
	Run()
```

### Standalone forms

A form returns only an error, so you read its answers afterwards with `Get`:

```go
form := gotui.NewForm().
	WithTitle("Sign up").
	WithInput("name", "Full name").
	WithInput("email", "Email").
	WithPassword("pass", "Password")

if err := form.Run(); err != nil {
	log.Fatal(err)
}

fmt.Println(form.Get("name"))
fmt.Println(form.Get("email"))
```

---

## 10. Custom models

The Simple App API covers most needs. When it doesn't (a game, a live dashboard, a custom animation), you can write your own component using a **Model**.

### The idea

A model is like a recipe card with three parts:

1. **`Init`**: what to do at the very start.
2. **`Update`**: what to do when something happens (a key press, a timer tick). It returns the new state.
3. **`View`**: how to draw the current state as text.

gotui repeats the same loop: *something happens → `Update` changes the state → `View` redraws the screen*.

```go
type Model interface {
	Init() Cmd
	Update(msg Msg) (Model, Cmd)
	View() string
}
```

- A **`Msg`** is "something that happened", for example a key press.
- A **`Cmd`** is "something to do next", for example quit the program.

### Example: a counter

```go
package main

import (
	"fmt"

	"github.com/Cgtlpa/gotui"
)

// The state: just one number.
type counter struct {
	count int
}

// Init runs once at the start. We have nothing to start, so return nil.
func (m counter) Init() gotui.Cmd { return nil }

// Update runs whenever something happens.
func (m counter) Update(msg gotui.Msg) (gotui.Model, gotui.Cmd) {
	// We only care about key presses. Ignore everything else.
	km, ok := msg.(gotui.KeyMsg)
	if !ok {
		return m, nil
	}

	switch km.Key.Type {
	case gotui.KeyRune: // a normal character key
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

// View draws the current state.
func (m counter) View() string {
	th := gotui.DarkTheme()

	color := th.FgColor(th.Foreground)
	if m.count > 0 {
		color = th.FgColor(gotui.ColorGreen)
	} else if m.count < 0 {
		color = th.FgColor(gotui.ColorRed)
	}

	return fmt.Sprintf(
		"%s%sCount%s: %s%d%s\n\nPress + / - to change, q to quit.\n",
		th.Bold(), th.FgColor(th.Primary), th.Reset(),
		color, m.count, th.Reset(),
	)
}

func main() {
	app := gotui.New()
	app.Custom(counter{count: 0})
	app.Run()
}
```

How it works:

- The number lives in the `counter` struct. This is the **state**.
- When you press `+`, `Update` adds 1 and returns the new model. gotui then calls `View`, and the screen shows the new number.
- Pressing `q` or Enter returns `gotui.Quit`, which ends the program.

### Commands (`Cmd`)

A `Cmd` is simply a function of type `func() Msg`. Common uses:

- `gotui.Quit`: exit the program.
- `gotui.Batch(cmd1, cmd2)`: run several commands at once.
- Your own functions for background work like timers or network calls. They return a message when done, and that message arrives in `Update`.

### Running a model on its own

You can skip the `App` and run a model directly:

```go
program := gotui.NewProgram(counter{})
program.Run()
```

### Sending messages from outside

If other code (for example a background goroutine) needs to tell your program something, use `Send`:

```go
program.Send(gotui.KeyMsg{Key: gotui.Key{Type: gotui.KeyRune, Rune: 'q'}})
```

---

## 11. Keyboard reference

These keys behave the same in every component:

| Key | Input / Password | Select | Confirm | Form | Text |
|-----|------------------|--------|---------|------|------|
| Typing | inserts a character | — | — | inserts a character | — |
| `←` / `→` | moves the cursor | — | toggles Yes/No | moves the cursor | — |
| `↑` / `↓` or `j` / `k` | — | moves the selection | — | — | — |
| `Tab` | — | — | — | next field | — |
| `Enter` | submits | confirms | confirms | submits on the last field | continues |
| `Backspace` | deletes a character | — | — | deletes a character | — |
| `Esc` | cancels | cancels | cancels | cancels | cancels |
| `Ctrl+C` | cancels | cancels | cancels | cancels | cancels |

Both `Esc` and `Ctrl+C` make `Run()` return `gotui.ErrCancelled`.

---

## 12. Method reference

### App methods

| Method | What it does |
|--------|--------------|
| `gotui.New()` | Creates a new app with the default theme. |
| `app.Title(text)` | Sets the heading shown above every screen. |
| `app.Center(bool)` | Centers (or left-aligns) text within the width. |
| `app.Width(n)` | Pads lines to `n` characters (`0` = auto). |
| `app.Height(n)` | Pads the view to `n` lines (`0` = auto). |
| `app.TextScale(n)` | Stretches text horizontally by `n` times. |
| `app.Theme(th)` | Sets the color theme. |
| `app.Text(text)` | Adds a read-only screen. Any key continues. |
| `app.Select(key, title, options...)` | Adds a single-choice list. |
| `app.Input(key, prompt, placeholder)` | Adds a one-line text input (max 64 characters by default). |
| `app.Password(key, prompt)` | Adds a masked input. |
| `app.Confirm(key, title)` | Adds a yes/no question (`true` = Yes). |
| `app.Form(key, label, ...)` | Adds a multi-field form. Pass key/label pairs. |
| `app.Custom(model)` | Adds your own custom model as a screen. |
| `app.Value(key)` | Gets a stored answer after `Run()` (`nil` if unset). |
| `app.Run()` | Plays all screens. Returns `nil`, or `gotui.ErrCancelled`, or another error. |

### Built-in themes

`gotui.DefaultTheme()`, `gotui.DarkTheme()`, `gotui.NeonTheme()`

---

## 13. Known limitations

- **Resizing:** Changing the terminal size while the app is running isn't handled. Resize before you start.
- **Font size:** `TextScale(n)` stretches characters horizontally. It does not change the real font size, which terminals don't allow.
- **No wrapping:** Long text is neither wrapped nor clipped.

---

## Where to go next

The `examples/` folder has runnable programs. Run any of them with `go run ./examples/<name>`:

| Example | What it shows |
|---------|---------------|
| `simple_input` | The Simple App API and reading answers with `app.Value()` |
| `form` | A multi-field form read with `Get(key)` |
| `custom` | A custom model plugged in with `app.Custom()` |
| `demo` | A combined showcase with themes and results |
| `layout` | Center, Width, Height, and TextScale |

To see every component rendered in your terminal without running an interactive app, use:

```bash
go run ./cmd/preview
```

Happy building! 🍵
