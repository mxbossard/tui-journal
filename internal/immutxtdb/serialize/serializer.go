package serialize

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"
	"unicode"
)

var NullChar = []byte{0}

var NotMatchingEncoder = errors.New("encoder dos not match")
var NotAsciiText = errors.New("supplied text is out of ASCII table")

type Serializer[T any] interface {
	Serialize(T, []byte) (int, error)
	Deserialize([]byte) (T, error)
}

type ByteSliceSerializer struct {
	Serializer[[]byte]
}

func (s ByteSliceSerializer) Serialize(i, o []byte) (int, error) {
	copy(o, i)
	return len(i), nil
}

func (s ByteSliceSerializer) Deserialize(i []byte) ([]byte, error) {
	return i, nil
}

type ByteArray128Serializer struct {
	Serializer[*[128]byte]
}

func (s ByteArray128Serializer) Serialize(i *[128]byte, o []byte) (int, error) {
	for k := range 128 {
		o[k] = (*i)[k]
	}
	return len(i), nil
}

func (s ByteArray128Serializer) Deserialize(i []byte) (*[128]byte, error) {
	var o [128]byte
	// copy(o[:], i)
	for k := 0; k < 128 && k < len(i); k++ {
		o[k] = i[k]
	}
	return &o, nil
}

type PtrSerializer struct {
	Serializer[[]byte]
}

func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

type AsciiSerializer struct {
	Serializer[string]
}

func (s AsciiSerializer) Serialize(i string, o []byte) (int, error) {
	if !IsASCII(i) {
		return 0, NotAsciiText
	}
	n := copy(o, i)
	o[n] = NullChar[0]
	return n + 1, nil
}

func (s AsciiSerializer) Deserialize(b []byte) (string, error) {
	o, _, _ := bytes.Cut(b, NullChar)
	return string(o), nil
}

type TimeSerializer struct {
	Serializer[time.Time]
}

func (s TimeSerializer) Serialize(i time.Time, o []byte) (int, error) {
	binary.AppendVarint(o, i.Unix())
	return 8, nil
}

func (s TimeSerializer) Deserialize(i []byte) (time.Time, error) {
	t, _ := binary.Varint(i[0:8])
	return time.Unix(t, 0), nil
}

type StructSerializer[T any] struct {
	Serializer[T]
}

func (s StructSerializer[T]) Serialize(i *T, o []byte) (int, error) {
	n, err := binary.Encode(o, binary.BigEndian, &i)
	return n, err
}

func (s StructSerializer[T]) Deserialize(i []byte) (o *T, err error) {
	_, err = binary.Decode(i, binary.BigEndian, o)
	return o, err
}
