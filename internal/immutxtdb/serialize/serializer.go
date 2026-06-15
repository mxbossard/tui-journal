package serialize

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"time"
	"unicode"

	"github.com/mxbossard/utilz/inoutz"
)

var NullChar = []byte{0}

var NotMatchingEncoder = errors.New("encoder dos not match")
var NotAsciiText = errors.New("supplied text is out of ASCII table")

type Serializer[T any] interface {
	Serialize(T, []byte) (int, error)
	Deserialize([]byte) (T, error)
}

type Serializer2[T any] interface {
	Serialize(T, *[]byte) (int, error)
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

type ByteArray16Serializer struct {
	Serializer[*[16]byte]
}

func (s ByteArray16Serializer) Serialize(i *[16]byte, o []byte) (int, error) {
	for k := range 16 {
		o[k] = (*i)[k]
	}
	return len(i), nil
}

func (s ByteArray16Serializer) Deserialize(i []byte) (*[16]byte, error) {
	var o [16]byte
	for k := 0; k < 16 && k < len(i); k++ {
		o[k] = i[k]
	}
	return &o, nil
}

type ByteArray32Serializer struct {
	Serializer[*[32]byte]
}

func (s ByteArray32Serializer) Serialize(i *[32]byte, o []byte) (int, error) {
	for k := range 32 {
		o[k] = (*i)[k]
	}
	return len(i), nil
}

func (s ByteArray32Serializer) Deserialize(i []byte) (*[32]byte, error) {
	var o [32]byte
	for k := 0; k < 32 && k < len(i); k++ {
		o[k] = i[k]
	}
	return &o, nil
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
	bw := inoutz.NewByteSliceWriter(o)
	var err error
	if i != nil {
		enc := gob.NewEncoder(bw)
		err = enc.Encode(*i)
	}
	return bw.Len(), err
}

func (s StructSerializer[T]) Deserialize(i []byte) (o *T, err error) {
	buf := bytes.NewBuffer(i)
	dec := gob.NewDecoder(buf)
	var l T
	err = dec.Decode(&l)
	return &l, err
}

type BinarySerializer struct {
	Serializer2[any]
}

func (s BinarySerializer) Serialize(i any, o *[]byte) (int, error) {
	buf := make([]byte, 100)
	n, err := binary.Encode(buf, binary.BigEndian, i)
	if err != nil {
		return -1, err
	}
	if len(*o) < n {
		err = fmt.Errorf("supplied byte slice is too small, need %d bytes", n)
		return -1, err
	}
	p := copy(*o, buf[0:n])
	if p != n {
		err = fmt.Errorf("bad byte count written %d/%d", p, n)
		return p, err
	}
	return n, err
}

func (s BinarySerializer) Deserialize(i []byte) (o any, err error) {
	_, err = binary.Decode(i, binary.BigEndian, o)
	return
}
