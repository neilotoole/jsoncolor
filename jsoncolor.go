package jsoncolor

import (
	"strconv"
)

// Colors specifies colorization of JSON output. Each field
// is a Color, which is simply the bytes of the terminal color code.
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

	// Punc is the color for JSON punctuation: the structural
	// characters [ ] { } , and : .
	//
	// Punc acts as the fallback color for all punctuation. The more
	// specific fields below (Brackets, Braces, Comma, Colon) override
	// Punc for their respective characters; when one of those fields is
	// the zero value (unset), the encoder falls back to Punc. Therefore
	// code that sets only Punc continues to colorize all punctuation
	// exactly as before.
	//
	// Note that the structural double-quote characters that delimit JSON
	// strings and keys are not controlled by Punc; they are colored by
	// the String and Key fields respectively.
	Punc Color

	// Brackets is the color for array bracket punctuation: [ and ] .
	// When unset (the zero value), it falls back to Punc.
	Brackets Color

	// Braces is the color for object brace punctuation: { and } .
	// When unset (the zero value), it falls back to Punc.
	Braces Color

	// Comma is the color for the comma punctuation that separates
	// array elements and object members: , .
	// When unset (the zero value), it falls back to Punc.
	Comma Color

	// Colon is the color for the colon punctuation that separates
	// object keys from their values: : .
	// When unset (the zero value), it falls back to Punc.
	Colon Color

	// TextMarshaler is the color for values implementing encoding.TextMarshaler.
	TextMarshaler Color
}

// appendNull appends a colorized "null" to b.
func (c *Colors) appendNull(b []byte) []byte {
	if c == nil {
		return append(b, "null"...)
	}

	b = append(b, c.Null...)
	b = append(b, "null"...)
	return append(b, ansiReset...)
}

// appendBool appends the colorized bool v to b.
func (c *Colors) appendBool(b []byte, v bool) []byte {
	if c == nil {
		if v {
			return append(b, "true"...)
		}

		return append(b, "false"...)
	}

	b = append(b, c.Bool...)
	if v {
		b = append(b, "true"...)
	} else {
		b = append(b, "false"...)
	}

	return append(b, ansiReset...)
}

// appendInt64 appends the colorized int64 v to b.
func (c *Colors) appendInt64(b []byte, v int64) []byte {
	if c == nil {
		return strconv.AppendInt(b, v, 10)
	}

	b = append(b, c.Number...)
	b = strconv.AppendInt(b, v, 10)
	return append(b, ansiReset...)
}

// appendUint64 appends the colorized uint64 v to b.
func (c *Colors) appendUint64(b []byte, v uint64) []byte {
	if c == nil {
		return strconv.AppendUint(b, v, 10)
	}

	b = append(b, c.Number...)
	b = strconv.AppendUint(b, v, 10)
	return append(b, ansiReset...)
}

// appendPunc appends the colorized punctuation mark v to b. The color is
// chosen by puncColor, which selects the granular field corresponding to v
// (Brackets, Braces, Comma, Colon), falling back to Punc when that field is
// unset.
func (c *Colors) appendPunc(b []byte, v byte) []byte {
	if c == nil {
		return append(b, v)
	}

	b = append(b, c.puncColor(v)...)
	b = append(b, v)
	return append(b, ansiReset...)
}

// puncColor returns the Color to use for punctuation mark v. It selects the
// granular field that governs v: Brackets for [ and ], Braces for { and },
// Comma for , and Colon for : . When the selected granular field is the zero
// value (unset), it falls back to Punc. This preserves backward compatibility
// for callers that set only Punc.
func (c *Colors) puncColor(v byte) Color {
	var clr Color
	switch v {
	case '[', ']':
		clr = c.Brackets
	case '{', '}':
		clr = c.Braces
	case ',':
		clr = c.Comma
	case ':':
		clr = c.Colon
	}

	if len(clr) == 0 {
		return c.Punc
	}
	return clr
}

// Color is used to render terminal colors. In effect, Color is
// the bytes of the ANSI prefix code. The zero value is valid (results in
// no colorization). When Color is non-zero, the encoder writes the prefix,
// then the actual value, then the ANSI reset code.
//
// Example value:
//
//	number := Color("\x1b[36m")
type Color []byte

// ansiReset is the ANSI ansiReset escape code.
const ansiReset = "\x1b[0m"

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
