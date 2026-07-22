package serialize

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"

	"github.com/mxbossard/utilz/errorz"
	"github.com/mxbossard/utilz/inoutz"
)

type AnySerializer[T any] struct {
	Serializer2[T]

	Length int
}

func serializeString(i string, length int, bs *[]byte) (int, error) {
	conv := []byte(i)
	s := len(conv)
	if len(conv) > length {
		return 0, fmt.Errorf("overflow")
	}
	// Copy the string at the end of the byte slice
	n := copy((*bs)[:s], conv)
	return n, nil
}

func deserializeString(bs []byte) (string, error) {
	return string(bytes.Trim(bs, "\x00")), nil
}

func gobSerialize(i any, length int, bs *[]byte) (int, error) {
	bw := inoutz.NewByteSliceWriter(*bs)
	enc := gob.NewEncoder(bw)
	err := enc.Encode(i)
	n := bw.Len()
	return n, err
}

func gobDeserialize[T any](bs []byte) (d T, err error) {
	buf := bytes.NewBuffer(bs)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&d)
	return
}

func serializeStringRightShift(i string, length int, bs *[]byte) (int, error) {
	conv := []byte(i)
	s := len(conv)
	if len(conv) > length {
		return 0, fmt.Errorf("overflow")
	}
	// Copy the string at the end of the byte slice
	n := copy((*bs)[length-s:], conv)
	return n + length - s, nil
}

func serializeFixedSized(fixedSized any, l int, bs *[]byte) (int, error) {
	buf := make([]byte, l)
	n, err := binary.Encode(buf, binary.BigEndian, fixedSized)
	if err != nil {
		return 0, err
	}
	if n > l {
		return 0, fmt.Errorf("overflow")
	}

	p := copy((*bs)[:n], buf[:n])
	return p, nil
}

func deserializeFixedSized[T any](bs []byte) (d T, err error) {
	n, err := binary.Decode(bs, binary.BigEndian, &d)
	if err != nil {
		return
	}
	_ = n
	return
}

func serializeFixedSizedRightShift(fixedSized any, l int, bs *[]byte) (int, error) {
	buf := make([]byte, l)
	n, err := binary.Encode(buf, binary.BigEndian, fixedSized)
	if err != nil {
		return 0, err
	}
	if n > l {
		return 0, fmt.Errorf("overflow")
	}

	// buf2 := make([]byte, l)
	p := copy((*bs)[l-n:], buf[:n])
	return p + l - n, nil
}

func (s AnySerializer[T]) Serialize(input T, o *[]byte) (n int, err error) {
	var anyInput any
	if i, ok := any(&input).(Serializable); ok {
		bs, err := i.Serialize()
		if err != nil {
			return 0, err
		}
		n = copy(*o, bs)
	} else {
		switch i := any(input).(type) {
		case string:
			// string is not fixed-sized
			n, err = serializeString(i, s.Length, o)
		case int:
			// int is not fixed-sized
			anyInput = int64(i)
			n, err = serializeFixedSized(anyInput, s.Length, o)
		default:
			anyInput = input
			n, err = serializeFixedSized(anyInput, s.Length, o)
		}
	}

	if errorz.Contains(err, "not fixed-sized") {
		// Attempt gob serializer
		err = fmt.Errorf("type %T not fixed-sized and do not implement Serializable interface: %w", input, err)
	}

	if err != nil {
		return
	}

	if n < s.Length {
		for k := n; k < s.Length; k++ {
			(*o)[k] = 0
			n++
		}
	}
	return
}

func (s AnySerializer[T]) Deserialize(input []byte) (o T, err error) {
	var ok bool
	if i, ok := any(&o).(Serializable); ok {
		err = i.Deserialize(input)
		return
	}

	switch any(o).(type) {
	case string:
		// string is not fixed-sized
		var s string
		s, err = deserializeString(input)
		o, _ = any(s).(T)
	case int:
		// int is not fixed-sized
		fixedSized, err := deserializeFixedSized[int64](input)
		if err != nil {
			return o, err
		}
		d := int(fixedSized)
		o, _ = any(d).(T)
	case []byte:
		d := make([]byte, s.Length)
		n := copy(d, input[:s.Length])
		if n < s.Length {
			return o, fmt.Errorf("missing input bytes %d < %d length", n, s.Length)
		}
		o, _ = any(d).(T)
	default:
		d, err := deserializeFixedSized[T](input)
		if err != nil {
			return o, err
		}
		o, ok = any(d).(T)
		if !ok {
			return o, fmt.Errorf("unable to deserialize %T data", o)
		}
	}

	return
}
