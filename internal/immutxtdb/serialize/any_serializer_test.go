package serialize

import (
	"testing"

	"github.com/mxbossard/utilz/serializ"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fixedSizedStruct struct {
	A int16
}

type variableSizedStruct struct {
	A string
	B int
	C []byte
	D int32
}

func (s variableSizedStruct) Serialize() ([]byte, error) {
	return serializ.ToBytes(s.A, 16, s.B, 4, s.C, 5, s.D, 4)
	// buf := new(bytes.Buffer)
	// var err error
	// // Limit Filepath to 128 bytes ?
	// err = binary.Write(buf, binary.BigEndian, []byte(s.A))
	// if err != nil {
	// 	return nil, err
	// }
	// return buf.Bytes(), nil
}

func (s *variableSizedStruct) Deserialize(bs []byte) error {
	return serializ.FromBytes(bs, &s.A, 16, &s.B, 4, &s.C, 5, &s.D, 4)
	// s.A = string(bs)
	// return nil
}

func TestAnySerializer_Serialize(t *testing.T) {
	var bs []byte
	var n int
	var err error

	bs = make([]byte, 4)
	as100 := AnySerializer[int8]{Length: 3}
	n, err = as100.Serialize(100, &bs)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{100, 0, 0, 0}, bs)

	bs = make([]byte, 4)
	as101 := AnySerializer[int16]{Length: 3}
	n, err = as101.Serialize(101, &bs)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{0, 101, 0, 0}, bs)

	bs = make([]byte, 4)
	as102 := AnySerializer[byte]{Length: 3}
	n, err = as102.Serialize(102, &bs)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{102, 0, 0, 0}, bs)

	bs = make([]byte, 4)
	as103 := AnySerializer[[2]byte]{Length: 3}
	n, err = as103.Serialize([2]byte{103, 103}, &bs)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{103, 103, 0, 0}, bs)

	bs = make([]byte, 4)
	as105 := AnySerializer[int]{Length: 3}
	n, err = as105.Serialize(105, &bs)
	assert.ErrorContains(t, err, "buffer too small")
	assert.Equal(t, 0, n)
	bs = make([]byte, 11)
	as105 = AnySerializer[int]{Length: 10}
	n, err = as105.Serialize(105, &bs)
	assert.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, []byte{0, 0, 0, 0, 0, 0, 0, 105, 0, 0, 0}, bs)

	bs = make([]byte, 4)
	as106 := AnySerializer[[]byte]{Length: 3}
	n, err = as106.Serialize([]byte{106}, &bs)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{106, 0, 0, 0}, bs)

	bs = make([]byte, 4)
	as107 := AnySerializer[[]byte]{Length: 3}
	n, err = as107.Serialize([]byte{107, 107, 107}, &bs)
	assert.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{107, 107, 107, 0}, bs)

	bs = make([]byte, 7)
	as200 := AnySerializer[string]{Length: 5}
	n, err = as200.Serialize("fee", &bs)
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte{0x66, 0x65, 0x65, 0, 0, 0, 0}, bs)

	bs = make([]byte, 7)
	as201 := AnySerializer[string]{Length: 5}
	n, err = as201.Serialize("fée", &bs) // é => 0xC3 0xA9
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte{0x66, 0xC3, 0xA9, 0x65, 0, 0, 0}, bs)

	bs = make([]byte, 7)
	as202 := AnySerializer[string]{Length: 5}
	n, err = as202.Serialize("😜", &bs) // 😜 => 0x01 0xF6 0x1C
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, []byte{0xF0, 0x9F, 0x98, 0x9C, 0, 0, 0}, bs)

	bs = make([]byte, 10)
	as300 := AnySerializer[fixedSizedStruct]{Length: 8}
	n, err = as300.Serialize(fixedSizedStruct{300}, &bs)
	assert.NoError(t, err)
	assert.Equal(t, 8, n)
	assert.Equal(t, []byte{0x01, 0x2c, 0, 0, 0, 0, 0, 0, 0, 0}, bs)

	bs = make([]byte, 64)
	as400 := AnySerializer[variableSizedStruct]{Length: 8}
	n, err = as400.Serialize(variableSizedStruct{A: "foo", B: 42, C: []byte{1, 2, 3}, D: int32(32)}, &bs)
	assert.NoError(t, err)
}

func TestAnySerializer_Deserialize(t *testing.T) {
	var bs []byte
	var n int
	var err error

	bs = make([]byte, 4)
	as100 := AnySerializer[int8]{Length: 3}
	n, err = as100.Serialize(100, &bs)
	require.NoError(t, err)
	assert.Equal(t, 3, n)
	assert.Equal(t, []byte{100, 0, 0, 0}, bs)
	o100, err := as100.Deserialize(bs)
	assert.NoError(t, err)
	assert.Equal(t, int8(100), o100)

	bs = make([]byte, 4)
	as101 := AnySerializer[int16]{Length: 3}
	n, err = as101.Serialize(101, &bs)
	require.NoError(t, err)
	o101, err := as101.Deserialize(bs)
	assert.NoError(t, err)
	assert.Equal(t, int16(101), o101)

	bs = make([]byte, 4)
	as102 := AnySerializer[byte]{Length: 3}
	n, err = as102.Serialize(102, &bs)
	require.NoError(t, err)
	o102, err := as102.Deserialize(bs)
	assert.NoError(t, err)
	assert.Equal(t, byte(102), o102)

	bs = make([]byte, 4)
	as103 := AnySerializer[[2]byte]{Length: 3}
	n, err = as103.Serialize([2]byte{103, 103}, &bs)
	require.NoError(t, err)
	o103, err := as103.Deserialize(bs)
	assert.NoError(t, err)
	assert.Equal(t, [2]byte{103, 103}, o103)

	bs = make([]byte, 11)
	as105 := AnySerializer[int]{Length: 10}
	n, err = as105.Serialize(105, &bs)
	require.NoError(t, err)
	o105, err := as105.Deserialize(bs)
	assert.NoError(t, err)
	assert.Equal(t, 105, o105)

	bs = make([]byte, 4)
	as106 := AnySerializer[[]byte]{Length: 3}
	n, err = as106.Serialize([]byte{106}, &bs)
	require.NoError(t, err)
	o106, err := as106.Deserialize(bs)
	assert.NoError(t, err)
	assert.Equal(t, []byte{106, 0, 0}, o106)

	bs = make([]byte, 4)
	as107 := AnySerializer[[]byte]{Length: 3}
	n, err = as107.Serialize([]byte{107, 107, 107}, &bs)
	require.NoError(t, err)
	o107, err := as107.Deserialize(bs)
	assert.NoError(t, err)
	assert.Equal(t, []byte{107, 107, 107}, o107)

	bs = make([]byte, 7)
	as200 := AnySerializer[string]{Length: 5}
	n, err = as200.Serialize("fee", &bs)
	require.NoError(t, err)
	o200, err := as200.Deserialize(bs)
	assert.NoError(t, err)
	assert.Equal(t, "fee", o200)

	bs = make([]byte, 7)
	as201 := AnySerializer[string]{Length: 5}
	n, err = as201.Serialize("fée", &bs) // é => 0xC3 0xA9
	require.NoError(t, err)
	o201, err := as201.Deserialize(bs)
	assert.NoError(t, err)
	assert.Equal(t, "fée", o201)

	bs = make([]byte, 7)
	as202 := AnySerializer[string]{Length: 5}
	n, err = as202.Serialize("😜", &bs) // 😜 => 0x01 0xF6 0x1C
	require.NoError(t, err)
	o202, err := as202.Deserialize(bs)
	assert.NoError(t, err)
	assert.Equal(t, "😜", o202)

	bs = make([]byte, 10)
	as300 := AnySerializer[fixedSizedStruct]{Length: 8}
	n, err = as300.Serialize(fixedSizedStruct{300}, &bs)
	require.NoError(t, err)
	o300, err := as300.Deserialize(bs)
	assert.NoError(t, err)
	assert.Equal(t, fixedSizedStruct{300}, o300)

	bs = make([]byte, 64)
	as400 := AnySerializer[variableSizedStruct]{Length: 8}
	n, err = as400.Serialize(variableSizedStruct{A: "foo", B: 42, C: []byte{1, 2, 3}, D: int32(32)}, &bs)
	require.NoError(t, err)
	o400, err := as400.Deserialize(bs)
	assert.NoError(t, err)
	assert.Equal(t, variableSizedStruct{A: "foo", B: 42, C: []byte{1, 2, 3, 0, 0}, D: int32(32)}, o400)
}
