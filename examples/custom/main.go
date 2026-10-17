// examples/custom demonstrates adding a fully custom Model via the App API.
//
// Run with: go run ./examples/custom

package main

import (
	"fmt"
	"log"

	"github.com/Cgtlpa/gotui"
)

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

func main() {
	app := gotui.New()
	app.Title("Custom Counter")
	app.Center(true)
	app.Custom(counter{count: 0})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Goodbye!")
}
