package hash

import (
	"crypto/sha512"
	"encoding/binary"

	"github.com/mxbossard/utilz/bytez"
)

// type RotatingHasher func(int, []byte) ([]byte, error)
type GlidingHasher func(int, []byte) ([]byte, error)
type KeyGlidingHasher[K comparable] func(int, K) (K, error)

func NewRotatingHasher(salt []byte, size int) GlidingHasher {
	return func(pos int, b []byte) ([]byte, error) {
		hash := sha512.New()
		hash.Write(salt)
		fixedPos := int32(pos)
		err := binary.Write(hash, binary.BigEndian, fixedPos)
		if err != nil {
			return nil, err
		}
		if len(b) != size {
			b = bytez.FixedSizeByteSlice(size, b)
		}
		hash.Write(b)
		hashed := hash.Sum(nil)
		fixedSizeHash := bytez.FixedSizeByteSlice(size, hashed)
		return fixedSizeHash, nil
	}
}
