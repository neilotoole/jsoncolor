package jsoncolor_test

import (
	"bytes"
	stdjson "encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/neilotoole/jsoncolor"
)

// stringEscapeInputs returns strings that place every single byte value, and
// a selection of multibyte runes, at every position of an ASCII base string,
// so that escaping is exercised at every alignment an 8-byte scanner can see.
// Plain strings of every length up to 24 are included as well.
func stringEscapeInputs() []string {
	const base = "abcdefghijklmnopq" // 17 bytes: two full 8-byte chunks and a tail

	inserts := make([]string, 0, 256+8)
	for c := 0; c < 256; c++ {
		inserts = append(inserts, string([]byte{byte(c)}))
	}
	inserts = append(inserts,
		"é", "日", " ", " ", "😀", "\xc3", "\xe6\x97", "\xf0\x9f\x98",
	)

	var inputs []string
	for _, ins := range inserts {
		for pos := 0; pos <= len(base); pos++ {
			inputs = append(inputs, base[:pos]+ins+base[pos:])
		}
	}
	for n := 0; n <= 24; n++ {
		inputs = append(inputs, strings.Repeat("x", n))
	}
	inputs = append(inputs,
		`"`, `\`, "\"\"\"\"\"\"\"\"\"", "<<<<<<<<>>>>>>>>&&&&&&&&",
		"\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f",
		"tab\there\nnewline\rcr\bbs\fff",
	)
	return inputs
}

// TestEncode_StringEscapesMatchStdlib verifies that every string in
// stringEscapeInputs encodes byte-for-byte as encoding/json encodes it, both
// with HTML escaping (Marshal) and without (Append with no flags).
func TestEncode_StringEscapesMatchStdlib(t *testing.T) {
	for i, s := range stringEscapeInputs() {
		t.Run(fmt.Sprintf("%d_%q", i, s), func(t *testing.T) {
			want, err := stdjson.Marshal(s)
			require.NoError(t, err)
			got, err := jsoncolor.Marshal(s)
			require.NoError(t, err)
			require.Equal(t, string(want), string(got), "with EscapeHTML")

			buf := &bytes.Buffer{}
			enc := stdjson.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			require.NoError(t, enc.Encode(s))
			want = bytes.TrimSuffix(buf.Bytes(), []byte{'\n'})

			got, err = jsoncolor.Append(nil, s, 0, nil, nil)
			require.NoError(t, err)
			require.Equal(t, string(want), string(got), "without EscapeHTML")
		})
	}
}
