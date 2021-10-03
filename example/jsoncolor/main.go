// Package main is a trivial program that outputs colorized JSON,
// using json.DefaultColors.
package main

import (
	"fmt"
	"os"

	"github.com/mattn/go-colorable"
	json "github.com/neilotoole/jsoncolor"
)

func main() {
	var enc *json.Encoder

	// Note: this check will fail if running inside Goland (and
	// other IDEs?) as IsColorTerminal will return false.
	if json.IsColorTerminal(os.Stdout) {
		// Safe to use color
		out := colorable.NewColorable(os.Stdout) // needed for Windows
		enc = json.NewEncoder(out)

		// DefaultColors are similar to jq
		clrs := json.DefaultColors()

		// Change some values, just for fun
		clrs.Bool = json.Color("\x1b[36m") // Change the bool color
		clrs.String = json.Color{}         // Disable the string color

		enc.SetColors(clrs)
	} else {
		// Can't use color; but the encoder will still work
		enc = json.NewEncoder(os.Stdout)
	}

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
