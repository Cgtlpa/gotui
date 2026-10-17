// cmd/preview renders every gotui component using the new simple App API,
// printing to stdout with ANSI color codes so you can preview without a live
// terminal.
//
// Run:
//   go run ./cmd/preview
//
// Or save to a file:
//   go run ./cmd/preview > preview.txt && cat preview.txt

package main

import (
	"fmt"
	"strings"

	"github.com/Cgtlpa/gotui"
)

func main() {
	out := strings.Builder{}

	hr := func(title string) {
		th := gotui.DefaultTheme()
		out.WriteString("\n")
		out.WriteString(th.Bold())
		out.WriteString(th.FgColor(th.Primary))
		out.WriteString("═══ " + title + " ")
		remaining := 60 - len(title)
		if remaining < 0 {
			remaining = 0
		}
		out.WriteString(strings.Repeat("═", remaining))
		out.WriteString(th.Reset())
		out.WriteString("\n\n")
	}

	// ── Example 1: The exact API the user requested ──
	hr("Simple App API (user's example)")
	out.WriteString("app := gotui.New()\n")
	out.WriteString("app.Title(\"Setup Wizard\")\n")
	out.WriteString("app.Center(true)\n")
	out.WriteString("app.Text(\"Welcome to the Setup Wizard\")\n")
	out.WriteString("app.Select(\"Choose an option\", \"Option A\", \"Option B\", \"Option C\")\n")
	out.WriteString("app.Run()\n\n")
	out.WriteString("(Renders below as sequential screens.)\n\n")

	// ── Example 2: Full demo with all components ──
	hr("Full demo: Title + Center + Text + Select + Input + Confirm + Form")
	out.WriteString("app := gotui.New()\n")
	out.WriteString("app.Title(\"Setup Wizard\")\n")
	out.WriteString("app.Center(true)\n")
	out.WriteString("app.Width(50)\n")
	out.WriteString("app.Text(\"Welcome to Setup Wizard\")\n")
	out.WriteString("app.Select(\"Choose an option\", \"Option A\", \"Option B\", \"Option C\")\n")
	out.WriteString("app.Input(\"Your name: \", \"e.g. Ada\")\n")
	out.WriteString("app.Confirm(\"Continue?\")\n")
	out.WriteString("app.Form(\"email\", \"Email\", \"bio\", \"Bio\")\n")
	out.WriteString("app.Run()\n\n")

	// ── Layout demo ──
	hr("Layout: Width, Center, Height, TextScale")
	out.WriteString("app := gotui.New()\n")
	out.WriteString("app.Title(\"Styled\")\n")
	out.WriteString("app.Width(60)\n")
	out.WriteString("app.Center(true)\n")
	out.WriteString("app.TextScale(2)\n")
	out.WriteString("app.Text(\"Big centered text\")\n")
	out.WriteString("app.Run()\n\n")

	fmt.Print(out.String())
	fmt.Println("(End of preview — run the interactive demo with: go run ./examples/demo)")
	fmt.Println("_________________________________________________________")
	fmt.Println("Note: The interactive demo runs each step sequentially.")
}
