package jsoncolor_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/segmentio/encoding/json"

	"github.com/neilotoole/jsoncolor"
	"github.com/stretchr/testify/require"

	stdjson "encoding/json"
)

// TestPackageDropIn checks that jsoncolor satisfies basic requirements
// to be a drop-in for encoding/json.
func TestPackageDropIn(t *testing.T) {
	// Verify encoding/json types exists
	var (
		_ = jsoncolor.Decoder{}
		_ = jsoncolor.Delim(0)
		_ = jsoncolor.Encoder{}
		_ = jsoncolor.InvalidUTF8Error{}
		_ = jsoncolor.InvalidUnmarshalError{}
		_ = jsoncolor.Marshaler(nil)
		_ = jsoncolor.MarshalerError{}
		_ = jsoncolor.Number("0")
		_ = jsoncolor.RawMessage{}
		_ = jsoncolor.SyntaxError{}
		_ = jsoncolor.Token(nil)
		_ = jsoncolor.UnmarshalFieldError{}
		_ = jsoncolor.UnmarshalTypeError{}
		_ = jsoncolor.Unmarshaler(nil)
		_ = jsoncolor.UnsupportedTypeError{}
		_ = jsoncolor.UnsupportedValueError{}
	)

	const prefix, indent = "", "  "

	testCases := []string{"testdata/sakila_actor.json", "testdata/sakila_payment.json"}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			b, readErr := ioutil.ReadFile(tc)
			require.NoError(t, readErr)

			// Test json.Valid equivalence
			var fv1, fv2 = json.Valid, jsoncolor.Valid
			require.Equal(t, fv1(b), fv2(b))

			// Test json.Unmarshal equivalence
			var fu1, fu2 = json.Unmarshal, jsoncolor.Unmarshal
			var m1, m2 interface{}
			err1, err2 := fu1(b, &m1), fu2(b, &m2)
			require.NoError(t, err1)
			require.NoError(t, err2)
			require.EqualValues(t, m1, m2)

			// Test json.Marshal equivalence
			var fm1, fm2 = json.Marshal, jsoncolor.Marshal
			gotMarshalB1, err1 := fm1(m1)
			require.NoError(t, err1)
			gotMarshalB2, err2 := fm2(m1)
			require.NoError(t, err2)
			require.Equal(t, gotMarshalB1, gotMarshalB2)

			// Test json.MarshalIndent equivalence
			var fmi1, fmi2 = json.MarshalIndent, jsoncolor.MarshalIndent
			gotMarshallIndentB1, err1 := fmi1(m1, prefix, indent)
			require.NoError(t, err1)
			gotMarshalIndentB2, err2 := fmi2(m1, prefix, indent)
			require.NoError(t, err2)
			require.Equal(t, gotMarshallIndentB1, gotMarshalIndentB2)

			// Test json.Compact equivalence
			var fc1, fc2 = json.Compact, jsoncolor.Compact
			var buf1, buf2 = &bytes.Buffer{}, &bytes.Buffer{}
			err1 = fc1(buf1, gotMarshallIndentB1)
			require.NoError(t, err1)
			err2 = fc2(buf2, gotMarshalIndentB2)
			require.NoError(t, err2)
			require.Equal(t, buf1.Bytes(), buf2.Bytes())
			// Double-check
			require.Equal(t, buf1.Bytes(), gotMarshalB1)
			require.Equal(t, buf2.Bytes(), gotMarshalB2)
			buf1.Reset()
			buf2.Reset()

			// Test json.Indent equivalence
			var fi1, fi2 = json.Indent, jsoncolor.Indent
			err1 = fi1(buf1, gotMarshalB1, prefix, indent)
			require.NoError(t, err1)
			err2 = fi2(buf2, gotMarshalB2, prefix, indent)
			require.NoError(t, err2)
			require.Equal(t, buf1.Bytes(), buf2.Bytes())
			buf1.Reset()
			buf2.Reset()

			// Test json.HTMLEscape equivalence
			var fh1, fh2 = json.HTMLEscape, jsoncolor.HTMLEscape
			fh1(buf1, gotMarshalB1)
			fh2(buf2, gotMarshalB2)
			require.Equal(t, buf1.Bytes(), buf2.Bytes())
		})
	}
}

func TestEncode(t *testing.T) {
	testCases := []struct {
		name    string
		pretty  bool
		color   bool
		sortMap bool
		v       interface{}
		want    string
	}{
		{name: "nil", pretty: false, v: nil, want: "null\n"},
		{name: "slice_empty", pretty: true, v: []int{}, want: "[]\n"},
		{name: "slice_1_pretty", pretty: true, v: []interface{}{1}, want: "[\n  1\n]\n"},
		{name: "slice_1_no_pretty", v: []interface{}{1}, want: "[1]\n"},
		{name: "slice_2_pretty", pretty: true, v: []interface{}{1, true}, want: "[\n  1,\n  true\n]\n"},
		{name: "slice_2_no_pretty", v: []interface{}{1, true}, want: "[1,true]\n"},
		{name: "map_int_empty", pretty: true, v: map[string]int{}, want: "{}\n"},
		{name: "map_interface_empty", pretty: true, v: map[string]interface{}{}, want: "{}\n"},
		{name: "map_interface_empty_sorted", pretty: true, sortMap: true, v: map[string]interface{}{}, want: "{}\n"},
		{name: "map_1_pretty", pretty: true, sortMap: true, v: map[string]interface{}{"one": 1}, want: "{\n  \"one\": 1\n}\n"},
		{name: "map_1_no_pretty", sortMap: true, v: map[string]interface{}{"one": 1}, want: "{\"one\":1}\n"},
		{name: "map_2_pretty", pretty: true, sortMap: true, v: map[string]interface{}{"one": 1, "two": 2}, want: "{\n  \"one\": 1,\n  \"two\": 2\n}\n"},
		{name: "map_2_no_pretty", sortMap: true, v: map[string]interface{}{"one": 1, "two": 2}, want: "{\"one\":1,\"two\":2}\n"},
		{name: "tinystruct", pretty: true, v: TinyStruct{FBool: true}, want: "{\n  \"f_bool\": true\n}\n"},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			enc := jsoncolor.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			enc.SetSortMapKeys(tc.sortMap)
			if tc.pretty {
				enc.SetIndent("", "  ")
			}
			if tc.color {
				clrs := jsoncolor.DefaultColors()
				enc.SetColors(clrs)
			}

			require.NoError(t, enc.Encode(tc.v))
			require.True(t, stdjson.Valid(buf.Bytes()))
			require.Equal(t, tc.want, buf.String())
		})
	}
}

func TestEncode_Slice(t *testing.T) {
	testCases := []struct {
		name   string
		pretty bool
		color  bool
		v      []interface{}
		want   string
	}{
		{name: "nil", pretty: true, v: nil, want: "null\n"},
		{name: "empty", pretty: true, v: []interface{}{}, want: "[]\n"},
		{name: "one", pretty: true, v: []interface{}{1}, want: "[\n  1\n]\n"},
		{name: "two", pretty: true, v: []interface{}{1, true}, want: "[\n  1,\n  true\n]\n"},
		{name: "three", pretty: true, v: []interface{}{1, true, "hello"}, want: "[\n  1,\n  true,\n  \"hello\"\n]\n"},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			enc := jsoncolor.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			if tc.pretty {
				enc.SetIndent("", "  ")
			}
			if tc.color {
				enc.SetColors(jsoncolor.DefaultColors())
			}

			require.NoError(t, enc.Encode(tc.v))
			require.True(t, stdjson.Valid(buf.Bytes()))
			require.Equal(t, tc.want, buf.String())
		})
	}
}

func TestEncode_SmallStruct(t *testing.T) {
	v := SmallStruct{
		FInt:   7,
		FSlice: []interface{}{64, true},
		FMap: map[string]interface{}{
			"m_float64": 64.64,
			"m_string":  "hello",
		},
		FTinyStruct: TinyStruct{FBool: true},
		FString:     "hello",
	}

	testCases := []struct {
		pretty bool
		color  bool
		want   string
	}{
		{pretty: false, color: false, want: "{\"f_int\":7,\"f_slice\":[64,true],\"f_map\":{\"m_float64\":64.64,\"m_string\":\"hello\"},\"f_tinystruct\":{\"f_bool\":true},\"f_string\":\"hello\"}\n"},
		{pretty: true, color: false, want: "{\n  \"f_int\": 7,\n  \"f_slice\": [\n    64,\n    true\n  ],\n  \"f_map\": {\n    \"m_float64\": 64.64,\n    \"m_string\": \"hello\"\n  },\n  \"f_tinystruct\": {\n    \"f_bool\": true\n  },\n  \"f_string\": \"hello\"\n}\n"},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(fmt.Sprintf("pretty_%v__color_%v", tc.pretty, tc.color), func(t *testing.T) {
			buf := &bytes.Buffer{}
			enc := jsoncolor.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			enc.SetSortMapKeys(true)

			if tc.pretty {
				enc.SetIndent("", "  ")
			}
			if tc.color {
				enc.SetColors(jsoncolor.DefaultColors())
			}

			require.NoError(t, enc.Encode(v))
			require.True(t, stdjson.Valid(buf.Bytes()))
			require.Equal(t, tc.want, buf.String())
		})
	}
}

func TestEncode_Map_Nested(t *testing.T) {
	v := map[string]interface{}{
		"m_bool1": true,
		"m_nest1": map[string]interface{}{
			"m_nest1_bool": true,
			"m_nest2": map[string]interface{}{
				"m_nest2_bool": true,
				"m_nest3": map[string]interface{}{
					"m_nest3_bool": true,
				},
			},
		},
		"m_string1": "hello",
	}

	testCases := []struct {
		pretty bool
		color  bool
		want   string
	}{
		{pretty: false, want: "{\"m_bool1\":true,\"m_nest1\":{\"m_nest1_bool\":true,\"m_nest2\":{\"m_nest2_bool\":true,\"m_nest3\":{\"m_nest3_bool\":true}}},\"m_string1\":\"hello\"}\n"},
		{pretty: true, want: "{\n  \"m_bool1\": true,\n  \"m_nest1\": {\n    \"m_nest1_bool\": true,\n    \"m_nest2\": {\n      \"m_nest2_bool\": true,\n      \"m_nest3\": {\n        \"m_nest3_bool\": true\n      }\n    }\n  },\n  \"m_string1\": \"hello\"\n}\n"},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(fmt.Sprintf("pretty_%v__color_%v", tc.pretty, tc.color), func(t *testing.T) {
			buf := &bytes.Buffer{}
			enc := jsoncolor.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			enc.SetSortMapKeys(true)
			if tc.pretty {
				enc.SetIndent("", "  ")
			}
			if tc.color {
				enc.SetColors(jsoncolor.DefaultColors())
			}

			require.NoError(t, enc.Encode(v))
			require.True(t, stdjson.Valid(buf.Bytes()))
			require.Equal(t, tc.want, buf.String())
		})
	}
}

// TestEncode_Map_StringNotInterface tests maps with a string key
// but the value type is not interface{}.
// For example, map[string]bool. This test is necessary because the
// encoder has a fast path for map[string]interface{}
func TestEncode_Map_StringNotInterface(t *testing.T) {
	testCases := []struct {
		pretty  bool
		color   bool
		sortMap bool
		v       map[string]bool
		want    string
	}{
		{pretty: false, sortMap: true, v: map[string]bool{}, want: "{}\n"},
		{pretty: false, sortMap: false, v: map[string]bool{}, want: "{}\n"},
		{pretty: true, sortMap: true, v: map[string]bool{}, want: "{}\n"},
		{pretty: true, sortMap: false, v: map[string]bool{}, want: "{}\n"},
		{pretty: false, sortMap: true, v: map[string]bool{"one": true}, want: "{\"one\":true}\n"},
		{pretty: false, sortMap: false, v: map[string]bool{"one": true}, want: "{\"one\":true}\n"},
		{pretty: false, sortMap: true, v: map[string]bool{"one": true, "two": false}, want: "{\"one\":true,\"two\":false}\n"},
		{pretty: true, sortMap: true, v: map[string]bool{"one": true, "two": false}, want: "{\n  \"one\": true,\n  \"two\": false\n}\n"},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(fmt.Sprintf("size_%d__pretty_%v__color_%v", len(tc.v), tc.pretty, tc.color), func(t *testing.T) {
			buf := &bytes.Buffer{}
			enc := jsoncolor.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			enc.SetSortMapKeys(tc.sortMap)
			if tc.pretty {
				enc.SetIndent("", "  ")
			}
			if tc.color {
				enc.SetColors(jsoncolor.DefaultColors())
			}

			require.NoError(t, enc.Encode(tc.v))
			require.True(t, stdjson.Valid(buf.Bytes()))
			require.Equal(t, tc.want, buf.String())
		})
	}
}

func TestEncode_RawMessage(t *testing.T) {
	type RawStruct struct {
		FString string               `json:"f_string"`
		FRaw    jsoncolor.RawMessage `json:"f_raw"`
	}

	raw := jsoncolor.RawMessage(`{"one":1,"two":2}`)

	testCases := []struct {
		name   string
		pretty bool
		color  bool
		v      interface{}
		want   string
	}{
		{name: "empty", pretty: false, v: jsoncolor.RawMessage(`{}`), want: "{}\n"},
		{name: "no_pretty", pretty: false, v: raw, want: "{\"one\":1,\"two\":2}\n"},
		{name: "pretty", pretty: true, v: raw, want: "{\n  \"one\": 1,\n  \"two\": 2\n}\n"},
		{name: "pretty_struct", pretty: true, v: RawStruct{FString: "hello", FRaw: raw}, want: "{\n  \"f_string\": \"hello\",\n  \"f_raw\": {\n    \"one\": 1,\n    \"two\": 2\n  }\n}\n"},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			enc := jsoncolor.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			enc.SetSortMapKeys(true)
			if tc.pretty {
				enc.SetIndent("", "  ")
			}
			if tc.color {
				enc.SetColors(jsoncolor.DefaultColors())
			}

			err := enc.Encode(tc.v)
			require.NoError(t, err)
			require.True(t, stdjson.Valid(buf.Bytes()))
			require.Equal(t, tc.want, buf.String())
		})
	}
}

// TestEncode_Map_StringNotInterface tests map[string]json.RawMessage.
// This test is necessary because the encoder has a fast path
// for map[string]interface{}.
func TestEncode_Map_StringRawMessage(t *testing.T) {
	t.Skipf(`Skipping due to intermittent behavior.
See: https://github.com/neilotoole/jsoncolor/issues/19`)
	raw := jsoncolor.RawMessage(`{"one":1,"two":2}`)

	testCases := []struct {
		pretty  bool
		color   bool
		sortMap bool
		v       map[string]jsoncolor.RawMessage
		want    string
	}{
		{pretty: false, sortMap: true, v: map[string]jsoncolor.RawMessage{}, want: "{}\n"},
		{pretty: false, sortMap: false, v: map[string]jsoncolor.RawMessage{}, want: "{}\n"},
		{pretty: true, sortMap: true, v: map[string]jsoncolor.RawMessage{}, want: "{}\n"},
		{pretty: true, sortMap: false, v: map[string]jsoncolor.RawMessage{}, want: "{}\n"},
		{pretty: false, sortMap: true, v: map[string]jsoncolor.RawMessage{"msg1": raw, "msg2": raw}, want: "{\"msg1\":{\"one\":1,\"two\":2},\"msg2\":{\"one\":1,\"two\":2}}\n"},
		{pretty: true, sortMap: true, v: map[string]jsoncolor.RawMessage{"msg1": raw, "msg2": raw}, want: "{\n  \"msg1\": {\n    \"one\": 1,\n    \"two\": 2\n  },\n  \"msg2\": {\n    \"one\": 1,\n    \"two\": 2\n  }\n}\n"},
		{pretty: true, sortMap: false, v: map[string]jsoncolor.RawMessage{"msg1": raw}, want: "{\n  \"msg1\": {\n    \"one\": 1,\n    \"two\": 2\n  }\n}\n"},
	}

	for _, tc := range testCases {
		tc := tc

		name := fmt.Sprintf("size_%d__pretty_%v__color_%v__sort_%v", len(tc.v), tc.pretty, tc.color, tc.sortMap)
		t.Run(name, func(t *testing.T) {

			buf := &bytes.Buffer{}
			enc := jsoncolor.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			enc.SetSortMapKeys(tc.sortMap)
			if tc.pretty {
				enc.SetIndent("", "  ")
			}
			if tc.color {
				enc.SetColors(jsoncolor.DefaultColors())
			}

			require.NoError(t, enc.Encode(tc.v))
			require.True(t, stdjson.Valid(buf.Bytes()))
			require.Equal(t, tc.want, buf.String())
		})
	}
}

func TestEncode_BigStruct(t *testing.T) {
	v := newBigStruct()

	testCases := []struct {
		pretty bool
		color  bool
		want   string
	}{
		{pretty: false, want: "{\"f_int\":-7,\"f_int8\":-8,\"f_int16\":-16,\"f_int32\":-32,\"f_int64\":-64,\"f_uint\":7,\"f_uint8\":8,\"f_uint16\":16,\"f_uint32\":32,\"f_uint64\":64,\"f_float32\":32.32,\"f_float64\":64.64,\"f_bool\":true,\"f_bytes\":\"aGVsbG8=\",\"f_nil\":null,\"f_string\":\"hello\",\"f_map\":{\"m_bool\":true,\"m_int64\":64,\"m_nil\":null,\"m_smallstruct\":{\"f_int\":7,\"f_slice\":[64,true],\"f_map\":{\"m_float64\":64.64,\"m_string\":\"hello\"},\"f_tinystruct\":{\"f_bool\":true},\"f_string\":\"hello\"},\"m_string\":\"hello\"},\"f_smallstruct\":{\"f_int\":7,\"f_slice\":[64,true],\"f_map\":{\"m_float64\":64.64,\"m_string\":\"hello\"},\"f_tinystruct\":{\"f_bool\":true},\"f_string\":\"hello\"},\"f_interface\":\"hello\",\"f_interfaces\":[64,\"hello\",true]}\n"},
		{pretty: true, want: "{\n  \"f_int\": -7,\n  \"f_int8\": -8,\n  \"f_int16\": -16,\n  \"f_int32\": -32,\n  \"f_int64\": -64,\n  \"f_uint\": 7,\n  \"f_uint8\": 8,\n  \"f_uint16\": 16,\n  \"f_uint32\": 32,\n  \"f_uint64\": 64,\n  \"f_float32\": 32.32,\n  \"f_float64\": 64.64,\n  \"f_bool\": true,\n  \"f_bytes\": \"aGVsbG8=\",\n  \"f_nil\": null,\n  \"f_string\": \"hello\",\n  \"f_map\": {\n    \"m_bool\": true,\n    \"m_int64\": 64,\n    \"m_nil\": null,\n    \"m_smallstruct\": {\n      \"f_int\": 7,\n      \"f_slice\": [\n        64,\n        true\n      ],\n      \"f_map\": {\n        \"m_float64\": 64.64,\n        \"m_string\": \"hello\"\n      },\n      \"f_tinystruct\": {\n        \"f_bool\": true\n      },\n      \"f_string\": \"hello\"\n    },\n    \"m_string\": \"hello\"\n  },\n  \"f_smallstruct\": {\n    \"f_int\": 7,\n    \"f_slice\": [\n      64,\n      true\n    ],\n    \"f_map\": {\n      \"m_float64\": 64.64,\n      \"m_string\": \"hello\"\n    },\n    \"f_tinystruct\": {\n      \"f_bool\": true\n    },\n    \"f_string\": \"hello\"\n  },\n  \"f_interface\": \"hello\",\n  \"f_interfaces\": [\n    64,\n    \"hello\",\n    true\n  ]\n}\n"},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(fmt.Sprintf("pretty_%v__color_%v", tc.pretty, tc.color), func(t *testing.T) {

			buf := &bytes.Buffer{}
			enc := jsoncolor.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			enc.SetSortMapKeys(true)
			if tc.pretty {
				enc.SetIndent("", "  ")
			}
			if tc.color {
				enc.SetColors(jsoncolor.DefaultColors())
			}

			require.NoError(t, enc.Encode(v))
			require.True(t, stdjson.Valid(buf.Bytes()))
			require.Equal(t, tc.want, buf.String())
		})
	}
}

// TestEncode_Map_Not_StringInterface tests map encoding where
// the map is not map[string]interface{} (for which the encoder
// has a fast path).
//
// NOTE: Currently the encoder is broken wrt colors enabled
//  for non-string map keys, though that is kinda JSON-illegal anyway.
func TestEncode_Map_Not_StringInterface(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := jsoncolor.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetSortMapKeys(true)
	enc.SetColors(jsoncolor.DefaultColors())
	enc.SetIndent("", "  ")

	v := map[int32]string{
		0: "zero",
		1: "one",
		2: "two",
	}

	require.NoError(t, enc.Encode(v))
	require.False(t, stdjson.Valid(buf.Bytes()),
		"expected to be invalid JSON because the encoder currently doesn't handle maps with non-string keys")
}

// BigStruct is a big test struct.
type BigStruct struct {
	FInt         int                    `json:"f_int"`
	FInt8        int8                   `json:"f_int8"`
	FInt16       int16                  `json:"f_int16"`
	FInt32       int32                  `json:"f_int32"`
	FInt64       int64                  `json:"f_int64"`
	FUint        uint                   `json:"f_uint"`
	FUint8       uint8                  `json:"f_uint8"`
	FUint16      uint16                 `json:"f_uint16"`
	FUint32      uint32                 `json:"f_uint32"`
	FUint64      uint64                 `json:"f_uint64"`
	FFloat32     float32                `json:"f_float32"`
	FFloat64     float64                `json:"f_float64"`
	FBool        bool                   `json:"f_bool"`
	FBytes       []byte                 `json:"f_bytes"`
	FNil         interface{}            `json:"f_nil"`
	FString      string                 `json:"f_string"`
	FMap         map[string]interface{} `json:"f_map"`
	FSmallStruct SmallStruct            `json:"f_smallstruct"`
	FInterface   interface{}            `json:"f_interface"`
	FInterfaces  []interface{}          `json:"f_interfaces"`
}

// SmallStruct is a small test struct.
type SmallStruct struct {
	FInt        int                    `json:"f_int"`
	FSlice      []interface{}          `json:"f_slice"`
	FMap        map[string]interface{} `json:"f_map"`
	FTinyStruct TinyStruct             `json:"f_tinystruct"`
	FString     string                 `json:"f_string"`
}

// Tiny Struct is a tiny test struct.
type TinyStruct struct {
	FBool bool `json:"f_bool"`
}

func newBigStruct() BigStruct {
	return BigStruct{
		FInt:     -7,
		FInt8:    -8,
		FInt16:   -16,
		FInt32:   -32,
		FInt64:   -64,
		FUint:    7,
		FUint8:   8,
		FUint16:  16,
		FUint32:  32,
		FUint64:  64,
		FFloat32: 32.32,
		FFloat64: 64.64,
		FBool:    true,
		FBytes:   []byte("hello"),
		FNil:     nil,
		FString:  "hello",
		FMap: map[string]interface{}{
			"m_int64":       int64(64),
			"m_string":      "hello",
			"m_bool":        true,
			"m_nil":         nil,
			"m_smallstruct": newSmallStruct(),
		},
		FSmallStruct: newSmallStruct(),
		FInterface:   interface{}("hello"),
		FInterfaces:  []interface{}{int64(64), "hello", true},
	}
}

func newSmallStruct() SmallStruct {
	return SmallStruct{
		FInt:   7,
		FSlice: []interface{}{64, true},
		FMap: map[string]interface{}{
			"m_float64": 64.64,
			"m_string":  "hello",
		},
		FTinyStruct: TinyStruct{FBool: true},
		FString:     "hello",
	}
}

func TestEquivalenceRecords(t *testing.T) {
	rec := makeRecords(t, 10000)[0]

	bufStdj := &bytes.Buffer{}
	err := stdjson.NewEncoder(bufStdj).Encode(rec)
	require.NoError(t, err)

	bufSegmentj := &bytes.Buffer{}
	err = json.NewEncoder(bufSegmentj).Encode(rec)
	require.NoError(t, err)
	require.NotEqual(t, bufStdj.String(), bufSegmentj.String(), "segmentj encodes time.Duration to string; stdlib does not")

	bufJ := &bytes.Buffer{}
	err = jsoncolor.NewEncoder(bufJ).Encode(rec)
	require.Equal(t, bufStdj.String(), bufJ.String())
}

// TextMarshaler implements encoding.TextMarshaler
type TextMarshaler struct {
	Text string
}

func (t TextMarshaler) MarshalText() ([]byte, error) {
	return []byte(t.Text), nil
}

func TestEncode_TextMarshaler(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := jsoncolor.NewEncoder(buf)
	enc.SetColors(&jsoncolor.Colors{
		TextMarshaler: jsoncolor.Color("\x1b[36m"),
	})

	text := TextMarshaler{Text: "example text"}

	require.NoError(t, enc.Encode(text))
	require.Equal(t, "\x1b[36m\"example text\"\x1b[0m\n", buf.String(),
		"expected TextMarshaler encoding to use Colors.TextMarshaler")
}
