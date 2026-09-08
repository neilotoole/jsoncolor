package jsoncolor_test

import (
	stdj "encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	segmentj "github.com/segmentio/encoding/json"

	"github.com/neilotoole/jsoncolor"
)

// cmpStruct is the struct shape for BenchmarkCmp: a mix of the common leaf
// types with one omitempty field.
type cmpStruct struct {
	ID       int64
	Name     string
	Email    string
	Active   bool
	Score    float64
	Count    int
	Tags     []string
	Created  time.Time
	Nickname string `json:",omitempty"`
}

// cmpShapes returns the values BenchmarkCmp encodes, one per shape: a
// decoded JSON document, integer and float slices, a slice of structs, rows
// of mixed interface values, string-keyed maps, and string slices that need
// no escaping, need escaping, are non-ASCII, or are very short.
func cmpShapes(tb testing.TB) []struct {
	name string
	v    interface{}
} {
	data, err := os.ReadFile("testdata/sakila_actor.json")
	if err != nil {
		tb.Fatal(err)
	}
	var sakila interface{}
	if err = stdj.Unmarshal(data, &sakila); err != nil {
		tb.Fatal(err)
	}

	ints := make([]int64, 1000)
	floats := make([]float64, 1000)
	for i := range ints {
		ints[i] = int64(i*7919-350000) * int64(i%13+1)
		floats[i] = float64(i)*3.14159/7 - 100
	}
	structs := make([]cmpStruct, 100)
	for i := range structs {
		structs[i] = cmpStruct{ID: int64(i), Name: "Alice Example", Email: "alice@example.com", Active: i%2 == 0,
			Score: float64(i) / 3, Count: i * 11, Tags: []string{"a", "bb", "ccc"}, Created: time.Unix(1631659220+int64(i), 0).UTC()}
	}
	rows := make([][]interface{}, 100)
	for i := range rows {
		rows[i] = []interface{}{int64(i), "hello world", 3.14, true, nil, time.Unix(1631659220, 0).UTC(), "goodbye world"}
	}
	mapSS := map[string]string{}
	mapSI := map[string]interface{}{}
	for i := 0; i < 50; i++ {
		k := "key" + strings.Repeat("x", i%5) + string(rune('a'+i%26))
		mapSS[k] = "value " + k
		mapSI[k] = i
	}
	strs := make([]string, 200)
	for i := range strs {
		strs[i] = "The quick brown fox jumps over the lazy dog"
	}
	escs := make([]string, 200)
	for i := range escs {
		escs[i] = "She said \"hi\"\n\tand <left>"
	}
	unis := make([]string, 200)
	for i := range unis {
		unis[i] = "naïve café — 日本語テキスト"
	}
	shorts := make([]string, 500)
	for i := range shorts {
		shorts[i] = "abc"
	}

	return []struct {
		name string
		v    interface{}
	}{
		{"sakila", sakila},
		{"ints", ints},
		{"floats", floats},
		{"structs", structs},
		{"rows", rows},
		{"mapSS", mapSS},
		{"mapSI", mapSI},
		{"strs", strs},
		{"strs_esc", escs},
		{"strs_uni", unis},
		{"strs_short", shorts},
	}
}

// BenchmarkCmp encodes one value of each shape with the segmentio/encoding
// upstream, with jsoncolor and no colors, with jsoncolor and an empty
// palette (the colorized walk with nothing to color), and with jsoncolor and
// DefaultColors. Reading the cells for one shape together separates the
// cost of the walk from the cost of the color bytes, and the upstream cell
// shows how far the fork has drifted. Each call appends into a reused
// buffer, so the numbers exclude output allocation.
func BenchmarkCmp(b *testing.B) {
	shapes := cmpShapes(b)
	clrs := jsoncolor.DefaultColors()
	for _, sh := range shapes {
		b.Run(sh.name+"/segmentj", func(b *testing.B) {
			var buf []byte
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var err error
				if buf, err = segmentj.Append(buf[:0], sh.v, segmentj.EscapeHTML); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(sh.name+"/jc_plain", func(b *testing.B) {
			var buf []byte
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var err error
				if buf, err = jsoncolor.Append(buf[:0], sh.v, jsoncolor.EscapeHTML, nil, nil); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(sh.name+"/jc_empty", func(b *testing.B) {
			var buf []byte
			empty := &jsoncolor.Colors{}
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var err error
				if buf, err = jsoncolor.Append(buf[:0], sh.v, jsoncolor.EscapeHTML, empty, nil); err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(sh.name+"/jc_color", func(b *testing.B) {
			var buf []byte
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var err error
				if buf, err = jsoncolor.Append(buf[:0], sh.v, jsoncolor.EscapeHTML, clrs, nil); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
