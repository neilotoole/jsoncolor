package jsoncolor_test

import (
	stdj "encoding/json"
	"io/ioutil"
	"testing"
	"time"

	segmentj "github.com/segmentio/encoding/json"

	"github.com/neilotoole/jsoncolor"
)

func makeBenchRecs() [][]interface{} {
	const maxRecs = 20000
	recs := make([][]interface{}, 0, maxRecs)
	for i := 0; i < maxRecs; i++ {
		rec := []interface{}{
			1,
			2,
			"3.00",
			true,
			time.Unix(1631659220, 0),
		}
		recs = append(recs, rec)
	}

	return recs
}

// The following benchmarks compare the encoding performance
// of JSON encoders. These are:
//
// - stdj: the std lib json encoder
// - segmentj: the encoder by segment.io
// - jsoncolor: this fork of segmentj that supports color

func BenchmarkStdj(b *testing.B) {
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

func BenchmarkStdj_Indent(b *testing.B) {
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

func BenchmarkSegmentj(b *testing.B) {
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

func BenchmarkSegmentj_Indent(b *testing.B) {
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
func BenchmarkJSONColorEnc(b *testing.B) {
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

func BenchmarkJSONColor_Indent(b *testing.B) {
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
