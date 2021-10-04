package jsoncolor_test

import (
	"bytes"
	stdj "encoding/json"
	"io"
	"io/ioutil"
	"testing"
	"time"

	segmentj "github.com/segmentio/encoding/json"

	"github.com/neilotoole/jsoncolor"
	nwidgerj "github.com/nwidger/jsoncolor"
)

func BenchmarkEncode(b *testing.B) {
	recs := makeRecords(b, 10000)

	benchmarks := []struct {
		name   string
		indent bool
		color  bool
		fn     newEncoderFunc
	}{
		{name: "stdlib_NoIndent", fn: newEncStdlib},
		{name: "stdlib_Indent", fn: newEncStdlib, indent: true},
		{name: "segmentj_NoIndent", fn: newEncSegmentj},
		{name: "segmentj_Indent", fn: newEncSegmentj, indent: true},
		{name: "neilotoole_NoIndent_NoColor", fn: newEncNeilotoole},
		{name: "neilotoole_Indent_NoColor", fn: newEncNeilotoole, indent: true},
		{name: "neilotoole_NoIndent_Color", fn: newEncNeilotoole, color: true},
		{name: "neilotoole_Indent_Color", fn: newEncNeilotoole, indent: true, color: true},
		{name: "nwidger_NoIndent_NoColor", fn: newEncNwidger},
		{name: "nwidger_Indent_NoColor", fn: newEncNwidger, indent: true},
		{name: "nwidger_indent_NoIndent_Colo", fn: newEncNwidger, color: true},
		{name: "nwidger_indent_Indent_Color", fn: newEncNwidger, indent: true, color: true},
	}

	for _, bm := range benchmarks {
		bm := bm
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()

			b.ResetTimer()

			for n := 0; n < b.N; n++ {
				w := &bytes.Buffer{}
				enc := bm.fn(w, bm.indent, bm.color)

				for i := range recs {
					err := enc.Encode(recs[i])
					if err != nil {
						b.Error(err)
					}
				}
			}
		})
	}
}

func makeRecords(tb testing.TB, n int) [][]interface{} {
	recs := make([][]interface{}, 0, n)

	// add a bunch of data from a file, just to make the recs bigger
	data, err := ioutil.ReadFile("testdata/sakila_actor.json")
	if err != nil {
		tb.Fatal(err)
	}

	f := new(interface{})
	if err = stdj.Unmarshal(data, f); err != nil {
		tb.Fatal(err)
	}

	type someStruct struct {
		i int64
		a string
		f interface{} // x holds JSON loaded from file
	}

	for i := 0; i < n; i++ {
		rec := []interface{}{
			int(1),
			int64(2),
			float32(2.71),
			float64(3.14),
			"hello world",
			someStruct{i: 8, a: "goodbye world", f: f},
			map[string]interface{}{"a": 9, "b": "ca va"},
			true,
			false,
			time.Unix(1631659220, 0),
			time.Millisecond * 1631659220,
		}
		recs = append(recs, rec)
	}

	return recs
}

type newEncoderFunc func(w io.Writer, indent, color bool) encoder

var (
	_ newEncoderFunc = newEncStdlib
	_ newEncoderFunc = newEncSegmentj
	_ newEncoderFunc = newEncNeilotoole
	_ newEncoderFunc = newEncNwidger
)

type encoder interface {
	SetEscapeHTML(on bool)
	SetIndent(prefix, indent string)
	Encode(v interface{}) error
}

func newEncStdlib(w io.Writer, indent, color bool) encoder {
	enc := stdj.NewEncoder(w)
	if indent {
		enc.SetIndent("", "  ")
	}
	enc.SetEscapeHTML(true)
	return enc
}

func newEncSegmentj(w io.Writer, indent, color bool) encoder {
	enc := segmentj.NewEncoder(w)
	if indent {
		enc.SetIndent("", "  ")
	}
	enc.SetEscapeHTML(true)
	return enc
}

func newEncNeilotoole(w io.Writer, indent, color bool) encoder {
	enc := jsoncolor.NewEncoder(w)
	if indent {
		enc.SetIndent("", "  ")
	}
	enc.SetEscapeHTML(true)

	if color {
		clrs := jsoncolor.DefaultColors()
		enc.SetColors(clrs)
	}

	return enc
}

func newEncNwidger(w io.Writer, indent, color bool) encoder {
	if !color {
		enc := nwidgerj.NewEncoder(w)
		enc.SetEscapeHTML(false)
		if indent {
			enc.SetIndent("", "  ")
		}
		return enc
	}

	// It's color
	f := nwidgerj.NewFormatter()
	f.SpaceColor = nwidgerj.DefaultSpaceColor
	f.CommaColor = nwidgerj.DefaultCommaColor
	f.ColonColor = nwidgerj.DefaultColonColor
	f.ObjectColor = nwidgerj.DefaultObjectColor
	f.ArrayColor = nwidgerj.DefaultArrayColor
	f.FieldQuoteColor = nwidgerj.DefaultFieldQuoteColor
	f.FieldColor = nwidgerj.DefaultFieldColor
	f.StringQuoteColor = nwidgerj.DefaultStringQuoteColor
	f.StringColor = nwidgerj.DefaultStringColor
	f.TrueColor = nwidgerj.DefaultTrueColor
	f.FalseColor = nwidgerj.DefaultFalseColor
	f.NumberColor = nwidgerj.DefaultNumberColor
	f.NullColor = nwidgerj.DefaultNullColor

	enc := nwidgerj.NewEncoderWithFormatter(w, f)
	enc.SetEscapeHTML(false)

	if indent {
		enc.SetIndent("", "  ")
	}

	return enc
}
