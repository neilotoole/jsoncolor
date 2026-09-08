package jsoncolor_test

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neilotoole/jsoncolor"
)

// mapStringValues returns string-keyed maps of string, bool and []string
// values, including nil and empty maps, nil and empty slice values, and keys
// that need escaping.
func mapStringValues() []interface{} {
	return []interface{}{
		map[string]string(nil),
		map[string]string{},
		map[string]string{"a": "x"},
		map[string]string{"b": "2", "a": "1", "c": "3", "k\"q": "v\n", "<h>": "&"},
		map[string]bool(nil),
		map[string]bool{},
		map[string]bool{"t": true, "f": false, "m": true},
		map[string][]string(nil),
		map[string][]string{},
		map[string][]string{"empty": {}, "nil": nil, "two": {"x", "y"}, "one": {"z"}},
		struct {
			M map[string]string
			B map[string]bool
			S map[string][]string
		}{M: map[string]string{"k": "v"}, B: map[string]bool{"k": true}, S: map[string][]string{"k": {"a", "b"}}},
	}
}

// TestEncode_MapStringValues verifies that string-keyed maps of string, bool
// and []string encode exactly as encoding/json does, compact and indented,
// with sorted keys, and that unsorted output contains the same members.
func TestEncode_MapStringValues(t *testing.T) {
	for i, v := range mapStringValues() {
		t.Run(fmt.Sprintf("%d_%T", i, v), func(t *testing.T) {
			want, err := stdjson.Marshal(v)
			require.NoError(t, err)
			got, err := jsoncolor.Marshal(v)
			require.NoError(t, err)
			require.Equal(t, string(want), string(got))

			want, err = stdjson.MarshalIndent(v, "", "  ")
			require.NoError(t, err)
			got, err = jsoncolor.MarshalIndent(v, "", "  ")
			require.NoError(t, err)
			require.Equal(t, string(want), string(got))

			// Unsorted: the member order is not defined, but the output must
			// be valid JSON that decodes to the same value.
			buf := &bytes.Buffer{}
			enc := jsoncolor.NewEncoder(buf)
			require.NoError(t, enc.Encode(v))
			var back, wantBack interface{}
			require.NoError(t, stdjson.Unmarshal(buf.Bytes(), &back))
			require.NoError(t, stdjson.Unmarshal(want, &wantBack))
			require.Equal(t, wantBack, back)
		})
	}
}

// TestEncode_MapStringValues_NoAlloc verifies that encoding these maps into
// a buffer with spare capacity allocates nothing, sorted or not.
func TestEncode_MapStringValues_NoAlloc(t *testing.T) {
	values := []interface{}{
		map[string]string{"b": "2", "a": "1", "c": "3"},
		map[string]bool{"t": true, "f": false},
		map[string][]string{"two": {"x", "y"}, "one": {"z"}},
	}
	buf := make([]byte, 0, 4096)
	for _, v := range values {
		for _, flags := range []jsoncolor.AppendFlags{0, jsoncolor.SortMapKeys} {
			allocs := testing.AllocsPerRun(100, func() {
				var err error
				if buf, err = jsoncolor.Append(buf[:0], v, flags, nil, nil); err != nil {
					t.Fatal(err)
				}
			})
			require.Zero(t, allocs, "%T flags=%v", v, flags)
		}
	}
}
