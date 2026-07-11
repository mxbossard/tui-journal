package idx

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

type A[K comparable] struct{}

func (a A[K]) Convert(in K, l int) ([]byte, error) {
	buf := make([]byte, l)
	_, err := binary.Encode(buf, binary.BigEndian, in)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func TestConvert(t *testing.T) {
	var res []byte
	var err error

	a0 := A[int8]{}
	res, err = a0.Convert(100, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, res)

	a1 := A[int16]{}
	res, err = a1.Convert(101, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 101, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, res)

	a2 := A[byte]{}
	res, err = a2.Convert(102, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{102, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, res)

	a3 := A[[2]byte]{}
	res, err = a3.Convert([2]byte{103, 103}, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{103, 103, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, res)

	a4 := A[string]{}
	res, err = a4.Convert("foo", 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0x66, 0, 0, 0x6f, 0, 0, 0x6f, 0, 0, 0}, res)
}
