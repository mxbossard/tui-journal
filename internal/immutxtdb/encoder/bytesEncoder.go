package encoder

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
)

type bytesEncoder struct {
	*basicEncoder[[]byte]
}

func NewBytesEncoder(version int32, stateSize, dataSize int) *bytesEncoder {
	e := &basicEncoder[[]byte]{
		uid:       bytesEncoderEuid,
		version:   version,
		stateSize: stateSize,
		dataSize:  dataSize,
	}
	return &bytesEncoder{
		basicEncoder: e,
	}
}

func (e bytesEncoder) Encode(seq int, s model.State, data []byte) ([]byte, error) {
	if len(data) > e.dataSize {
		return nil, errors.New("data is longer than configured dataSize")
	}

	buf := make([]byte, e.wordSize())
	k := 0
	n, err := binary.Encode(buf[k:], binary.BigEndian, int32(seq))
	if err != nil {
		return nil, fmt.Errorf("encoding seq: %w", err)
	}
	k += n

	n, err = binary.Encode(buf[k:], binary.BigEndian, []byte(s))
	if err != nil {
		return nil, fmt.Errorf("encoding delimiterChar: %w", err)
	}
	k += e.stateSize

	dataLen := len(data)
	n, err = binary.Encode(buf[k:], binary.BigEndian, int32(dataLen))
	if err != nil {
		return nil, fmt.Errorf("encoding data length: %w", err)
	}
	k += n

	n, err = binary.Encode(buf[k:], binary.BigEndian, data)
	if err != nil {
		return nil, fmt.Errorf("encoding data: %w", err)
	}
	k += e.dataSize

	return buf, nil
}

func (e *bytesEncoder) Decode(buf []byte) (int, model.State, []byte, error) {
	var seq, dataLen int32
	var s model.State
	var data []byte
	// fmt.Printf("decoding config: statSize: %d ; dataSize: %d\n", e.stateSize, e.dataSize)

	if len(buf) < e.wordSize() {
		return int(seq), s, data, fmt.Errorf("cannot decode data of length: %d < wordSize: %d", len(buf), e.wordSize())
	}

	k := 0
	n, err := binary.Decode(buf[k:k+4], binary.BigEndian, &seq)
	if err != nil {
		return int(seq), s, data, fmt.Errorf("decoding seq: %w", err)
	}
	k += n

	var stateData = make([]byte, e.stateSize)
	n, err = binary.Decode(buf[k:k+e.stateSize], binary.BigEndian, &stateData)
	if err != nil {
		return int(seq), s, data, fmt.Errorf("decoding state: %w", err)
	}
	k += e.stateSize
	s = model.State(stateData)

	n, err = binary.Decode(buf[k:k+4], binary.BigEndian, &dataLen)
	if err != nil {
		return int(seq), s, data, fmt.Errorf("decoding data length: %w", err)
	}
	k += n

	if dataLen > int32(e.dataSize) {
		return int(seq), s, data, fmt.Errorf("bad encoded data size: %d", dataLen)
	}

	data = make([]byte, dataLen)
	n, err = binary.Decode(buf[k:k+int(dataLen)], binary.BigEndian, &data)
	if err != nil {
		return int(seq), s, data, fmt.Errorf("decoding text: %w", err)
	}
	k += e.dataSize

	return int(seq), s, data, nil

}

func (e *bytesEncoder) DecodeLastWord(buf []byte) (int, model.State, []byte, error) {
	lastWordStart := (len(buf)/e.wordSize() - 1) * e.wordSize()
	return e.Decode(buf[lastWordStart:])
}

func (e bytesEncoder) DecodeAll(order model.Order, buf []byte, push func(int, model.State, []byte, error)) {
	wordSize := e.wordSize()
	if order == model.TopToBottom {
		for k := 0; k < len(buf); k += wordSize {
			seq, state, text, err := e.Decode(buf[k:])
			push(seq, state, text, err)
		}
	} else {
		wordCount := len(buf) / wordSize
		for k := (wordCount - 1) * wordSize; k >= 0; k -= wordSize {
			seq, state, text, err := e.Decode(buf[k : k+wordSize])
			push(seq, state, text, err)
		}

	}
}
