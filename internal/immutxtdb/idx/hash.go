package idx

import (
	"crypto/sha512"
	"encoding/binary"
)

func NewRotatingHasher(salt []byte, size int) func(int, []byte) ([]byte, error) {
	return func(pos int, b []byte) ([]byte, error) {
		hash := sha512.New()
		hash.Write(salt)
		fixedPos := int32(pos)
		err := binary.Write(hash, binary.BigEndian, fixedPos)
		if err != nil {
			return nil, err
		}
		if len(b) != size {
			b = FixedSizeByteSlice(size, b)
		}
		hash.Write(b)
		hashed := hash.Sum(nil)
		fixedSizeHash := FixedSizeByteSlice(size, hashed)
		return fixedSizeHash, nil
	}
}
