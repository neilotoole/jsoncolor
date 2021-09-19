// Package main is a trivial program that outputs colorized JSON,
// using jsoncolor.DefaultColors.
package main

import (
	"fmt"
	"github.com/nwidger/jsoncolor"
	"os"
)

func main() {
	var enc *jsoncolor.Encoder

	f := jsoncolor.NewFormatter()
	f.Indent = "  "
	enc = jsoncolor.NewEncoderWithFormatter(os.Stdout, f)
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
