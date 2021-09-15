package main

import (
	"os"

	"github.com/neilotoole/jsoncolor/helper/fatihcolor"
)

func main() {
	m := map[string]interface{}{
		"a": 1,
		"b": true,
		"c": "hello",
	}

	clrs := fatihcolor.NewDefaultColors()
	enc := fatihcolor.NewEncoder(os.Stdout, clrs)

	if err := enc.Encode(m); err != nil {
		panic(err)
	}
}
