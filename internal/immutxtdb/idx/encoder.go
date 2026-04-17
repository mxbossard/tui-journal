package idx

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

type Euid uint64

var NotMatchingEncoder = errors.New("encoder dos not match")
var NotAsciiText = errors.New("supplied text is out of ASCII table")

type KeyEncoder[K comparable] func(key K) ([]byte, error)
type ValEncoder[V any] func(val V) ([]byte, error)

type IdxEncoder interface {
	StateSize() int
	KeySize() int
	ValSize() int
	WordSize() int
	Header() []byte
	Match(header []byte) bool
	Setup(header []byte) error
	Encode(seq int, t time.Time, s State, key []byte, val []byte) ([]byte, error)
	// Decode first word in supplied byte slice.
	Decode(data []byte) (seq int, t time.Time, s State, key []byte, val []byte, err error)
	// Decode last word in supplied byte slice.
	DecodeLastWord(data []byte) (seq int, t time.Time, s State, key []byte, val []byte, err error)
	DecodeAll(order Order, data []byte, callback func(seq int, t time.Time, s State, key []byte, val []byte, err error) bool)
}

type basicIdxEncoder struct {
	IdxEncoder
	stateSize int
	keySize   int
	valSize   int
	uid       Euid
	version   int32
}

func NewBasicEncoder(uid Euid, version int32, stateSize, keySize, valSize int) *basicIdxEncoder {
	return &basicIdxEncoder{
		uid:       uid,
		version:   version,
		stateSize: stateSize,
		keySize:   keySize,
		valSize:   valSize,
	}
}

func (e basicIdxEncoder) StateSize() int {
	return int(e.stateSize)
}

func (e basicIdxEncoder) KeySize() int {
	return int(e.keySize)
}

func (e basicIdxEncoder) ValSize() int {
	return int(e.valSize)
}

func (e basicIdxEncoder) WordSize() int {
	return 8 + int(e.stateSize) + int(e.keySize) + int(e.valSize) + 4
}

func (e basicIdxEncoder) Header() []byte {
	b := make([]byte, e.WordSize())
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
	n, err = binary.Encode(b[k:k+4], binary.BigEndian, int32(e.keySize))
	if err != nil {
		panic(err)
	}
	k += n
	n, err = binary.Encode(b[k:k+4], binary.BigEndian, int32(e.valSize))
	if err != nil {
		panic(err)
	}
	k += n
	return b
}

func (e *basicIdxEncoder) Match(header []byte) bool {
	if len(header) < 12 {
		return false
	}
	k := 0
	var euid Euid
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

func (e *basicIdxEncoder) Setup(header []byte) error {
	if len(header) < 20 {
		return errors.New("need a header of 20 bytes minimum to setup")
	}

	if !e.Match(header) {
		return NotMatchingEncoder
	}
	k := 12
	var stateSize, keySize, valSize int32
	n, err := binary.Decode(header[k:k+4], binary.BigEndian, &stateSize)
	if err != nil {
		return err
	}
	k += n
	n, err = binary.Decode(header[k:k+4], binary.BigEndian, &keySize)
	if err != nil {
		return err
	}
	k += n
	n, err = binary.Decode(header[k:k+4], binary.BigEndian, &valSize)
	if err != nil {
		return err
	}
	k += n

	// fmt.Printf("setup config: statSize: %d ; dataSize: %d\n", stateSize, dataSize)

	e.stateSize = int(stateSize)
	e.keySize = int(keySize)
	e.valSize = int(valSize)
	return nil
}

func (e basicIdxEncoder) Encode(seq int, t time.Time, s State, k []byte, v []byte) ([]byte, error) {
	// if k == nil {
	// 	panic("key must not be nil")
	// }
	if len(k) > e.keySize {
		return nil, errors.New("key is longer than configured keySize")
	}

	// val := make([]byte, e.valSize)
	// if b, ok := any(v).([]byte); ok {
	// 	val = b
	// 	// } else if rv := reflect.ValueOf(v); !rv.IsNil() {
	// } else if e.valSerializer != nil {
	// 	var err error
	// 	err = e.valSerializer.Serialize(v, val)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }
	if len(v) > e.valSize {
		return nil, errors.New("data is longer than configured dataSize")
	}

	buf := make([]byte, e.WordSize())
	i := 0
	n, err := binary.Encode(buf[i:], binary.BigEndian, int32(seq))
	if err != nil {
		return nil, fmt.Errorf("encoding seq: %w", err)
	}
	i += n

	n, err = binary.Encode(buf[i:], binary.BigEndian, int32(t.Unix()/86400))
	if err != nil {
		return nil, fmt.Errorf("encoding time: %w", err)
	}
	i += n

	n, err = binary.Encode(buf[i:], binary.BigEndian, []byte(s))
	if err != nil {
		return nil, fmt.Errorf("encoding delimiterChar: %w", err)
	}
	i += e.stateSize
	// fmt.Printf("encoded state: %v\n", []byte(s))

	if len(k) > 0 {
		n, err = binary.Encode(buf[i:], binary.BigEndian, k)
		if err != nil {
			return nil, fmt.Errorf("encoding key: %w", err)
		}
		i += e.keySize
	}

	valLen := len(v)
	n, err = binary.Encode(buf[i:], binary.BigEndian, int32(valLen))
	if err != nil {
		return nil, fmt.Errorf("encoding value length: %w", err)
	}
	i += n

	if valLen > 0 {
		n, err = binary.Encode(buf[i:], binary.BigEndian, v)
		if err != nil {
			return nil, fmt.Errorf("encoding value: %w", err)
		}
		i += e.valSize
	}

	return buf, nil
}

func (e *basicIdxEncoder) Decode(buf []byte) (int, time.Time, State, []byte, []byte, error) {
	var seq, dataLen int32
	var s State
	var tim int32
	var t time.Time
	var key []byte
	var val []byte
	// fmt.Printf("decoding config: statSize: %d ; dataSize: %d\n", e.stateSize, e.dataSize)

	if len(buf) < e.WordSize() {
		return int(seq), t, s, key, val, fmt.Errorf("cannot decode data of length: %d < wordSize: %d", len(buf), e.WordSize())
	}

	k := 0
	n, err := binary.Decode(buf[k:k+4], binary.BigEndian, &seq)
	if err != nil {
		return int(seq), t, s, key, val, fmt.Errorf("decoding seq: %w", err)
	}
	k += n

	n, err = binary.Decode(buf[k:k+4], binary.BigEndian, &tim)
	if err != nil {
		return int(seq), t, s, key, val, fmt.Errorf("decoding time: %w", err)
	}
	k += 4
	t = time.Unix(int64(86400*tim), 0)

	var stateData = make([]byte, e.stateSize)
	n, err = binary.Decode(buf[k:k+e.stateSize], binary.BigEndian, &stateData)
	if err != nil {
		return int(seq), t, s, key, val, fmt.Errorf("decoding state: %w", err)
	}
	k += e.stateSize
	s = State(stateData)

	key = make([]byte, e.keySize)
	n, err = binary.Decode(buf[k:k+e.keySize], binary.BigEndian, &key)
	if err != nil {
		return int(seq), t, s, key, val, fmt.Errorf("decoding key: %w", err)
	}
	k += e.keySize

	n, err = binary.Decode(buf[k:k+4], binary.BigEndian, &dataLen)
	if err != nil {
		return int(seq), t, s, key, val, fmt.Errorf("decoding data length: %w", err)
	}
	k += n

	if dataLen > int32(e.valSize) {
		return int(seq), t, s, key, val, fmt.Errorf("bad encoded data length: %d", dataLen)
	}

	if dataLen > 0 {
		val = make([]byte, dataLen)
		n, err = binary.Decode(buf[k:k+int(dataLen)], binary.BigEndian, &val)
		if err != nil {
			return int(seq), t, s, key, val, fmt.Errorf("decoding value: %w", err)
		}
		k += e.valSize

		// if _, ok := any(val).([]byte); ok {
		// 	val = any(valData).(V)
		// } else if e.valSerializer != nil {
		// 	val, err = e.valSerializer.Deserialize(valData)
		// 	if err != nil {
		// 		return int(seq), s, key, val, fmt.Errorf("deserializing value: %w", err)
		// 	}
		// }
	}

	return int(seq), t, s, key, val, nil

}

func (e *basicIdxEncoder) DecodeLastWord(buf []byte) (int, time.Time, State, []byte, []byte, error) {
	lastWordStart := (len(buf)/e.WordSize() - 1) * e.WordSize()
	return e.Decode(buf[lastWordStart:])
}

func (e basicIdxEncoder) DecodeAll(order Order, buf []byte, push func(int, time.Time, State, []byte, []byte, error) bool) {
	wordSize := e.WordSize()
	if order == TopToBottom {
		for k := 0; k < len(buf); k += wordSize {
			seq, state, t, key, val, err := e.Decode(buf[k:])
			push(seq, state, t, key, val, err)
		}
	} else if order == BottomToTop {
		wordCount := len(buf) / wordSize
		for k := (wordCount - 1) * wordSize; k >= 0; k -= wordSize {
			seq, state, t, key, val, err := e.Decode(buf[k : k+wordSize])
			if !push(seq, state, t, key, val, err) {
				// Stop iterating
				break
			}
		}
	} else {
		panic(fmt.Sprintf("%s ordering not supported", order))
	}
}
