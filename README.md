# jsoncolor

Package `neilotoole/jsoncolor` provides a replacement for `encoding/json`
that can output colorized JSON.

## Usage

Get the package:

```shell
go get -u github.com/neilotoole/jsoncolor
```

Import it:

```go
import "github.com/neilotoole/jsoncolor"
```

Then to use it:

```go
func main() {
    var out io.Writer = os.Stdout
    var enc *jsoncolor.Encoder
	
    if jsoncolor.IsColorTerminal(out) {
        out = colorable.NewColorable(out)
        enc = jsoncolor.NewEncoder(out)
        enc.SetColors(jsoncolor.DefaultColors())
    } else {
        enc = jsoncolor.NewEncoder(out)
    }
    
    m := map[string]interface{}{
        "a": 1,
        "b": true,
        "c": "hello",
    }
    
    if err := enc.Encode(m); err != nil {
        panic(err)
    }	
}
```

## `jc`

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
encoding package.

Much gratitude to the Segment team for the superb work they put in on that package.
