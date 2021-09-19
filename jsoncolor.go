package jsoncolor

import (
	"io"
	"os"
	"strconv"

	"github.com/mattn/go-isatty"

	"golang.org/x/term"
)

// Colors specifies colorization of JSON output.
type Colors struct {
	// Null is the color for JSON nil.
	Null Color

	// Bool is the color for boolean values.
	Bool Color

	// Number is the color for number values.
	Number Color

	// String is the color for string values.
	String Color

	// Key is the color for JSON keys.
	Key Color

	// Bytes is the color for byte data.
	Bytes Color

	// Time is the color for datetime values.
	Time Color

	// Punc is the color for JSON punctuation.
	Punc Color
}

// AppendNull appends a colorized "null" to b.
func (c Colors) AppendNull(b []byte) []byte {
	b = append(b, c.Null.Prefix...)
	b = append(b, "null"...)
	return append(b, c.Null.Suffix...)
}

// AppendBool appends the colorized bool v to b.
func (c Colors) AppendBool(b []byte, v bool) []byte {
	b = append(b, c.Bool.Prefix...)

	if v {
		b = append(b, "true"...)
	} else {
		b = append(b, "false"...)
	}

	return append(b, c.Bool.Suffix...)
}

// AppendKey appends the colorized key v to b.
func (c Colors) AppendKey(b []byte, v []byte) []byte {
	b = append(b, c.Key.Prefix...)
	b = append(b, v...)
	return append(b, c.Key.Suffix...)
}

// AppendInt64 appends the colorized int64 v to b.
func (c Colors) AppendInt64(b []byte, v int64) []byte {
	b = append(b, c.Number.Prefix...)
	b = strconv.AppendInt(b, v, 10)
	return append(b, c.Number.Suffix...)
}

// AppendUint64 appends the colorized uint64 v to b.
func (c Colors) AppendUint64(b []byte, v uint64) []byte {
	b = append(b, c.Number.Prefix...)
	b = strconv.AppendUint(b, v, 10)
	return append(b, c.Number.Suffix...)
}

// AppendPunc appends the colorized punctuation mark v to b.
func (c Colors) AppendPunc(b []byte, v byte) []byte {
	b = append(b, c.Punc.Prefix...)
	b = append(b, v)
	return append(b, c.Punc.Suffix...)
}

// Color is used to render terminal colors. The Prefix
// value is written, then the actual value, then the suffix.
type Color struct {
	// Prefix is the terminal color code prefix to print before the value (may be empty).
	Prefix []byte

	// Suffix is the terminal color code suffix to print after the value (may be empty).
	Suffix []byte
}

// reset is the ANSI reset escape code
const reset = "\x1b[0m"

// DefaultColors returns the default Colors configuration.
// These colors attempt to follow jq's default colorization.
func DefaultColors() Colors {
	return Colors{
		Null:   Color{Prefix: []byte("\x1b[2m"), Suffix: []byte(reset)},
		Bool:   Color{Prefix: []byte("\x1b[1m"), Suffix: []byte(reset)},
		Number: Color{Prefix: []byte("\x1b[36m"), Suffix: []byte(reset)},
		String: Color{Prefix: []byte("\x1b[32m"), Suffix: []byte(reset)},
		Key:    Color{Prefix: []byte("\x1b[34;1m"), Suffix: []byte(reset)},
		Bytes:  Color{Prefix: []byte("\x1b[2m"), Suffix: []byte(reset)},
		Time:   Color{Prefix: []byte("\x1b[32;2m"), Suffix: []byte(reset)},
		Punc:   Color{Prefix: []byte("\x1b[1m"), Suffix: []byte(reset)},
	}
}

// IsColorTerminal returns true if w is a colorable terminal.
func IsColorTerminal(w io.Writer) bool {
	if w == nil {
		return false
	}

	if !isTerminal(w) {
		return false
	}

	if os.Getenv("TERM") == "dumb" {
		return false
	}

	f, ok := w.(*os.File)
	if !ok {
		return false
	}

	if isatty.IsCygwinTerminal(f.Fd()) {
		return false
	}

	return true
}

// isTerminal returns true if w is a terminal.
func isTerminal(w io.Writer) bool {
	switch v := w.(type) {
	case *os.File:
		return term.IsTerminal(int(v.Fd()))
	default:
		return false
	}
}
