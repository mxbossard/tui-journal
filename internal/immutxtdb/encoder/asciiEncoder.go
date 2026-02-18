package encoder

import (
	"errors"
	"fmt"
	"unicode"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
)

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

type asciiEncoder struct {
	*basicEncoder[string]
	bytesEncoder *bytesEncoder
}

func NewAsciiEncoder(version int32, stateSize, dataSize int) *asciiEncoder {
	e := &basicEncoder[string]{
		uid:       asciiEncoderEuid,
		version:   version,
		stateSize: stateSize,
		dataSize:  dataSize,
	}
	bytesEncoder := NewBytesEncoder(0, stateSize, dataSize)
	bytesEncoder.uid = asciiEncoderEuid
	return &asciiEncoder{
		basicEncoder: e,
		bytesEncoder: bytesEncoder,
	}
}
func (e *asciiEncoder) Setup(header []byte) error {
	err := e.bytesEncoder.Setup(header)
	if err != nil {
		return err
	}
	return e.basicEncoder.Setup(header)
}

func (e asciiEncoder) Encode(seq int, s model.State, text string) ([]byte, error) {
	if !isASCII(text) {
		return nil, NotAsciiText
	}
	data := []byte(text)
	if len(data) != len(text) {
		return nil, errors.New("encoded text is longer than text")
	}
	return e.bytesEncoder.Encode(seq, s, data)
}

func (e asciiEncoder) Decode(buf []byte) (int, model.State, string, error) {
	if len(buf) < e.wordSize() {
		return 0, nil, "", fmt.Errorf("cannot decode data of length: %d < wordSize: %d", len(buf), e.wordSize())
	}

	seq, s, data, err := e.bytesEncoder.Decode(buf)
	if err != nil {
		return 0, nil, "", err
	}

	text := string(data)
	return int(seq), s, text, nil
}

func (e asciiEncoder) DecodeLastWord(buf []byte) (int, model.State, string, error) {
	lastWordStart := (len(buf)/e.wordSize() - 1) * e.wordSize()
	return e.Decode(buf[lastWordStart:])
}

func (e asciiEncoder) DecodeAll(order model.Order, buf []byte, push func(int, model.State, string, error)) {
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
