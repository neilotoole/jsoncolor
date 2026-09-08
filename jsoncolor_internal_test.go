package jsoncolor

import (
	"bytes"
	stdjson "encoding/json"
	"testing"
	"time"

	"github.com/segmentio/encoding/json"

	"github.com/stretchr/testify/require"
)

func TestEquivalenceStdlibCode(t *testing.T) {
	if codeJSON == nil {
		codeInit()
	}

	bufStdj := &bytes.Buffer{}
	err := stdjson.NewEncoder(bufStdj).Encode(codeStruct)
	require.NoError(t, err)

	bufSegmentj := &bytes.Buffer{}
	err = json.NewEncoder(bufSegmentj).Encode(codeStruct)
	require.NoError(t, err)
	require.Equal(t, bufStdj.String(), bufSegmentj.String())

	bufJ := &bytes.Buffer{}
	err = NewEncoder(bufJ).Encode(codeStruct)
	require.Equal(t, bufStdj.String(), bufJ.String())
}

// Types for parityExtraValues.
type parityInner struct {
	X int    `json:",omitempty"`
	Y string `json:",omitempty"`
}

type parityNilEmbedOnly struct{ *parityInner }

type parityNilEmbedMiddle struct {
	A int
	*parityInner
	C int
}

type parityNilEmbedLast struct {
	A int
	C int
	*parityInner
}

type parityAllOmitEmpty struct {
	A int    `json:",omitempty"`
	B string `json:",omitempty"`
}

type parityFuncField struct{ F func() }

// parityExtraValues supplements the vendored testValues with shapes that
// exercise nil embedded struct pointers, omitempty through embedded pointers, empty
// indented structs, and encode errors, on both the fast and slow paths.
var parityExtraValues = []interface{}{
	parityNilEmbedOnly{},
	parityNilEmbedMiddle{A: 1, C: 2},
	parityNilEmbedLast{A: 1, C: 2},
	parityNilEmbedMiddle{A: 1, parityInner: &parityInner{}, C: 2},
	parityNilEmbedMiddle{A: 1, parityInner: &parityInner{X: 5, Y: "y"}, C: 2},
	parityAllOmitEmpty{},
	struct{ In parityAllOmitEmpty }{},
	parityFuncField{},
	[]func(){nil},
	map[string]func(){"a": nil},
	[]interface{}{1, parityFuncField{}},
}

// slowColors is an empty, non-nil palette. The encoder takes the fast
// paths only when clrs is nil, so passing slowColors routes the same
// colorless encode through the general walk; since every Color is empty,
// no ANSI escapes are written and the output must be byte-identical.
var slowColors = &Colors{}

// TestEncode_FastPathParity asserts that the colorless fast paths
// produce byte-identical output to the general (slow) walk for every
// value in testValues. Append takes the fast path when clrs is nil;
// slowColors forces the general walk. Neither side emits color. The two
// outputs must be equal, otherwise the fast path has diverged.
//
// SortMapKeys is included in every config because map iteration is
// otherwise nondeterministic across the two calls — without sorting,
// fast and slow would see different orderings of the same map and
// diverge harmlessly.
func TestEncode_FastPathParity(t *testing.T) {
	prefix, indent2 := "", "  "
	tab := "\t"
	configs := []struct {
		name    string
		flags   AppendFlags
		prefix  string
		indent  string
		indentr bool
	}{
		{name: "flat", flags: SortMapKeys},
		{name: "flat_escapeHTML", flags: EscapeHTML | SortMapKeys},
		{name: "indent_2sp", flags: SortMapKeys, indentr: true, prefix: prefix, indent: indent2},
		{name: "indent_2sp_escapeHTML", flags: EscapeHTML | SortMapKeys, indentr: true, prefix: prefix, indent: indent2},
		{name: "indent_tab", flags: SortMapKeys, indentr: true, prefix: prefix, indent: tab},
		{name: "indent_with_prefix", flags: SortMapKeys, indentr: true, prefix: ">>", indent: indent2},
		{name: "indent_disabled", flags: SortMapKeys, indentr: true, prefix: "", indent: ""},
	}

	for _, cfg := range configs {
		t.Run(cfg.name, func(t *testing.T) {
			for _, v := range append(testValues[:], parityExtraValues...) {
				t.Run(testName(v), func(t *testing.T) {
					// Each call uses its own Indenter clone because the
					// Indenter carries per-encode depth state.
					var fastIndentr, slowIndentr *Indenter
					if cfg.indentr {
						fastIndentr = NewIndenter(cfg.prefix, cfg.indent)
						slowIndentr = NewIndenter(cfg.prefix, cfg.indent)
					}

					fast, fastErr := Append(nil, v, cfg.flags, nil, fastIndentr)
					slow, slowErr := Append(nil, v, cfg.flags, slowColors, slowIndentr)

					if (fastErr == nil) != (slowErr == nil) {
						t.Fatalf("error mismatch: fast=%v slow=%v", fastErr, slowErr)
					}
					if fastErr != nil {
						if fastErr.Error() != slowErr.Error() {
							t.Fatalf("error text mismatch: fast=%v slow=%v", fastErr, slowErr)
						}
						if !bytes.Equal(fast, slow) {
							t.Fatalf("error-path byte mismatch: fast=%q slow=%q", fast, slow)
						}
						return
					}

					if !bytes.Equal(fast, slow) {
						t.Errorf("fast/slow byte mismatch")
						t.Logf("fast (%d bytes): %q", len(fast), string(fast))
						t.Logf("slow (%d bytes): %q", len(slow), string(slow))
					}
				})
			}
		})
	}
}

// benchMixedRecord is a typed value used by BenchmarkFastVsSlow to
// isolate fast-path savings from interface boxing and buffer growth.
type benchMixedRecord struct {
	I   int
	I64 int64
	F32 float32
	F64 float64
	S   string
	B1  bool
	B2  bool
	T   time.Time
}

// BenchmarkFastVsSlow compares the colorless fast path against the
// general (slow) walk on identical data, with the output buffer pre-sized so
// growth is not measured. The only variable is whether the fast path is
// taken: fast=Append with nil colors, slow=Append with slowColors. Any
// nonzero delta is the fast path's contribution.
func BenchmarkFastVsSlow(b *testing.B) {
	rec := benchMixedRecord{
		I: 1, I64: 2, F32: 2.71, F64: 3.14,
		S: "hello world", B1: true, B2: false,
		T: time.Unix(1631659220, 0),
	}

	cases := []struct {
		name   string
		indent bool
	}{
		{name: "flat"},
		{name: "indent", indent: true},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			var indentr *Indenter
			if tc.indent {
				indentr = NewIndenter("", "  ")
			}
			buf := make([]byte, 0, 4096)

			b.Run("fast", func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					buf, _ = Append(buf[:0], rec, EscapeHTML|SortMapKeys, nil, indentr)
				}
			})
			b.Run("slow", func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					buf, _ = Append(buf[:0], rec, EscapeHTML|SortMapKeys, slowColors, indentr)
				}
			})
		})
	}
}

// TestEncode_UnsortedMapErrorRollsBack verifies that when an element of a
// map fails to encode with SortMapKeys unset, both the fast and slow paths
// roll the buffer back to its length on entry, matching the sorted branches
// and every other container. Single-key maps keep iteration deterministic.
func TestEncode_UnsortedMapErrorRollsBack(t *testing.T) {
	values := []interface{}{
		map[string]RawMessage{"a": RawMessage(`{bad`)},
		map[string]interface{}{"a": func() {}},
		map[string]int{"a": 1, "b": 2}, // control: must succeed on both paths
	}

	for _, indentr := range []*Indenter{nil, NewIndenter("", "  ")} {
		for _, v := range values {
			name := testName(v)
			if indentr != nil {
				name += "/indent"
			}
			t.Run(name, func(t *testing.T) {
				prefix := []byte("prefix")

				fast, fastErr := Append(append([]byte(nil), prefix...), v, 0, nil, indentr)
				slow, slowErr := Append(append([]byte(nil), prefix...), v, 0, slowColors, indentr)

				if (fastErr == nil) != (slowErr == nil) {
					t.Fatalf("error mismatch: fast=%v slow=%v", fastErr, slowErr)
				}
				if fastErr == nil {
					return
				}
				if fastErr.Error() != slowErr.Error() {
					t.Fatalf("error text mismatch: fast=%v slow=%v", fastErr, slowErr)
				}
				if string(fast) != string(prefix) {
					t.Errorf("fast path did not roll back: %q", fast)
				}
				if string(slow) != string(prefix) {
					t.Errorf("slow path did not roll back: %q", slow)
				}
			})
		}
	}
}
