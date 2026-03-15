package idx

import (
	_ "encoding/binary"
	_ "fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicEncoder_Header(t *testing.T) {
	expectedStateSize := 10
	expectedKeySize := 32
	expectedValSize := 100
	expectedWordSize := 12 + expectedStateSize + expectedKeySize + expectedValSize

	// Encode & Decode with AsciiEncoder
	e1 := NewBasicEncoder(bytesEncoderEuid, 0, expectedStateSize, expectedKeySize, expectedValSize)
	assert.NotNil(t, e1)
	assert.Equal(t, expectedWordSize, e1.WordSize())
	header := e1.Header()
	assert.NotNil(t, header)
	assert.Len(t, header, expectedWordSize)
}

func TestBasicEncoder_Match(t *testing.T) {
	expectedStateSize := 10
	expectedKeySize := 32
	expectedValSize := 100
	expectedWordSize := 12 + expectedStateSize + expectedKeySize + expectedValSize
	e1 := NewBasicEncoder(bytesEncoderEuid, 0, expectedStateSize, expectedKeySize, expectedValSize)
	assert.NotNil(t, e1)
	assert.Equal(t, expectedWordSize, e1.WordSize())
	header := e1.Header()

	m := e1.Match([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.False(t, m)
	m = e1.Match(header)
	assert.True(t, m)
}

func TestBasicEncoder_EncodeDecode(t *testing.T) {
	expectedStateSize := 10
	expectedKeySize := 8
	expectedValSize := 100
	expectedWordSize := 12 + expectedStateSize + expectedKeySize + expectedValSize
	e1 := NewBasicEncoder(bytesEncoderEuid, 0, expectedStateSize, expectedKeySize, expectedValSize)
	assert.NotNil(t, e1)
	assert.Equal(t, expectedWordSize, e1.WordSize())
	header := e1.Header()

	expectedSeq := 3
	expectedState := BuildState(expectedStateSize, "abcdefg")
	expectedTime := time.Now()
	key := []byte("key")
	expectedKey := append(key, []byte{0, 0, 0, 0, 0}...)
	expectedVal := []byte("foobarbaz")

	buf, err := e1.Encode(expectedSeq, expectedTime, expectedState, key, expectedVal)
	assert.NoError(t, err)
	require.NotNil(t, buf)
	assert.Len(t, buf, expectedWordSize)

	_, _, _, _, _, err = e1.Decode([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.Error(t, err)

	seq, tim, s, key, val, err := e1.Decode(buf)
	assert.NoError(t, err)
	assert.Equal(t, expectedSeq, seq)
	require.NotNil(t, s)
	assert.Equal(t, expectedState, s)
	require.NotNil(t, tim)
	assert.Equal(t, expectedTime.Unix()/86400, tim.Unix()/86400)
	require.NotNil(t, key)
	assert.Equal(t, expectedKey, key)
	require.NotNil(t, val)
	assert.Equal(t, expectedVal, val)

	// Setup & Decode with a new AsciiEncoder
	e2 := NewBasicEncoder(bytesEncoderEuid, 0, 0, 0, 0)
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
	seq, tim, s, key, val, err = e2.Decode(buf)
	assert.NoError(t, err)
	assert.Equal(t, expectedSeq, seq)
	require.NotNil(t, s)
	assert.Equal(t, expectedState, s)
	require.NotNil(t, tim)
	assert.Equal(t, expectedTime.Unix()/86400, tim.Unix()/86400)
	require.NotNil(t, key)
	assert.Equal(t, expectedKey, key)
	require.NotNil(t, val)
	assert.Equal(t, expectedVal, val)
}

func TestBasicEncoder_DecodeAll(t *testing.T) {
	expectedStateSize := 10
	expectedKeySize := 4
	expectedValSize := 100

	// Iterate with AsciiEncoder
	expectedState1 := BuildState(expectedStateSize, "pif")
	expectedTime1 := time.Now()
	expectedState2 := BuildState(expectedStateSize, "paf")
	expectedTime2 := time.Now()
	expectedState3 := BuildState(expectedStateSize, "pof")
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
	e3 := NewBasicEncoder(0, 0, expectedStateSize, expectedKeySize, expectedValSize)
	buf, err := e3.Encode(0, expectedTime1, expectedState1, key1, expectedText1)
	assert.NoError(t, err)
	bufs = append(bufs, buf...)
	buf, err = e3.Encode(1, expectedTime2, expectedState2, key2, expectedText2)
	assert.NoError(t, err)
	bufs = append(bufs, buf...)
	buf, err = e3.Encode(2, expectedTime3, expectedState3, key3, expectedText3)
	assert.NoError(t, err)
	bufs = append(bufs, buf...)

	k := 0
	e3.DecodeAll(TopToBottom, bufs, func(seq int, tim time.Time, s State, key []byte, text []byte, err error) bool {
		assert.Equal(t, k, seq)
		assert.NoError(t, err)
		switch k {
		case 0:
			assert.Equal(t, expectedState1, s)
			assert.Equal(t, expectedTime1.Unix()/86400, tim.Unix()/86400)
			assert.Equal(t, expectedKey1, key)
			assert.Equal(t, expectedText1, text)
		case 1:
			assert.Equal(t, expectedState2, s)
			assert.Equal(t, expectedTime2.Unix()/86400, tim.Unix()/86400)
			assert.Equal(t, expectedKey2, key)
			assert.Equal(t, expectedText2, text)
		case 2:
			assert.Equal(t, expectedState3, s)
			assert.Equal(t, expectedTime3.Unix()/86400, tim.Unix()/86400)
			assert.Equal(t, expectedKey3, key)
			assert.Equal(t, expectedText3, text)

		}
		k++
		return true
	})

	k = 0
	e3.DecodeAll(BottomToTop, bufs, func(seq int, tim time.Time, s State, key []byte, text []byte, err error) bool {
		assert.Equal(t, 2-k, seq)
		assert.NoError(t, err)
		switch k {
		case 0:
			assert.Equal(t, expectedState3, s)
			assert.Equal(t, expectedTime3.Unix()/86400, tim.Unix()/86400)
			assert.Equal(t, expectedKey3, key)
			assert.Equal(t, expectedText3, text)
		case 1:
			assert.Equal(t, expectedState2, s)
			assert.Equal(t, expectedTime2.Unix()/86400, tim.Unix()/86400)
			assert.Equal(t, expectedKey2, key)
			assert.Equal(t, expectedText2, text)
		case 2:
			assert.Equal(t, expectedState1, s)
			assert.Equal(t, expectedTime1.Unix()/86400, tim.Unix()/86400)
			assert.Equal(t, expectedKey1, key)
			assert.Equal(t, expectedText1, text)

		}
		k++
		return true
	})
}

func TestBasicEncoder_DecodeLastWord(t *testing.T) {
	expectedStateSize := 10
	expectedKeySize := 4
	expectedValSize := 100

	// Iterate with AsciiEncoder
	expectedState1 := BuildState(expectedStateSize, "pif")
	expectedTime1 := time.Now()
	expectedState2 := BuildState(expectedStateSize, "paf")
	expectedTime2 := time.Now()
	expectedState3 := BuildState(expectedStateSize, "pof")
	expectedTime3 := time.Now()
	key1 := []byte("k1")
	key2 := []byte("k2")
	key3 := []byte("k3")
	expectedKey3 := append(key3, 0, 0)
	expectedText1 := []byte("foo")
	expectedText2 := []byte("bar")
	expectedText3 := []byte("baz")

	var bufs []byte
	e3 := NewBasicEncoder(0, 0, expectedStateSize, expectedKeySize, expectedValSize)
	buf, err := e3.Encode(0, expectedTime1, expectedState1, key1, expectedText1)
	assert.NoError(t, err)
	bufs = append(bufs, buf...)
	buf, err = e3.Encode(1, expectedTime2, expectedState2, key2, expectedText2)
	assert.NoError(t, err)
	bufs = append(bufs, buf...)
	buf, err = e3.Encode(2, expectedTime3, expectedState3, key3, expectedText3)
	assert.NoError(t, err)
	bufs = append(bufs, buf...)

	// Decode last Word
	seq, tim, s, key, text, err := e3.DecodeLastWord(bufs)
	assert.NoError(t, err)
	assert.Equal(t, 2, seq)
	assert.Equal(t, expectedState3, s)
	assert.Equal(t, expectedTime3.Unix()/86400, tim.Unix()/86400)
	assert.Equal(t, expectedKey3, key)
	assert.Equal(t, expectedText3, text)
}
