package idx

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func convertString(i string, l int) ([]byte, error) {
	s := len(i)
	if s > l {
		return nil, fmt.Errorf("overflow")
	}
	buf := make([]byte, l)
	conv := []byte(i)
	copy(buf[l-len(conv):], conv)
	return buf, nil
}

func convertFixedSized(fixedSized any, l int) ([]byte, error) {
	buf := make([]byte, l)
	n, err := binary.Encode(buf, binary.BigEndian, fixedSized)
	if err != nil {
		return nil, err
	}
	if n > l {
		return nil, fmt.Errorf("overflow")
	}

	buf2 := make([]byte, l)
	copy(buf2[l-n:], buf[:n])
	return buf2, nil
}

type ComparableEncoder[K comparable] struct{}

func (e ComparableEncoder[K]) Convert(input K, l int) ([]byte, error) {
	var fixedSized any
	switch i := any(input).(type) {
	case string:
		// string is not fixed-sized
		return convertString(i, l)
	case int:
		// int is not fixed-sized
		fixedSized = int64(i)
	default:
		fixedSized = input
	}

	return convertFixedSized(fixedSized, l)
}

type AnyEncoder[K any] struct{}

func (e AnyEncoder[K]) Convert(input K, l int) ([]byte, error) {
	var fixedSized any
	switch i := any(input).(type) {
	case string:
		// string is not fixed-sized
		return convertString(i, l)
	case int:
		// int is not fixed-sized
		fixedSized = int64(i)
	default:
		fixedSized = input
	}

	return convertFixedSized(fixedSized, l)
}

func TestComparableEncoder(t *testing.T) {
	var res []byte
	var err error

	a100 := ComparableEncoder[int8]{}
	res, err = a100.Convert(100, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100}, res)

	a101 := ComparableEncoder[int16]{}
	res, err = a101.Convert(101, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 101}, res)

	a102 := ComparableEncoder[byte]{}
	res, err = a102.Convert(102, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 102}, res)

	a103 := ComparableEncoder[[2]byte]{}
	res, err = a103.Convert([2]byte{103, 103}, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 103, 103}, res)

	a105 := ComparableEncoder[int]{}
	res, err = a105.Convert(105, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 105}, res)

	a200 := ComparableEncoder[string]{}
	res, err = a200.Convert("fee", 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0x66, 0x65, 0x65}, res)

	a201 := ComparableEncoder[string]{}
	res, err = a201.Convert("fée", 12) // & => 0xC3 0xA9
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0x66, 0xC3, 0xA9, 0x65}, res)

	a202 := ComparableEncoder[string]{}
	res, err = a202.Convert("😜", 12) // 😜 => 0x01 0xF6 0x1C
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0xF0, 0x9F, 0x98, 0x9C}, res)

	a210 := ComparableEncoder[string]{}
	res, err = a210.Convert("fee", 2)
	assert.Error(t, err)
	assert.Nil(t, res)

	a211 := ComparableEncoder[string]{}
	res, err = a211.Convert("😜", 2)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestAnyEncoder(t *testing.T) {
	var res []byte
	var err error

	a100 := AnyEncoder[int8]{}
	res, err = a100.Convert(100, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100}, res)

	a101 := AnyEncoder[int16]{}
	res, err = a101.Convert(101, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 101}, res)

	a102 := AnyEncoder[byte]{}
	res, err = a102.Convert(102, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 102}, res)

	a103 := AnyEncoder[[2]byte]{}
	res, err = a103.Convert([2]byte{103, 103}, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 103, 103}, res)

	a105 := AnyEncoder[int]{}
	res, err = a105.Convert(105, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 105}, res)

	a106 := AnyEncoder[[]byte]{}
	res, err = a106.Convert([]byte{106}, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 106}, res)

	a107 := AnyEncoder[[]byte]{}
	res, err = a107.Convert([]byte{107, 107, 107}, 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 107, 107, 107}, res)

	a200 := AnyEncoder[string]{}
	res, err = a200.Convert("fee", 12)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0x66, 0x65, 0x65}, res)

	a201 := AnyEncoder[string]{}
	res, err = a201.Convert("fée", 12) // & => 0xC3 0xA9
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0x66, 0xC3, 0xA9, 0x65}, res)

	a202 := AnyEncoder[string]{}
	res, err = a202.Convert("😜", 12) // 😜 => 0x01 0xF6 0x1C
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0xF0, 0x9F, 0x98, 0x9C}, res)

	a210 := AnyEncoder[string]{}
	res, err = a210.Convert("fee", 2)
	assert.Error(t, err)
	assert.Nil(t, res)

	a211 := AnyEncoder[string]{}
	res, err = a211.Convert("😜", 2)
	assert.Error(t, err)
	assert.Nil(t, res)

	a220 := AnyEncoder[[]byte]{}
	res, err = a220.Convert([]byte{107, 107, 107}, 2)
	assert.Error(t, err)
	assert.Nil(t, res)

}
