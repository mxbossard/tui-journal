package idx

import (
	_ "encoding/binary"
	_ "fmt"
	"testing"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx/idxrepo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsciiEncoder(t *testing.T) {
	expectedStateSize := 10
	expectedKeySize := 4
	expectedDataSize := 100
	expectedWordSize := 12 + expectedStateSize + expectedKeySize + expectedDataSize

	// Encode & Decode with AsciiEncoder
	e1 := idxrepo.NewAsciiEncoder(0, expectedStateSize, expectedKeySize, expectedDataSize)
	assert.NotNil(t, e1)
	assert.Equal(t, expectedWordSize, e1.WordSize())
	header := e1.Header()
	assert.NotNil(t, header)
	assert.Len(t, header, expectedWordSize)

	m := e1.Match([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.False(t, m)
	m = e1.Match(header)
	assert.True(t, m)

	expectedSeq := 3
	expectedState := BuildStringState(expectedStateSize, "abcdefg")
	expectedTime := time.Now()
	key := []byte("key")
	expectedKey := append(key, 0)
	expectedText := []byte("foobarbaz")

	buf, err := e1.Encode(expectedSeq, expectedTime, expectedState, key, expectedText)
	assert.NoError(t, err)
	require.NotNil(t, buf)
	assert.Len(t, buf, expectedWordSize)

	_, _, _, _, _, err = e1.Decode([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.Error(t, err)

	seq, tim, s, k, text, err := e1.Decode(buf)
	assert.NoError(t, err)
	assert.Equal(t, expectedSeq, seq)
	assert.Equal(t, expectedState, s)
	assert.Equal(t, expectedTime.Unix()/86400, tim.Unix()/86400)
	assert.Equal(t, expectedKey, k)
	assert.Equal(t, expectedText, text)

	// Setup & Decode with a new AsciiEncoder
	e2 := idxrepo.NewAsciiEncoder(0, 0, 0, 0)
	_, _, _, _, _, err = e2.Decode(buf)
	assert.Error(t, err) // Not configured error

	ok := e2.Match([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.False(t, ok)
	err = e2.Setup([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.Error(t, err)

	ok = e2.Match(header)
	assert.True(t, ok)
	err = e2.Setup(header)
	assert.NoError(t, err)
	seq, tim, s, k, text, err = e2.Decode(buf)
	assert.NoError(t, err)
	assert.Equal(t, expectedSeq, seq)
	assert.Equal(t, expectedTime.Unix()/86400, tim.Unix()/86400)
	assert.Equal(t, expectedState, s)
	assert.Equal(t, expectedText, text)

	// Iterate with AsciiEncoder
	expectedState1 := BuildStringState(expectedStateSize, "pif")
	expectedTime1 := time.Now()
	expectedState2 := BuildStringState(expectedStateSize, "paf")
	expectedTime2 := time.Now()
	expectedState3 := BuildStringState(expectedStateSize, "pof")
	expectedTime3 := time.Now()
	key1 := []byte("k1")
	expectedKey1 := append(key1, 0, 0)
	key2 := []byte("k2")
	expectedKey2 := append(key2, 0, 0)
	key3 := []byte("k3")
	expectedKey3 := append(key3, 0, 0)
	expectedText1 := []byte("foo")
	expectedText2 := []byte("bar")
	expectedText3 := []byte("baz")

	var bufs []byte
	e3 := idxrepo.NewAsciiEncoder(0, expectedStateSize, expectedKeySize, expectedDataSize)
	buf, err = e3.Encode(0, expectedTime1, expectedState1, key1, expectedText1)
	assert.NoError(t, err)
	bufs = append(bufs, buf...)
	buf, err = e3.Encode(1, expectedTime2, expectedState2, key2, expectedText2)
	assert.NoError(t, err)
	bufs = append(bufs, buf...)
	buf, err = e3.Encode(2, expectedTime3, expectedState3, key3, expectedText3)
	assert.NoError(t, err)
	bufs = append(bufs, buf...)

	i := 0
	e3.DecodeAll(TopToBottom, bufs, func(seq int, tim time.Time, s State, key []byte, text []byte, err error) bool {
		assert.Equal(t, i, seq)
		assert.NoError(t, err)
		switch i {
		case 0:
			assert.Equal(t, expectedTime1.Unix()/86400, tim.Unix()/86400)
			assert.Equal(t, expectedState1, s)
			assert.Equal(t, expectedKey1, key)
			assert.Equal(t, expectedText1, text)
		case 1:
			assert.Equal(t, expectedTime2.Unix()/86400, tim.Unix()/86400)
			assert.Equal(t, expectedState2, s)
			assert.Equal(t, expectedKey2, key)
			assert.Equal(t, expectedText2, text)
		case 2:
			assert.Equal(t, expectedTime3.Unix()/86400, tim.Unix()/86400)
			assert.Equal(t, expectedState3, s)
			assert.Equal(t, expectedKey3, key)
			assert.Equal(t, expectedText3, text)

		}
		i++
		return true
	})

	i = 0
	e3.DecodeAll(BottomToTop, bufs, func(seq int, tim time.Time, s State, key []byte, text []byte, err error) bool {
		assert.Equal(t, 2-i, seq)
		assert.NoError(t, err)
		switch i {
		case 0:
			assert.Equal(t, expectedTime3.Unix()/86400, tim.Unix()/86400)
			assert.Equal(t, expectedState3, s)
			assert.Equal(t, expectedKey3, key)
			assert.Equal(t, expectedText3, text)
		case 1:

			assert.Equal(t, expectedTime2.Unix()/86400, tim.Unix()/86400)
			assert.Equal(t, expectedState2, s)
			assert.Equal(t, expectedKey2, key)
			assert.Equal(t, expectedText2, text)
		case 2:

			assert.Equal(t, expectedTime1.Unix()/86400, tim.Unix()/86400)
			assert.Equal(t, expectedState1, s)
			assert.Equal(t, expectedKey1, key)
			assert.Equal(t, expectedText1, text)

		}
		i++
		return true
	})

	// Decode last Word
	seq, tim, s, key, text, err = e3.DecodeLastWord(bufs)
	assert.NoError(t, err)
	assert.Equal(t, 2, seq)
	assert.Equal(t, expectedTime3.Unix()/86400, tim.Unix()/86400)
	assert.Equal(t, expectedState3, s)
	assert.Equal(t, expectedText3, text)
}
