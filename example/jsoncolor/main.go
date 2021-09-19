// Package main is a trivial program that outputs colorized JSON,
// using jsoncolor.DefaultColors.
package main

import (
	"fmt"
	"github.com/mattn/go-colorable"
	"github.com/neilotoole/jsoncolor"
	"os"
)

func main() {
	var enc *jsoncolor.Encoder

	// Note: this check will fail if running inside Goland (and
	// other IDEs?) as IsColorTerminal will return false.
	if jsoncolor.IsColorTerminal(os.Stdout) {
		// Safe to use color
		out := colorable.NewColorable(os.Stdout) // needed for Windows
		enc = jsoncolor.NewEncoder(out)
		enc.SetColors(jsoncolor.DefaultColors())
	} else {
		// Can't use color; but the encoder will still work
		enc = jsoncolor.NewEncoder(os.Stdout)
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
