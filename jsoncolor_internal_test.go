package jsoncolor

import (
	"bytes"
	stdjson "encoding/json"
	"testing"

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
