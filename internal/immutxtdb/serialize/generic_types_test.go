package serialize

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func encodeString(i string, l int) ([]byte, error) {
	s := len(i)
	if s > l {
		return nil, fmt.Errorf("overflow")
	}
	buf := make([]byte, l)
	conv := []byte(i)
	copy(buf[l-len(conv):], conv)
	return buf, nil
}

func encodeFixedSized(fixedSized any, l int) ([]byte, error) {
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

type ComparableSerializer[K comparable] struct {
	length int
}

func (e ComparableSerializer[K]) Serialize(input K) ([]byte, error) {
	var fixedSized any
	switch i := any(input).(type) {
	case string:
		// string is not fixed-sized
		return encodeString(i, e.length)
	case int:
		// int is not fixed-sized
		fixedSized = int64(i)
	default:
		fixedSized = input
	}

	return encodeFixedSized(fixedSized, e.length)
}

type AnyEncoder[K any] struct {
	length int
}

func (e AnyEncoder[K]) Serialize(input K) ([]byte, error) {
	var fixedSized any
	switch i := any(input).(type) {
	case string:
		// string is not fixed-sized
		return encodeString(i, e.length)
	case int:
		// int is not fixed-sized
		fixedSized = int64(i)
	default:
		fixedSized = input
	}

	return encodeFixedSized(fixedSized, e.length)
}

func TestComparableEncoder(t *testing.T) {
	var res []byte
	var err error

	a100 := ComparableSerializer[int8]{length: 12}
	res, err = a100.Serialize(100)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100}, res)

	a101 := ComparableSerializer[int16]{length: 12}
	res, err = a101.Serialize(101)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 101}, res)

	a102 := ComparableSerializer[byte]{length: 12}
	res, err = a102.Serialize(102)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 102}, res)

	a103 := ComparableSerializer[[2]byte]{length: 12}
	res, err = a103.Serialize([2]byte{103, 103})
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 103, 103}, res)

	a105 := ComparableSerializer[int]{length: 12}
	res, err = a105.Serialize(105)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 105}, res)

	a200 := ComparableSerializer[string]{length: 12}
	res, err = a200.Serialize("fee")
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0x66, 0x65, 0x65}, res)

	a201 := ComparableSerializer[string]{length: 12}
	res, err = a201.Serialize("fée") // & => 0xC3 0xA9
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0x66, 0xC3, 0xA9, 0x65}, res)

	a202 := ComparableSerializer[string]{length: 12}
	res, err = a202.Serialize("😜") // 😜 => 0x01 0xF6 0x1C
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0xF0, 0x9F, 0x98, 0x9C}, res)

	a210 := ComparableSerializer[string]{length: 2}
	res, err = a210.Serialize("fee")
	assert.Error(t, err)
	assert.Nil(t, res)

	a211 := ComparableSerializer[string]{length: 2}
	res, err = a211.Serialize("😜")
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestAnyEncoder(t *testing.T) {
	var res []byte
	var err error

	a100 := AnyEncoder[int8]{length: 12}
	res, err = a100.Serialize(100)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100}, res)

	a101 := AnyEncoder[int16]{length: 12}
	res, err = a101.Serialize(101)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 101}, res)

	a102 := AnyEncoder[byte]{length: 12}
	res, err = a102.Serialize(102)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 102}, res)

	a103 := AnyEncoder[[2]byte]{length: 12}
	res, err = a103.Serialize([2]byte{103, 103})
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 103, 103}, res)

	a105 := AnyEncoder[int]{length: 12}
	res, err = a105.Serialize(105)
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 105}, res)

	a106 := AnyEncoder[[]byte]{length: 12}
	res, err = a106.Serialize([]byte{106})
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 106}, res)

	a107 := AnyEncoder[[]byte]{length: 12}
	res, err = a107.Serialize([]byte{107, 107, 107})
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 107, 107, 107}, res)

	a200 := AnyEncoder[string]{length: 12}
	res, err = a200.Serialize("fee")
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0x66, 0x65, 0x65}, res)

	a201 := AnyEncoder[string]{length: 12}
	res, err = a201.Serialize("fée") // é => 0xC3 0xA9
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0x66, 0xC3, 0xA9, 0x65}, res)

	a202 := AnyEncoder[string]{length: 12}
	res, err = a202.Serialize("😜") // 😜 => 0x01 0xF6 0x1C
	assert.NoError(t, err)
	assert.Len(t, res, 12)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0xF0, 0x9F, 0x98, 0x9C}, res)

	a210 := AnyEncoder[string]{length: 2}
	res, err = a210.Serialize("fee")
	assert.Error(t, err)
	assert.Nil(t, res)

	a211 := AnyEncoder[string]{length: 2}
	res, err = a211.Serialize("😜")
	assert.Error(t, err)
	assert.Nil(t, res)

	a220 := AnyEncoder[[]byte]{length: 2}
	res, err = a220.Serialize([]byte{107, 107, 107})
	assert.Error(t, err)
	assert.Nil(t, res)

}

func writeByteToSlice(bs []byte, b byte) {
	bs[0] = b
}

func TestWriteByteToSlice(t *testing.T) {
	b := make([]byte, 1)

	writeByteToSlice(b, 'a')
	assert.Len(t, b, 1)
	assert.Equal(t, []byte{'a'}, b)
}

func writeByteToSlicePtr(bs *[]byte, b byte) {
	(*bs)[0] = b
}

func TestWriteByteToSlicePtr(t *testing.T) {
	b := make([]byte, 1)

	writeByteToSlicePtr(&b, 'a')
	assert.Len(t, b, 1)
	assert.Equal(t, []byte{'a'}, b)
}
