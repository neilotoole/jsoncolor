[![Actions Status](https://github.com/neilotoole/jsoncolor/workflows/Go/badge.svg)](https://github.com/neilotoole/jsoncolor/actions?query=workflow%3AGo)
[![Go Report Card](https://goreportcard.com/badge/neilotoole/jsoncolor)](https://goreportcard.com/report/neilotoole/jsoncolor)
[![release](https://img.shields.io/badge/release-v0.10.0-green.svg)](https://github.com/neilotoole/jsoncolor#v0100)
[![Go Reference](https://pkg.go.dev/badge/github.com/neilotoole/jsoncolor.svg)](https://pkg.go.dev/github.com/neilotoole/jsoncolor)
[![license](https://img.shields.io/github/license/neilotoole/jsoncolor)](./LICENSE)

# jsoncolor

Package `neilotoole/jsoncolor` is a drop-in replacement for stdlib
[`encoding/json`](https://pkg.go.dev/encoding/json) that outputs colorized JSON.

Why? Well, [`jq`](https://jqlang.github.io/jq/) colorizes its output by default, and color output
is desirable for many Go CLIs. This package performs colorization (and indentation) inline
in the encoder, and is significantly faster than stdlib at indentation.

From the example [`jc`](./cmd/jc/main.go) app:

![jsoncolor-output](./splash.png)

## Usage

Get the package per the normal mechanism (requires Go 1.25+):

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

To enable colorization, invoke [`enc.SetColors`](https://pkg.go.dev/github.com/neilotoole/jsoncolor#Encoder.SetColors).

The [`Colors`](https://pkg.go.dev/github.com/neilotoole/jsoncolor#Colors) struct
holds color config. The zero value and `nil` are both safe for use (resulting in no colorization).

The [`DefaultColors`](https://pkg.go.dev/github.com/neilotoole/jsoncolor#DefaultColors) func
returns a `Colors` struct that produces results similar to `jq`:

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
    // The granular punctuation fields (Brackets, Braces, Comma,
    // Colon) are intentionally left as the zero value so that they
    // fall back to Punc, preserving the default (uncolored)
    // punctuation behavior.
    TextMarshaler: Color("\x1b[32m"), // Same as String
  }
}
```

As seen above, use the `Color` zero value (`Color{}`) to
disable colorization for that JSON element.

### Color reset

`Color` is the prefix only. The encoder closes every colorized token with the
fixed sequence `\x1b[0m` (SGR 0), and there is no way for a caller to supply a
different closer.

`\x1b[0m` resets *every* terminal attribute, not just the ones the prefix set.
That is the safe default: no attribute can leak out of a token. The trade-off is
composability. jsoncolor output is not safe to nest inside a region that is
already styled, because the first colorized token clears that styling for the
remainder of the output. Rendering colorized JSON inside a diff line that has a
background color, inside a TUI panel with an inherited style, or after a styled
log prefix will each drop the surrounding style.

An attribute-specific closer (`\x1b[39m` for "default foreground", `\x1b[22m`
for "normal intensity", and so on) would leave untouched attributes alone, but
jsoncolor cannot compute one. `Color` arrives as already-rendered opaque bytes,
so given `\x1b[34;1m` the encoder has no way to know that those parameters mean
blue and bold. See [#75](https://github.com/neilotoole/jsoncolor/issues/75) for
the full analysis, including the API options if this is ever addressed.

### Helper for `fatih/color`

It can be inconvenient to use terminal codes, e.g. `json.Color("\x1b[36m")`.
A helper package provides an adapter for [`fatih/color`](https://github.com/fatih/color).

```go
  // import "github.com/neilotoole/jsoncolor/helper/fatihcolor"
  // import "github.com/fatih/color"
  // import "github.com/mattn/go-colorable"
  
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

This package is a full drop-in for stdlib [`encoding/json`](https://pkg.go.dev/encoding/json)
(thanks to the ancestral [`segmentio/encoding/json`](https://pkg.go.dev/github.com/segmentio/encoding/json)
pkg being a full drop-in).

To drop-in, just use an import alias:

```go
  import json "github.com/neilotoole/jsoncolor"
```

## Example app: `jc`

See [`cmd/jc`](cmd/jc/main.go) for a trivial CLI implementation that can accept JSON input,
and output that JSON in color.

```shell
# From project root
$ go install ./cmd/jc
$ cat ./testdata/sakila_actor.json | jc
```

## Benchmarks

Note that this package contains [`golang_bench_test.go`](./golang_bench_test.go), which
is inherited from `segmentj`. But here we're interested in [`benchmark_test.go:BenchmarkEncode`](./benchmark_test.go),
which benchmarks encoding performance versus other JSON encoder packages.
The results below benchmark the following:

- Stdlib [`encoding/json`](https://pkg.go.dev/encoding/json) (`go1.17.1`).
- [`segmentj`](https://github.com/segmentio/encoding): `v0.1.14`, which was when `jsoncolor` was forked. The newer `segmentj` code performs even better.
- `neilotoole/jsoncolor`: (this package) `v0.6.0`.
- [`nwidger/jsoncolor`](https://github.com/nwidger/jsoncolor): `v0.3.0`, latest at time of benchmarks.

Note that two other Go JSON colorization packages ([`hokaccha/go-prettyjson`](https://github.com/hokaccha/go-prettyjson) and
[`TylerBrock/colorjson`](https://github.com/TylerBrock/colorjson)) are excluded from
these benchmarks because they do not provide a stdlib-compatible `Encoder` impl.

```
$ go test -bench=BenchmarkEncode -benchtime="5s"
goarch: amd64
pkg: github.com/neilotoole/jsoncolor
cpu: Intel(R) Core(TM) i9-9880H CPU @ 2.30GHz
BenchmarkEncode/stdlib_NoIndent-16                           181          33047390 ns/op         8870685 B/op     120022 allocs/op
BenchmarkEncode/stdlib_Indent-16                             124          48093178 ns/op        10470366 B/op     120033 allocs/op
BenchmarkEncode/segmentj_NoIndent-16                         415          14658699 ns/op         3788911 B/op      10020 allocs/op
BenchmarkEncode/segmentj_Indent-16                           195          30628798 ns/op         5404492 B/op      10025 allocs/op
BenchmarkEncode/neilotoole_NoIndent_NoColor-16               362          16522399 ns/op         3789034 B/op      10020 allocs/op
BenchmarkEncode/neilotoole_Indent_NoColor-16                 303          20146856 ns/op         5460753 B/op      10021 allocs/op
BenchmarkEncode/neilotoole_NoIndent_Color-16                 295          19989420 ns/op        10326019 B/op      10029 allocs/op
BenchmarkEncode/neilotoole_Indent_Color-16                   246          24714163 ns/op        11996890 B/op      10030 allocs/op
BenchmarkEncode/nwidger_NoIndent_NoColor-16                   10         541107983 ns/op        92934231 B/op    4490210 allocs/op
BenchmarkEncode/nwidger_Indent_NoColor-16                      7         798088086 ns/op        117258321 B/op   6290213 allocs/op
BenchmarkEncode/nwidger_indent_NoIndent_Colo-16               10         542002051 ns/op        92935639 B/op    4490224 allocs/op
BenchmarkEncode/nwidger_indent_Indent_Color-16                 7         799928353 ns/op        117259195 B/op   6290220 allocs/op
```

As always, take benchmarks with a large grain of salt, as they're based on a (small) synthetic benchmark.
More benchmarks would give a better picture (and note as well that the benchmarked `segmentj` is an older version, `v0.1.14`).

All that having been said, what can we surmise from these particular results?

- `segmentj` performs better than `stdlib` at all encoding tasks.
- `jsoncolor` performs better than `segmentj` for indentation (which makes sense, as indentation is performed inline).
- `jsoncolor` performs better than `stdlib` at all encoding tasks.

Again, trust these benchmarks at your peril. Create your own benchmarks for your own workload.

## Notes

- The [`.golangci.yml`](./.golangci.yml) linter settings have been fiddled with to hush some
  linting issues inherited from the `segmentio` codebase at the time of forking. Thus, the linter report
  may not be of great use. In an ideal world, the `jsoncolor` functionality would be [ported](https://github.com/neilotoole/jsoncolor/issues/15) to a
  more recent (and better-linted) version of the `segementio` codebase.
- The `segmentio` encoder (at least as of `v0.1.14`) encodes `time.Duration` as string, while `stdlib` outputs as `int64`.
  This package follows `stdlib`.
- The [`Colors.Punc`](https://pkg.go.dev/github.com/neilotoole/jsoncolor#Colors) field is the
  fallback color for punctuation (`[]{},:`). As of `v0.9.0` the individual classes can also be set
  via `Colors.Brackets`, `Colors.Braces`, `Colors.Comma`, and `Colors.Colon`, each falling back to
  `Colors.Punc` when unset. (The structural `"` is colored by `Colors.String`/`Colors.Key`, not
  `Colors.Punc`.)

## Contributing

Contributions are welcome! See [`CONTRIBUTING.md`](CONTRIBUTING.md) for how to build,
test, and submit changes, and please follow the [Code of Conduct](CODE_OF_CONDUCT.md).
To report a security vulnerability, see [`SECURITY.md`](SECURITY.md).

<a name="history"></a>
## CHANGELOG

History: this package started as an extract of [`sq`](https://github.com/neilotoole/sq)'s JSON
encoding package, which itself was a fork of the
[`segmentio/encoding`](https://github.com/segmentio/encoding) JSON encoding package. Note that the
original `sq` JSON encoder was forked from Segment's codebase at `v0.1.14`, so
the codebases have drifted significantly by now.

### [v0.10.0](https://github.com/neilotoole/jsoncolor/releases/tag/v0.10.0)

Encoder correctness fixes, a fast path for uncolored output, a performance
pass against the segmentio upstream, and a stricter CI test run. There are no
exported API changes. The minor version bump signals that several of the
fixes change output at the byte level, so stored fixtures may need
regenerating; in particular, colored output from a partially populated
palette changes (see [#54](https://github.com/neilotoole/jsoncolor/issues/54)),
though nothing visible changes on a terminal.

- [#73](https://github.com/neilotoole/jsoncolor/pull/73): Backspace and form feed in strings are now written as `\b` and `\f` instead of `\u0008` and `\u000c`. This matches `encoding/json` since [Go 1.22](https://go.dev/doc/go1.22#encoding/json), which made the same change; this package had kept the older spelling. Both forms decode to the same bytes, so nothing changes semantically, but output containing either character differs byte-for-byte from earlier releases. Anything that compares encoder output against stored fixtures, golden files or checksums may need those regenerated.
- [#44](https://github.com/neilotoole/jsoncolor/issues/44): Fast-path the encode walk when no colors and no indenter are set, skipping the per-token nil-receiver dispatch. Output is byte-identical to the general walk, enforced by a parity test that also covers error paths.
- [#53](https://github.com/neilotoole/jsoncolor/issues/53): `Colors.TextMarshaler` falls back to `Colors.String` when unset, the same way the granular punctuation fields fall back to `Colors.Punc`. Previously an unset field produced an uncolored token followed by a stray ANSI reset.
- [#54](https://github.com/neilotoole/jsoncolor/issues/54): An empty `Color` now emits neither a prefix nor a reset, as the `Color` docs always promised. `&Colors{}` produces output identical to a nil palette, and `DefaultColors()` no longer emits a reset after every punctuation mark.
- [#56](https://github.com/neilotoole/jsoncolor/issues/56): A nil embedded struct pointer no longer leaves a double or trailing comma behind, and an indented struct whose every field is omitted encodes as `{}`, matching `encoding/json`.
- [#58](https://github.com/neilotoole/jsoncolor/issues/58): A failed `Encode` no longer leaks indenter depth into later `Encode` calls on the same `Encoder`.
- [#59](https://github.com/neilotoole/jsoncolor/issues/59): `omitempty` on a field promoted through an embedded struct pointer is evaluated on the promoted field, not on the pointer word. Previously zero-valued fields were emitted, and with the embedded pointer as the last field the check could read past the end of the struct.
- [#62](https://github.com/neilotoole/jsoncolor/issues/62): `map[string]RawMessage` with an invalid value now returns an error when map keys are unsorted instead of silently emitting invalid JSON. All map branches now roll the output buffer back to its entry length on error.
- CI runs the test suite under the race detector as well, which enables `checkptr` and would have caught the out-of-bounds read fixed in #59.
- [#65](https://github.com/neilotoole/jsoncolor/pull/65): Updated dependencies: `mattn/go-colorable` v0.1.15, `golang.org/x/sys` v0.47.0, and `golang.org/x/term` v0.45.0 (plus test-only `stretchr/testify` v1.12.1).
- [#66](https://github.com/neilotoole/jsoncolor/issues/66): A trusted `RawMessage` that is not well-formed JSON no longer leaks indenter depth into later `Encode` calls. End of input inside a container is now reported as a syntax error internally, so the verbatim fallback starts from a clean depth.
- [#67](https://github.com/neilotoole/jsoncolor/issues/67): A nil `[]byte` is now a single null token colored by `Colors.Null`, instead of a null nested inside the `Colors.Bytes` prefix with two resets.
- [#70](https://github.com/neilotoole/jsoncolor/issues/70): Fields promoted through an embedded struct pointer are skipped before anything is written when the pointer is nil, replacing the write-then-truncate rollback. No behavior change.
- [#73](https://github.com/neilotoole/jsoncolor/pull/73): Encoder performance. Strings are scanned for escapes eight bytes at a time and copied whole when clean; integers use a dedicated base-10 formatter; `map[string]string`, `map[string]bool` and `map[string][]string` no longer go through reflect and encode with zero allocations; struct member prefixes are precomputed; and the colorized walk decides indentation once per container. A per-shape head-to-head benchmark against the segmentio upstream, `BenchmarkCmp`, is included.

Taken together, the changes in this release alter encode time against v0.9.1 as follows, measured with `BenchmarkCmp` on an Apple M1 Max (Go 1.26, interleaved runs, benchstat). The colorless and colored columns use a nil palette and `DefaultColors()` respectively. The segmentio upstream reference cells did not move.

| Shape                                | Colorless              | Colored |
|--------------------------------------|------------------------|---------|
| `[]string`, 43-byte ASCII            | -65%                   | -64%    |
| `map[string]string`                  | -61% (101 allocs to 0) | -59%    |
| `[]int`                              | -24%                   | -22%    |
| `[]struct`                           | -21%                   | -25%    |
| `[]string`, short                    | -13%                   | -21%    |
| `[][]any` rows of mixed values       | -14%                   | -19%    |
| Decoded Sakila JSON document (`any`) | -15%                   | -19%    |
| `map[string]any`                     | -9%                    | -14%    |
| `[]string` with escapes              | -7%                    | -10%    |
| `[]float64`                          | -3%                    | -5%     |
| `[]string`, non-ASCII                | no change              | -6%     |

`BenchmarkCodeEncoder`, the large struct document from the Go standard library's tests, is 25% faster. Across all encode benchmarks the geomean improvement is 18%.

### [v0.9.1](https://github.com/neilotoole/jsoncolor/releases/tag/v0.9.1)

Documentation and repository housekeeping; no functional changes to the library.

- Add a package doc comment so [pkg.go.dev](https://pkg.go.dev/github.com/neilotoole/jsoncolor) shows a package overview.
- Add `CONTRIBUTING.md`, a Contributor Covenant `CODE_OF_CONDUCT.md`, and GitHub issue and pull request templates.
- Refresh `SECURITY.md`: update supported versions and switch to private vulnerability reporting.
- Restrict the CI workflows' `GITHUB_TOKEN` to least-privilege (`contents: read`) permissions.
- README: add Contributing and License sections, and point the release badge at the changelog.

### [v0.9.0](https://github.com/neilotoole/jsoncolor/releases/tag/v0.9.0)

- [#16](https://github.com/neilotoole/jsoncolor/issues/16): Add individually-configurable punctuation color fields — `Colors.Brackets`, `Colors.Braces`, `Colors.Comma`, and `Colors.Colon` — each falling back to `Colors.Punc` when unset, so existing configs are unaffected.
- [#19](https://github.com/neilotoole/jsoncolor/issues/19): Fix nondeterministic object key order when encoding `RawMessage`. Values are now re-encoded via on-the-fly tokenization, preserving source key order, instead of round-tripping through a `map`.
- [#37](https://github.com/neilotoole/jsoncolor/issues/37): Exported the `indenter` type (now `Indenter`) and added the `NewIndenter` constructor, so external callers can construct the indenter argument accepted by `Append`. The `Append` signature now reads `Append(b []byte, x interface{}, flags AppendFlags, clrs *Colors, indentr *Indenter)`.

### [v0.8.0](https://github.com/neilotoole/jsoncolor/releases/tag/v0.8.0)

- Bumped minimum Go version from 1.17 to 1.25.
- Updated dependencies to latest: `fatih/color` v1.19.0, `mattn/go-colorable` v0.1.14, `golang.org/x/sys` v0.45.0, and `golang.org/x/term` v0.43.0 (plus test-only `stretchr/testify` v1.11.1 and `segmentio/encoding` v0.5.4).
- Migrated the `golangci-lint` config to the v2 format; CI now runs `golangci-lint` v2.12.2 (action `v8`) and the test matrix targets Go 1.25 and 1.26.
- Modernized internal code to satisfy the updated linters (e.g. `reflect.Ptr`→`reflect.Pointer`, `io/ioutil`→`io`, `unsafe.Slice`/`unsafe.StringData` for string/byte conversion). No behavior change.

### [v0.7.2](https://github.com/neilotoole/jsoncolor/releases/tag/v0.7.2)

- [#38](https://github.com/neilotoole/jsoncolor/issues/38): Fix `TestCodec` failure on Go 1.22+ and update CI.
  - Use semantic JSON comparison in `TestCodec` to handle stdlib escape sequence changes.
  - Bump minimum Go version from 1.16 to 1.17.
  - Update CI workflows: expand test matrix to Go 1.17/1.24/1.26, fix `golangci-lint` workflow.

### [v0.7.1](https://github.com/neilotoole/jsoncolor/releases/tag/v0.7.1)

- [#27](https://github.com/neilotoole/jsoncolor/pull/27): Improved Windows terminal color support checking.

### [v0.7.0](https://github.com/neilotoole/jsoncolor/releases/tag/v0.7.0)

- [#21](https://github.com/neilotoole/jsoncolor/pull/21): Support for [`encoding.TextMarshaler`](https://pkg.go.dev/encoding#TextMarshaler).
- [#22](https://github.com/neilotoole/jsoncolor/pull/22): Removed redundant dependencies.
- [#26](https://github.com/neilotoole/jsoncolor/pull/26): Updated dependencies.

## Acknowledgments

- [`jq`](https://stedolan.github.io/jq/): sine qua non.
- [`segmentio/encoding`](https://github.com/segmentio/encoding): `jsoncolor` is layered into Segment's JSON encoder. They did the hard work. Much gratitude to that team.
- [`sq`](https://github.com/neilotoole/sq): `jsoncolor` is effectively an extract of code created specifically for `sq`.
- [`mattn/go-colorable`](https://github.com/mattn/go-colorable): no project is complete without `mattn` having played a role.
- [`fatih/color`](https://github.com/fatih/color): the color library.
- [`@hermannm`](https://github.com/hermannm): for several PRs.

### Related

- [`nwidger/jsoncolor`](https://github.com/nwidger/jsoncolor)
- [`hokaccha/go-prettyjson`](https://github.com/hokaccha/go-prettyjson)
- [`TylerBrock/colorjson`](https://github.com/TylerBrock/colorjson)

## License

[MIT](LICENSE) © Neil O'Toole
