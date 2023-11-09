package jsoncolor

import (
	"strings"
	"testing"
)

// Based on https://github.com/segmentio/encoding/blob/v0.1.14/ascii/valid_test.go
var testCases = [...]struct {
	valid      bool
	validPrint bool
	str        string
}{
	{valid: true, validPrint: true, str: ""},
	{valid: true, validPrint: true, str: "hello"},
	{valid: true, validPrint: true, str: "Hello World!"},
	{valid: true, validPrint: true, str: "Hello\"World!"},
	{valid: true, validPrint: true, str: "Hello\\World!"},
	{valid: true, validPrint: false, str: "Hello\nWorld!"},
	{valid: true, validPrint: false, str: "Hello\rWorld!"},
	{valid: true, validPrint: false, str: "Hello\tWorld!"},
	{valid: true, validPrint: false, str: "Hello\bWorld!"},
	{valid: true, validPrint: false, str: "Hello\fWorld!"},
	{valid: true, validPrint: true, str: "H~llo World!"},
	{valid: true, validPrint: true, str: "H~llo"},
	{valid: false, validPrint: false, str: "你好"},
	{valid: true, validPrint: true, str: "~"},
	{valid: false, validPrint: false, str: "\x80"},
	{valid: true, validPrint: false, str: "\x7F"},
	{valid: false, validPrint: false, str: "\xFF"},
	{valid: true, validPrint: true, str: "some kind of long string with only ascii characters."},
	{valid: false, validPrint: false, str: "some kind of long string with a non-ascii character at the end.\xff"},
	{valid: true, validPrint: true, str: strings.Repeat("1234567890", 1000)},
}

func TestAsciiValid(t *testing.T) {
	for _, tc := range testCases {
		t.Run(limit(tc.str), func(t *testing.T) {
			expect := tc.validPrint

			if valid := asciiValidPrint([]byte(tc.str)); expect != valid {
				t.Errorf("expected %t but got %t", expect, valid)
			}
		})
	}
}

func TestAsciiValidPrint(t *testing.T) {
	for _, tc := range testCases {
		t.Run(limit(tc.str), func(t *testing.T) {
			expect := tc.validPrint

			if valid := asciiValidPrint([]byte(tc.str)); expect != valid {
				t.Errorf("expected %t but got %t", expect, valid)
			}
		})
	}
}

func limit(s string) string {
	if len(s) > 17 {
		return s[:17] + "..."
	}
	return s
}
