package main

import (
	"os"

	"github.com/neilotoole/jsoncolor/helper/fatihcolor"

	"github.com/neilotoole/jsoncolor"
)

func main() {
	main1()
	main2()
	main3()
}

func main1() {
	m := map[string]interface{}{
		"a": 1,
		"b": true,
		"c": "hello",
	}

	clrs := jsoncolor.DefaultColors()
	enc := jsoncolor.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetColors(clrs)

	if err := enc.Encode(m); err != nil {
		panic(err)
	}
}

func main2() {
	m := map[string]interface{}{
		"a": 1,
		"b": true,
		"c": "hello",
	}

	clrs := fatihcolor.NewDefaultColors()
	enc := fatihcolor.NewEncoder(os.Stdout, clrs)
	enc.SetIndent("", "  ")

	if err := enc.Encode(m); err != nil {
		panic(err)
	}
}
func main3() {
	m := map[string]interface{}{
		"a": 1,
		"b": true,
		"c": "hello",
	}

	clrs := fatihcolor.NewDefaultColors()
	coreClrs := fatihcolor.ToCoreColors(clrs)
	enc := jsoncolor.NewEncoder(os.Stdout)
	enc.SetColors(coreClrs)
	enc.SetIndent("", "  ")

	if err := enc.Encode(m); err != nil {
		panic(err)
	}
}
