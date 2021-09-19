# jsoncolor

Package `neilotoole/jsoncolor` provides a replacement for `encoding/json`
that can output colorized JSON.

## Usage

Get the package per the normal mechanism:

```shell
go get -u github.com/neilotoole/jsoncolor
```

Use as follows:

```go
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
```

## Example app: `jc`

See `cmd/jc` for a trivial CLI implementation that can accept JSON input
and output in color.

```shell
# From project root
go install ./cmd/jc
cat ./testdata/sakila_actor.json | jc
```

### History

This package is an extract of [neilotoole/sq](https://github.com/neilotoole/sq)'s `jsonw`
package, which itself is a fork of the [segment.io/encoding](https://github.com/segmentio/encoding) JSON
encoding package. Note that `jsoncolor` was forked from Segment's package at `v0.1.14`, so
this codebase is quite of out sync by now.

### Acknowledgments

- [jq](https://stedolan.github.io/jq/): sine qua non.
- [`segmentio/encoding`](https://github.com/segmentio/encoding): this package is layered into Segment's JSON encoder. Much gratitude to that team.
- [`neilotoole/sq`](https://github.com/neilotoole/sq): `jsoncolor` is effectively an extract of the code created specifically for the `sq` tool.

### Related

- [nwidger/jsoncolor](https://github.com/nwidger/jsoncolor)
- [hokaccha/go-prettyjson](https://github.com/hokaccha/go-prettyjson): doesn't provide an encoder.
- [TylerBrock/colorjson](https://github.com/TylerBrock/colorjson): doesn't provide an encoder.


