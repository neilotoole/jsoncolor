package jsoncolor_test

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/neilotoole/jsoncolor/helper/fatihcolor"

	"github.com/mattn/go-colorable"
	json "github.com/neilotoole/jsoncolor"
)

// ExampleEncoder shows use of neilotoole/jsoncolor Encoder.
func ExampleEncoder() {
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
	}
}

// Example_fatihColor shows use of the fatihcolor helper package
// with jsoncolor.
func Example_fatihColor() {
	var enc *json.Encoder

	// Note: this check will fail if running inside Goland (and
	// other IDEs?) as IsColorTerminal will return false.
	if json.IsColorTerminal(os.Stdout) {
		out := colorable.NewColorable(os.Stdout)
		enc = json.NewEncoder(out)

		fclrs := fatihcolor.DefaultColors()

		// Change some values, just for fun
		fclrs.Number = color.New(color.FgBlue)
		fclrs.String = color.New(color.FgCyan)

		clrs := fatihcolor.ToCoreColors(fclrs)
		enc.SetColors(clrs)
	} else {
		enc = json.NewEncoder(os.Stdout)
	}
	enc.SetIndent("", "  ")

	m := map[string]interface{}{
		"a": 1,
		"b": true,
		"c": "hello",
	}

	if err := enc.Encode(m); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
