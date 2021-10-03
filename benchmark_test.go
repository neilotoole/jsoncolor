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

func BenchmarkEncoder_Encode(b *testing.B) {
	benchmarks := []struct {
		name   string
		indent bool
		color  bool
		fn     newEncoderFunc
	}{
		{name: "stdlib_no_indent", fn: newEncStdlib},
	}

	for _, bm := range benchmarks {
		bm := bm
		b.Run(bm.name, func(b *testing.B) {

		})
	}

}

// The following benchmarks compare the encoding performance
// of JSON encoders. These are:
//
// - stdj: the std lib json encoder
// - segmentj: the encoder by segment.io
// - jsoncolor: this fork of segmentj that supports color
func Benchmark_stdlib_NoIndent(b *testing.B) {
	b.ReportAllocs()
	recs := makeBenchRecs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		enc := stdj.NewEncoder(ioutil.Discard)
		enc.SetEscapeHTML(false)

		for i := range recs {
			err := enc.Encode(recs[i])
			if err != nil {
				b.Error(err)
			}
		}
	}
}

func Benchmark_stdlib_Indent(b *testing.B) {
	b.ReportAllocs()
	recs := makeBenchRecs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		enc := stdj.NewEncoder(ioutil.Discard)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")

		for i := range recs {
			err := enc.Encode(recs[i])
			if err != nil {
				b.Error(err)
			}
		}
	}
}

func Benchmark_segmentj_NoIndent(b *testing.B) {
	b.ReportAllocs()
	recs := makeBenchRecs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		enc := segmentj.NewEncoder(ioutil.Discard)
		enc.SetEscapeHTML(false)

		for i := range recs {
			err := enc.Encode(recs[i])
			if err != nil {
				b.Error(err)
			}
		}
	}
}

func Benchmark_segmentj_Indent(b *testing.B) {
	b.ReportAllocs()
	recs := makeBenchRecs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		enc := segmentj.NewEncoder(ioutil.Discard)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")

		for i := range recs {
			err := enc.Encode(recs[i])
			if err != nil {
				b.Error(err)
			}
		}
	}
}

func Benchmark_neilotoolejsoncolor_NoIndent(b *testing.B) {
	b.ReportAllocs()
	recs := makeBenchRecs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		enc := jsoncolor.NewEncoder(ioutil.Discard)
		enc.SetEscapeHTML(false)

		for i := range recs {
			err := enc.Encode(recs[i])
			if err != nil {
				b.Error(err)
			}
		}
	}
}

func Benchmark_neilotoolejsoncolor_Indent(b *testing.B) {
	b.ReportAllocs()
	recs := makeBenchRecs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		enc := jsoncolor.NewEncoder(ioutil.Discard)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")

		for i := range recs {
			err := enc.Encode(recs[i])
			if err != nil {
				b.Error(err)
			}
		}
	}
}

func Benchmark_neilotoolejsoncolor_NoIndent_Color(b *testing.B) {
	b.ReportAllocs()
	recs := makeBenchRecs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		enc := jsoncolor.NewEncoder(ioutil.Discard)
		enc.SetEscapeHTML(false)
		enc.SetColors(jsoncolor.DefaultColors())

		for i := range recs {
			err := enc.Encode(recs[i])
			if err != nil {
				b.Error(err)
			}
		}
	}
}

func Benchmark_neilotoolejsoncolor_Indent_Color(b *testing.B) {
	b.ReportAllocs()
	recs := makeBenchRecs()
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		enc := jsoncolor.NewEncoder(ioutil.Discard)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		enc.SetColors(jsoncolor.DefaultColors())

		for i := range recs {
			err := enc.Encode(recs[i])
			if err != nil {
				b.Error(err)
			}
		}
	}
}

func Benchmark_nwidgerjsoncolor_Indent_Color(b *testing.B) {
	b.ReportAllocs()
	recs := makeBenchRecs()
	buf := &bytes.Buffer{}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		f := newNwidgerColorFormatter()
		enc := nwidgerj.NewEncoderWithFormatter(buf, f)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")

		for i := range recs {
			err := enc.Encode(recs[i])
			if err != nil {
				b.Error(err)
			}
		}
	}
}

func Benchmark_nwidgerjsoncolor_NoIndent_Color(b *testing.B) {
	b.ReportAllocs()
	recs := makeBenchRecs()
	buf := &bytes.Buffer{}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		f := newNwidgerColorFormatter()
		enc := nwidgerj.NewEncoderWithFormatter(buf, f)
		enc.SetEscapeHTML(false)
		for i := range recs {
			err := enc.Encode(recs[i])
			if err != nil {
				b.Error(err)
			}
		}
	}
}

func newNwidgerColorFormatter() *nwidgerj.Formatter {
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
	return f
}

func makeBenchRecs() [][]interface{} {
	const maxRecs = 20000
	recs := make([][]interface{}, 0, maxRecs)
	for i := 0; i < maxRecs; i++ {
		rec := []interface{}{
			1,
			3.14,
			"6.77",
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
	enc := stdj.NewEncoder(w)
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
