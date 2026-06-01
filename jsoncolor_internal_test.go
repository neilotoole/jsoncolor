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

// TestEncode_FastPathParity asserts that the colorless encode paths
// produce byte-identical output to the colorized walk for every value
// in testValues. Append (the public entry point) takes the fast path
// when clrs is nil; appendInternal with forceSlow=true forces the
// colorized walk regardless. The two outputs must be equal, otherwise
// the fast path has diverged.
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
			for _, v := range testValues {
				t.Run(testName(v), func(t *testing.T) {
					// Each call uses its own Indenter clone because the
					// Indenter carries per-encode depth state.
					var fastIndentr, slowIndentr *Indenter
					if cfg.indentr {
						fastIndentr = NewIndenter(cfg.prefix, cfg.indent)
						slowIndentr = NewIndenter(cfg.prefix, cfg.indent)
					}

					fast, fastErr := Append(nil, v, cfg.flags, nil, fastIndentr)
					slow, slowErr := appendInternal(nil, v, cfg.flags, nil, slowIndentr, true)

					if (fastErr == nil) != (slowErr == nil) {
						t.Fatalf("error mismatch: fast=%v slow=%v", fastErr, slowErr)
					}
					if fastErr != nil {
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
// colorized walk on identical data, with the output buffer pre-sized so
// growth is not measured. The only variable is whether the fast path is
// taken: fast=Append (forceSlow=false), slow=appendInternal with
// forceSlow=true. Any nonzero delta is the fast path's contribution.
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
					buf, _ = appendInternal(buf[:0], rec, EscapeHTML|SortMapKeys, nil, indentr, true)
				}
			})
		})
	}
}
