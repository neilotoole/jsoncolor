package main

import (
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/neilotoole/jsoncolor"

	"github.com/neilotoole/jsoncolor/helper/fatihcolor"
)

func main() {
	var (
		out io.Writer = os.Stdout
		enc *jsoncolor.Encoder
	)

	// Note: this check will fail if running in Goland, as IsColorTerminal
	// will return false.
	if jsoncolor.IsColorTerminal(out) {
		fclrs := fatihcolor.DefaultColors()

		// Change some values, just for fun
		fclrs.Number = color.New(color.FgBlue) // Change one of the values for fun
		fclrs.String = color.New(color.FgCyan) // Change one of the values for fun

		clrs := fatihcolor.ToCoreColors(fclrs)
		out = colorable.NewColorable(out.(*os.File))
		enc = jsoncolor.NewEncoder(out)
		enc.SetColors(clrs)
	} else {
		enc = jsoncolor.NewEncoder(out)
	}
	enc.SetIndent("", "  ")

	m := map[string]interface{}{
		"a": 1,
		"b": true,
		"c": "hello",
	}

	if err := enc.Encode(m); err != nil {
		panic(err)
	}
}
