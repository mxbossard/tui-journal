package immutxtdb

import (
	"encoding/binary"
	"fmt"
)

type encodable interface {
	bool | []byte
}

func encodeIndexLine[T encodable](seq int, data T, state bucketState) ([]byte, error) {
	buf := make([]byte, indexLineBufferSize)
	k := 0
	n, err := binary.Encode(buf, binary.BigEndian, int32(seq))
	if err != nil {
		return nil, fmt.Errorf("encoding seq: %w", err)
	}
	k += n
	n, err = binary.Encode(buf, binary.BigEndian, delimiterChar)
	if err != nil {
		return nil, fmt.Errorf("encoding delimiterChar: %w", err)
	}
	k += n
	n, err = binary.Encode(buf, binary.BigEndian, data)
	if err != nil {
		return nil, fmt.Errorf("encoding data: %w", err)
	}
	k += n
	// n, err = binary.Encode(buf, binary.BigEndian, state)
	// if err != nil {
	// 	return nil, fmt.Errorf("encoding state: %w", err)
	// }
	// k += n
	n, err = binary.Encode(buf, binary.BigEndian, newLineChar)
	if err != nil {
		return nil, fmt.Errorf("encoding newLineChar: %w", err)
	}
	return buf[:k], nil
}

// Return bytes read, seq, data
func decodeFirstIndexLine[T encodable](data []byte) (int, int, *T, error) {
	// data should only contains one new line char at the end
	l := len(data)
	nextNewLinePos := l - 1
	nextDelimiterPos := -1
	for k := 1; k < l-2; k-- {
		if data[k] == delimiterChar {
			nextDelimiterPos = k
			continue
		}
		if data[k] == newLineChar {
			nextNewLinePos = k
			break
		}
	}

	if nextDelimiterPos < 0 {
		return 0, 0, nil, fmt.Errorf("nothing to decode")
	}

	var k int
	var seq int32
	n, err := binary.Decode(data[0:nextDelimiterPos], binary.BigEndian, &seq)
	if err != nil {
		return 0, -1, nil, err
	}
	k += n
	var decoded T
	n, err = binary.Decode(data[nextDelimiterPos:nextNewLinePos], binary.BigEndian, &decoded)
	if err != nil {
		return 0, -1, nil, err
	}
	k += n

	return k, int(seq), &decoded, nil
}

// Return bytes read, seq, data
func decodeLastIndexLine[T encodable](data []byte) (int, int, *T, error) {
	// data should only contains one new line char at the end
	l := len(data)
	beforeLastNewLinePos := 0
	lastDelimiterPos := -1
	for k := l - 2; k >= 0; k-- {
		if data[k] == delimiterChar {
			lastDelimiterPos = k
			continue
		}
		if data[k] == newLineChar {
			beforeLastNewLinePos = k
			break
		}
	}

	var k int
	var seq int32
	n, err := binary.Decode(data[beforeLastNewLinePos:lastDelimiterPos], binary.BigEndian, &seq)
	if err != nil {
		return 0, -1, nil, err
	}
	k += n
	var decoded T
	n, err = binary.Decode(data[lastDelimiterPos:l-2], binary.BigEndian, &decoded)
	if err != nil {
		return 0, -1, nil, err
	}
	k += n

	return k, int(seq), &decoded, nil
}
