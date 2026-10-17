// examples/simple_input demonstrates the simple App API.
//
// Run with: go run ./examples/simple_input

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

	app.Text("Welcome to Silen")
	app.Select("wm", "Choose a window manager", "InstantWM", "Niri", "DWM")
	app.Input("name", "Your name: ", "e.g. Ada")
	app.Confirm("confirm", "Continue?")
	app.Form("email", "Email", "bio", "Bio")

	if err := app.Run(); err != nil {
		if errors.Is(err, gotui.ErrCancelled) {
			fmt.Println("\nSetup cancelled.")
			return
		}
		log.Fatal(err)
	}

	fmt.Println("\nSetup complete!")
	fmt.Printf("  WM:    %s\n", app.Value("wm"))
	fmt.Printf("  Name:  %s\n", app.Value("name"))
	fmt.Printf("  Email: %s\n", app.Value("email"))
	fmt.Printf("  Bio:   %s\n", app.Value("bio"))
}
