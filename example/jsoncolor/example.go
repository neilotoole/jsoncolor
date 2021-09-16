package main

import (
	"io"
	"os"

	"github.com/mattn/go-colorable"

	"github.com/neilotoole/jsoncolor"
)

func main() {
	var (
		out io.Writer = os.Stdout
		enc *jsoncolor.Encoder
	)

	// Note: this check will fail if running in Goland, as IsColorTerminal
	// will return false.
	if jsoncolor.IsColorTerminal(out) {
		out = colorable.NewColorable(out.(*os.File))
		enc = jsoncolor.NewEncoder(out)
		clrs := jsoncolor.DefaultColors()
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
