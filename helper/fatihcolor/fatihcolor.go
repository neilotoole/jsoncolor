// Package fatihcolor provides a bridge between fatih/color
// and neilotoole/jsoncolor's native colorization mechanism.
package fatihcolor

import (
	"bytes"
	"io"
	"os"

	"github.com/mattn/go-colorable"

	"github.com/fatih/color"
	"github.com/neilotoole/jsoncolor"
)

// Colors describes color and pretty-printing options.
type Colors struct {
	// Bool is the color for boolean values.
	Bool *color.Color

	// Bytes is the color for byte / binary values.
	Bytes *color.Color

	// Datetime is the color for time-related values.
	Datetime *color.Color

	// Null is the color for null.
	Null *color.Color

	// Number is the color for number values, including int,
	// float, decimal etc.
	Number *color.Color

	// String is the color for string values.
	String *color.Color

	// Key is the color for keys such as a JSON field name.
	Key *color.Color

	// Punc is the color for punctuation such as colons, braces, etc.
	// Frequently Punc will just be color.Bold.
	Punc *color.Color
}

// NewDefaultColors returns default Colors instance.
func NewDefaultColors() *Colors {
	return &Colors{
		Bool:     color.New(color.FgYellow),
		Bytes:    color.New(color.Faint),
		Datetime: color.New(color.FgGreen, color.Faint),
		Key:      color.New(color.FgBlue, color.Bold),
		Null:     color.New(color.Faint),
		Number:   color.New(color.FgCyan),
		String:   color.New(color.FgGreen),
		Punc:     color.New(color.Bold),
	}
}

// ToCoreColors converts clrs to a core jsoncolor.Colors instance.
func ToCoreColors(clrs *Colors) jsoncolor.Colors {
	return jsoncolor.Colors{
		Null:   ToCoreColor(clrs.Null),
		Bool:   ToCoreColor(clrs.Bool),
		Number: ToCoreColor(clrs.Number),
		String: ToCoreColor(clrs.String),
		Key:    ToCoreColor(clrs.Key),
		Bytes:  ToCoreColor(clrs.Bytes),
		Time:   ToCoreColor(clrs.Datetime),
		Punc:   ToCoreColor(clrs.Punc),
	}
}

// ToCoreColor creates a jsoncolor.Color instance from a fatih/color
// instance.
func ToCoreColor(c *color.Color) jsoncolor.Color {
	// Dirty conversion function ahead: print
	// a space using c, then grab the bytes printed
	// before and after the space, and those are the
	// bytes we need for the prefix and suffix.
	// There's definitely a better way of doing this, but
	// it works for now.

	if c == nil {
		return jsoncolor.Color{}
	}

	// Make a copy because the pkg-level color.NoColor could be false.
	c2 := *c
	c2.EnableColor()

	b := []byte(c2.Sprint(" "))
	i := bytes.IndexByte(b, ' ')
	if i <= 0 {
		return jsoncolor.Color{}
	}

	return jsoncolor.Color{Prefix: b[:i], Suffix: b[i+1:]}
}

// NewEncoder returns a jsoncolor.Encoder configured to output
// in color, if clrs is non-nil. If clrs is nil, a "plain" jsoncolor.Encoder
// is returned. Effectively, this function provides a bridge between
// github.com/fatih/color and jsoncolor.Colors.
//
// Note that if clrs is non-nil, the encoder may write to an io.Writer
// instance that is not the out arg (e.g. the encoder's io.Writer may
// be a decorated version of out).
//
// Note also that the returned encoder may need to be further customized
// by invoking SetIndent, SetEscapeHTML, SetSortMapKeys, SetTrustRawMessage,
// etc.
func NewEncoder(out io.Writer, clrs *Colors) *jsoncolor.Encoder {
	if clrs == nil {
		return jsoncolor.NewEncoder(out)
	}

	if !jsoncolor.IsColorTerminal(out) {
		out = colorable.NewNonColorable(out)
	} else {
		out = colorable.NewColorable(out.(*os.File))
	}

	enc := jsoncolor.NewEncoder(out)
	coreClrs := ToCoreColors(clrs)
	enc.SetColors(coreClrs)
	return enc
}
