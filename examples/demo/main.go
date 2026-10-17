// examples/demo is a combined showcase of every App component and theme,
// including result retrieval via app.Value().
//
// Run with: go run ./examples/demo

package main

import (
	"errors"
	"fmt"
	"log"

	"gotui"
)

func main() {
	app := gotui.New()
	app.Title("Welcome to Silen Setup")
	app.Center(true)
	app.Width(50)
	app.Theme(gotui.NeonTheme())

	app.Text("A Go TUI framework with the customizability of Bubble Tea " +
		"and the simplicity of Huh.\n\nThis demo walks through input, " +
		"select, confirm, and form components.")

	app.Select("wm", "Pick your role", "Developer", "Designer", "Manager", "Other")
	app.Input("name", "Your name: ", "e.g. Ada Lovelace")
	app.Confirm("confirm", "Create account?")
	app.Form("email", "Email", "bio", "Bio")

	if err := app.Run(); err != nil {
		if errors.Is(err, gotui.ErrCancelled) {
			fmt.Println("\nSetup cancelled.")
			return
		}
		log.Fatal(err)
	}

	fmt.Println("\n═══ Setup Complete ═══")
	fmt.Printf("  Role:   %s\n", app.Value("wm"))
	fmt.Printf("  Name:   %s\n", app.Value("name"))
	fmt.Printf("  Email:  %s\n", app.Value("email"))
	fmt.Printf("  Bio:    %s\n", app.Value("bio"))
	fmt.Println("══════════════════════")
}
