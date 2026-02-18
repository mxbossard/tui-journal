package encoder

import (
	_ "encoding/binary"
	_ "fmt"
	"testing"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoder_AsciiEncoder(t *testing.T) {
	expectedStateSize := 10
	expectedDataSize := 100
	expectedWordSize := 8 + expectedStateSize + expectedDataSize

	// Encode & Decode with AsciiEncoder
	e1 := NewAsciiEncoder(0, expectedStateSize, expectedDataSize)
	assert.NotNil(t, e1)
	assert.Equal(t, expectedWordSize, e1.wordSize())
	header := e1.Header()
	assert.NotNil(t, header)
	assert.Len(t, header, expectedWordSize)

	m := e1.Match([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.False(t, m)
	m = e1.Match(header)
	assert.True(t, m)

	expectedSeq := 3
	expectedState := model.BuildState(expectedStateSize, "abcdefg")
	expectedText := "foobarbaz"

	buf, err := e1.Encode(expectedSeq, expectedState, expectedText)
	assert.NoError(t, err)
	require.NotNil(t, buf)
	assert.Len(t, buf, expectedWordSize)

	_, _, _, err = e1.Decode([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.Error(t, err)

	seq, s, text, err := e1.Decode(buf)
	assert.NoError(t, err)
	assert.Equal(t, expectedSeq, seq)
	assert.Equal(t, expectedState, s)
	assert.Equal(t, expectedText, text)

	// Setup & Decode with a new AsciiEncoder
	e2 := NewAsciiEncoder(0, 0, 0)
	_, _, _, err = e2.Decode(buf)
	assert.Error(t, err) // Not configured error

	ok := e2.Match([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.False(t, ok)
	err = e2.Setup([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.Error(t, err)

	ok = e2.Match(header)
	assert.True(t, ok)
	err = e2.Setup(header)
	assert.NoError(t, err)
	seq, s, text, err = e2.Decode(buf)
	assert.NoError(t, err)
	assert.Equal(t, expectedSeq, seq)
	assert.Equal(t, expectedState, s)
	assert.Equal(t, expectedText, text)

	// Iterate with AsciiEncoder
	expectedText1 := "foo"
	expectedText2 := "bar"
	expectedText3 := "baz"
	expectedState1 := model.BuildState(expectedStateSize, "pif")
	expectedState2 := model.BuildState(expectedStateSize, "paf")
	expectedState3 := model.BuildState(expectedStateSize, "pof")

	var bufs []byte
	e3 := NewAsciiEncoder(0, expectedStateSize, expectedDataSize)
	buf, err = e3.Encode(0, expectedState1, expectedText1)
	assert.NoError(t, err)
	bufs = append(bufs, buf...)
	buf, err = e3.Encode(1, expectedState2, expectedText2)
	assert.NoError(t, err)
	bufs = append(bufs, buf...)
	buf, err = e3.Encode(2, expectedState3, expectedText3)
	assert.NoError(t, err)
	bufs = append(bufs, buf...)

	k := 0
	e3.DecodeAll(model.TopToBottom, bufs, func(seq int, s model.State, text string, err error) {
		assert.Equal(t, k, seq)
		assert.NoError(t, err)
		switch k {
		case 0:
			assert.Equal(t, expectedState1, s)
			assert.Equal(t, expectedText1, text)
		case 1:
			assert.Equal(t, expectedState2, s)
			assert.Equal(t, expectedText2, text)
		case 2:
			assert.Equal(t, expectedState3, s)
			assert.Equal(t, expectedText3, text)

		}
		k++
	})

	k = 0
	e3.DecodeAll(model.BottomToTop, bufs, func(seq int, s model.State, text string, err error) {
		assert.Equal(t, 2-k, seq)
		assert.NoError(t, err)
		switch k {
		case 0:
			assert.Equal(t, expectedState3, s)
			assert.Equal(t, expectedText3, text)
		case 1:
			assert.Equal(t, expectedState2, s)
			assert.Equal(t, expectedText2, text)
		case 2:
			assert.Equal(t, expectedState1, s)
			assert.Equal(t, expectedText1, text)

		}
		k++
	})

	// Decode last Word
	seq, s, text, err = e3.DecodeLastWord(bufs)
	assert.NoError(t, err)
	assert.Equal(t, 2, seq)
	assert.Equal(t, expectedState3, s)
	assert.Equal(t, expectedText3, text)
}
