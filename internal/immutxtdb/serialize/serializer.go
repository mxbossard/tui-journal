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
	Serialize(T, []byte) error
	Deserialize([]byte) (T, error)
}

type ByteSliceSerializer struct {
	Serializer[[]byte]
}

func (s ByteSliceSerializer) Serialize(i, o []byte) error {
	copy(o, i)
	return nil
}

func (s ByteSliceSerializer) Deserialize(i []byte) ([]byte, error) {
	return i, nil
}

// type ByteArraySerializer struct {
// 	Serializer[*[128]byte]
// }

// func (s ByteArraySerializer) Serialize(i *[128]byte, o []byte) error {
// 	o = i[:128]
// 	return nil
// }

// func (s ByteArraySerializer) Deserialize(i []byte) (*[128]byte, error) {
// 	return &i, nil
// }

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

func (s AsciiSerializer) Serialize(i string, o []byte) error {
	if !IsASCII(i) {
		return NotAsciiText
	}
	n := copy(o, i)
	o[n] = NullChar[0]
	return nil
}

func (s AsciiSerializer) Deserialize(b []byte) (string, error) {
	o, _, _ := bytes.Cut(b, NullChar)
	return string(o), nil
}

type TimeSerializer struct {
	Serializer[time.Time]
}

func (s TimeSerializer) Serialize(i time.Time, o []byte) error {
	binary.AppendVarint(o, i.Unix())
	return nil
}

func (s TimeSerializer) Deserialize(i []byte) (time.Time, error) {
	t, _ := binary.Varint(i[0:8])
	return time.Unix(t, 0), nil
}
