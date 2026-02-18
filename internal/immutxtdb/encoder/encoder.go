package encoder

import (
	"encoding/binary"
	"errors"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
)

type euid uint64

var (
	asciiEncoderEuid = euid(binary.BigEndian.Uint64([]byte("ascii000")))
	bytesEncoderEuid = euid(binary.BigEndian.Uint64([]byte("bytes000")))
)

var NotMatchingEncoder = errors.New("encoder dos not match")
var NotAsciiText = errors.New("supplied text is out of ASCII table")

type Encoder[T any] interface {
	wordSize() int
	Header() []byte
	Match(header []byte) bool
	Setup(header []byte) error
	Encode(seq int, s model.State, data T) ([]byte, error)
	// Decode first word in supplied byte slice.
	Decode([]byte) (int, model.State, T, error)
	// Decode last word in supplied byte slice.
	DecodeLastWord([]byte) (int, model.State, T, error)
	DecodeAll(model.Order, []byte, func(int, model.State, T, error))
}

type basicEncoder[T any] struct {
	Encoder[T]
	stateSize int
	dataSize  int
	uid       euid
	version   int32
}

func (e basicEncoder[T]) wordSize() int {
	return 8 + int(e.stateSize) + int(e.dataSize)
}

func (e basicEncoder[T]) Header() []byte {
	b := make([]byte, e.wordSize())
	k := 0
	n, err := binary.Encode(b[k:k+8], binary.BigEndian, e.uid)
	if err != nil {
		panic(err)
	}
	k += n
	n, err = binary.Encode(b[k:k+4], binary.BigEndian, e.version)
	if err != nil {
		panic(err)
	}
	k += n
	n, err = binary.Encode(b[k:k+4], binary.BigEndian, int32(e.stateSize))
	if err != nil {
		panic(err)
	}
	k += n
	n, err = binary.Encode(b[k:k+4], binary.BigEndian, int32(e.dataSize))
	if err != nil {
		panic(err)
	}
	k += n
	return b
}

func (e *basicEncoder[T]) Match(header []byte) bool {
	if len(header) < 12 {
		return false
	}
	k := 0
	var euid euid
	var version int32
	n, err := binary.Decode(header[k:k+8], binary.BigEndian, &euid)
	if err != nil {
		panic(err)
	}
	k += n
	n, err = binary.Decode(header[k:k+4], binary.BigEndian, &version)
	if err != nil {
		panic(err)
	}
	k += n
	return euid == e.uid && version == e.version
}

func (e *basicEncoder[T]) Setup(header []byte) error {
	if len(header) < 20 {
		return errors.New("need a header of 20 bytes minimum to setup")
	}

	if !e.Match(header) {
		return NotMatchingEncoder
	}
	k := 12
	var stateSize, dataSize int32
	n, err := binary.Decode(header[k:k+4], binary.BigEndian, &stateSize)
	if err != nil {
		return err
	}
	k += n
	n, err = binary.Decode(header[k:k+4], binary.BigEndian, &dataSize)
	if err != nil {
		return err
	}
	k += n

	// fmt.Printf("setup config: statSize: %d ; dataSize: %d\n", stateSize, dataSize)

	e.stateSize = int(stateSize)
	e.dataSize = int(dataSize)
	return nil
}
