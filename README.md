[![Actions Status](https://github.com/neilotoole/jsoncolor/workflows/Go/badge.svg)](https://github.com/neilotoole/jsoncolor/actions?query=workflow%3AGo)
[![Go Report Card](https://goreportcard.com/badge/neilotoole/jsoncolor)](https://goreportcard.com/report/neilotoole/jsoncolor)
[![release](https://img.shields.io/badge/release-v0.2.0-green.svg)](https://github.com/neilotoole/jsoncolor/releases/tag/v0.2.0)
[![Go Reference](https://pkg.go.dev/badge/github.com/neilotoole/jsoncolor.svg)](https://pkg.go.dev/github.com/neilotoole/jsoncolor)
[![license](https://img.shields.io/github/license/neilotoole/jsoncolor)](./LICENSE)

# jsoncolor

Package `neilotoole/jsoncolor` is a drop-in replacement for `encoding/json`
that outputs colorized JSON.

Why? Well, `jq` colorizes its output by default. And at the time this package was
created, I was not aware of any other JSON colorization package that performed
colorization in-line in the encoder.

From the example [`jc`](./cmd/jc) app:

![jsoncolor-output](https://github.com/neilotoole/jsoncolor/wiki/images/jsoncolor-example-output2.png)

## Usage

Get the package per the normal mechanism:

```shell
go get -u github.com/neilotoole/jsoncolor
```

Then:

```go
package main

import (
  "fmt"
  "github.com/mattn/go-colorable"
  json "github.com/neilotoole/jsoncolor"
  "os"
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
```

### Configuration

To enable colorization, invoke `enc.SetColors`.

The `jsoncolor.Colors` struct holds color config. The zero value
and `nil` are both safe for use (resulting in no colorization).

The `DefaultColors` func returns a `Colors` struct that produces results
similar to `jq`:

```go
// DefaultColors returns the default Colors configuration.
// These colors largely follow jq's default colorization,
// with some deviation.
func DefaultColors() *Colors {
  return &Colors{
    Null:   Color("\x1b[2m"),
    Bool:   Color("\x1b[1m"),
    Number: Color("\x1b[36m"),
    String: Color("\x1b[32m"),
    Key:    Color("\x1b[34;1m"),
    Bytes:  Color("\x1b[2m"),
    Time:   Color("\x1b[32;2m"),
    Punc:   Color{}, // No colorization
  }
}
```

As seen above, use the `Color` zero value (`Color{}`) to
disable colorization for that JSON element.


### Helper for `fatih/color`

It can be inconvenient to use terminal codes, e.g. `json.Color("\x1b[36m")`.
A helper package provides an adapter for the [`fatih/color`](https://github.com/fatih/color) package.

```go
  // import "github.com/neilotoole/jsoncolor/helper/fatihcolor"
  // import "github.com/fatih/color"

  out := colorable.NewColorable(os.Stdout) // needed for Windows
  enc = json.NewEncoder(out)

  fclrs := fatihcolor.DefaultColors()
  // Change some values, just for fun
  fclrs.Number = color.New(color.FgBlue)
  fclrs.String = color.New(color.FgCyan)
  
  clrs := fatihcolor.ToCoreColors(fclrs)
  enc.SetColors(clrs)
```

### Drop-in for `encoding/json`

This package is a full drop-in for stdlib `encoding/json`
(thanks to the `segmentio/encoding/json` pkg being a full drop-in).

To drop-in, just use an import alias:

```go
  import json "github.com/neilotoole/jsoncolor"
```

## Example app: `jc`

See [`./cmd/jc`](.cmd/jc) for a trivial CLI implementation that can accept JSON input,
and output that JSON in color.

```shell
# From project root
go install ./cmd/jc
cat ./testdata/sakila_actor.json | jc
```


### History

This package is an extract of [`sq`](https://github.com/neilotoole/sq)'s JSON encoding
package, which itself is a fork of the [`segment.io/encoding`](https://github.com/segmentio/encoding) JSON
encoding package.

Note that the original `jsoncolor` codebase was forked from Segment's codebase at `v0.1.14`, so
the codebases are quite of out sync by now.

### Notes

- The `.golangci.yml` linter settings have been fiddled with to hush linting issues inherited from
  the `segmentio` codebase at the time of forking. Thus, the linter report may not be of great use.
  In an ideal world, the `jsoncolor` functionality would be ported to a more recent (and better-linted)
  version of the `segementio` codebase.

### Acknowledgments

- [`jq`](https://stedolan.github.io/jq/): sine qua non.
- [`segmentio/encoding`](https://github.com/segmentio/encoding): `jsoncolor` is layered into Segment's JSON encoder. They did the hard work. Much gratitude to that team.
- [`sq`](https://github.com/neilotoole/sq): `jsoncolor` is effectively an extract of the code created specifically for `sq`.
- [`mattn/go-colorable`](https://github.com/mattn/go-colorable): no project is complete without `mattn` having played a role.
- [`fatih/color`](https://github.com/fatih/color): the color library.

### Related

- [`nwidger/jsoncolor`](https://github.com/nwidger/jsoncolor)
- [`hokaccha/go-prettyjson`](https://github.com/hokaccha/go-prettyjson)
- [`TylerBrock/colorjson`](https://github.com/TylerBrock/colorjson)
