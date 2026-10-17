// Package gotui is a Go TUI (Terminal User Interface) framework with the
// customizability of Bubble Tea (charm.sh) and the simplicity of Huh (charm.sh).
//
// Built for beginners, powerful enough for experts. Zero external dependencies
// — pure Go standard library.
//
// Quick start:
//
//	app := gotui.New()
//	app.Title("Setup Wizard")
//	app.Center(true)
//	app.Text("Welcome to the Setup Wizard")
//	app.Select("choice", "Choose an option", "InstantWM", "Niri", "DWM")
//	app.Input("name", "Your name: ", "e.g. Ada")
//	app.Confirm("confirm", "Continue?")
//
//	if err := app.Run(); err != nil {
//	    if errors.Is(err, gotui.ErrCancelled) {
//	        fmt.Println("Cancelled")
//	    }
//	    log.Fatal(err)
//	}
//
//	fmt.Println("Choice:", app.Value("choice"))
//
// See the README for a complete tutorial and API reference.
package gotui
