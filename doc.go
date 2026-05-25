// Package jsoncolor is a drop-in replacement for the standard library's
// encoding/json that emits colorized (and optionally indented) JSON.
//
// It is aimed at command-line tools that want jq-style colored output.
// Colorization and indentation are performed inline in the encoder, which makes
// it considerably faster than indenting via the standard library's json.Indent.
//
// The API mirrors encoding/json: use [NewEncoder] with [Encoder.Encode], or the
// package-level [Marshal] and [MarshalIndent] functions, exactly as you would
// with the standard library. To enable color, construct a [Colors] value (or use
// [DefaultColors], which approximates jq's palette) and call [Encoder.SetColors].
// A nil or zero-value [Colors] disables colorization, so jsoncolor remains a
// faithful drop-in when color is not desired.
//
// jsoncolor is layered onto a fork of github.com/segmentio/encoding/json; see
// SEGMENTIO_README.md for the upstream package's documentation.
package jsoncolor
