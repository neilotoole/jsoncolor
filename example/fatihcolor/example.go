// Package main is a trivial program that outputs colorized JSON,
// demonstrating how to use the fatihcolor helper to build
// the jsoncolor.Colors struct.
package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"github.com/neilotoole/jsoncolor"

	"github.com/neilotoole/jsoncolor/helper/fatihcolor"
)

func main() {
	var enc *jsoncolor.Encoder

	// Note: this check will fail if running inside Goland (and
	// other IDEs?) as IsColorTerminal will return false.
	if jsoncolor.IsColorTerminal(os.Stdout) {
		fclrs := fatihcolor.DefaultColors()

		// Change some values, just for fun
		fclrs.Number = color.New(color.FgBlue)
		fclrs.String = color.New(color.FgCyan)

		clrs := fatihcolor.ToCoreColors(fclrs)
		out := colorable.NewColorable(os.Stdout)
		enc = jsoncolor.NewEncoder(out)
		enc.SetColors(clrs)
	} else {
		enc = jsoncolor.NewEncoder(os.Stdout)
	}
	enc.SetIndent("", "  ")

	m := map[string]interface{}{
		"a": 1,
		"b": true,
		"c": "hello",
	}

	if err := enc.Encode(m); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
