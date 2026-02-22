package idx

import (
	"encoding/binary"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
)

var (
	voidEncoderEuid  = Euid(binary.BigEndian.Uint64([]byte("void0000")))
	bytesEncoderEuid = Euid(binary.BigEndian.Uint64([]byte("bytes000")))
	asciiEncoderEuid = Euid(binary.BigEndian.Uint64([]byte("ascii000")))
)

func NewVoidEncoder(version int32, stateSize, keySize, valSize int) IdxEncoder[Void] {
	return NewAbstractEncoder[Void](voidEncoderEuid, version, stateSize, keySize, valSize, nil)
}

func NewByteSliceEncoder(version int32, stateSize, keySize, valSize int) IdxEncoder[[]byte] {
	return NewAbstractEncoder[[]byte](bytesEncoderEuid, version, stateSize, keySize, valSize, nil)
}

func NewByteArrayEncoder(version int32, stateSize, keySize, valSize int) IdxEncoder[[128]byte] {
	return NewAbstractEncoder[[128]byte](bytesEncoderEuid, version, stateSize, keySize, valSize, nil)
}

func NewAsciiEncoder(version int32, stateSize, keySize, valSize int) IdxEncoder[string] {
	return NewAbstractEncoder(asciiEncoderEuid, version, stateSize, keySize, valSize, serialize.AsciiSerializer{})
}
