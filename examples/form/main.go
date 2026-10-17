// examples/form demonstrates the multi-field form via the App API.
//
// Run with: go run ./examples/form

package main

import (
	"fmt"
	"log"

	"gotui"
)

func main() {
	app := gotui.New()
	app.Title("Create an account")
	app.Center(true)
	app.Width(50)

	app.Form("name", "Full name", "email", "Email", "password", "Password")

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("\nAccount created!")
}
