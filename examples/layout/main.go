// examples/layout demonstrates the Layout options: Center, Width, Height,
// and TextScale. These can be applied to the whole app.
//
// Run with: go run ./examples/layout

package main

import (
	"fmt"
	"log"

	"github.com/Cgtlpa/gotui"
)

func main() {
	// Centered input with fixed width.
	app := gotui.New()
	app.Title("Styled Input")
	app.Center(true)
	app.Width(40)
	app.Input("name", "Name: ", "enter your name")

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Hello!\n\n")

	// Big text with TextScale 2.
	app2 := gotui.New()
	app2.Title("Big Text")
	app2.Center(true)
	app2.Width(60)
	app2.TextScale(2)
	app2.Text("This text uses TextScale 2\nIt looks bigger!")

	if err := app2.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nDemo complete!")
}
