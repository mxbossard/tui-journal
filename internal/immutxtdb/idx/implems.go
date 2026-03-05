package idx

import (
	"encoding/binary"
)

var (
	voidEncoderEuid  = Euid(binary.BigEndian.Uint64([]byte("void0000")))
	bytesEncoderEuid = Euid(binary.BigEndian.Uint64([]byte("bytes000")))
	asciiEncoderEuid = Euid(binary.BigEndian.Uint64([]byte("ascii000")))
)

func NewVoidEncoder(version int32, stateSize, keySize, valSize int) IdxEncoder {
	return NewBasicEncoder(voidEncoderEuid, version, stateSize, keySize, valSize)
}

func NewByteSliceEncoder(version int32, stateSize, keySize, valSize int) IdxEncoder {
	return NewBasicEncoder(bytesEncoderEuid, version, stateSize, keySize, valSize)
}

func NewByteArrayEncoder(version int32, stateSize, keySize, valSize int) IdxEncoder {
	return NewBasicEncoder(bytesEncoderEuid, version, stateSize, keySize, valSize)
}

func NewAsciiEncoder(version int32, stateSize, keySize, valSize int) IdxEncoder {
	return NewBasicEncoder(asciiEncoderEuid, version, stateSize, keySize, valSize)
}
