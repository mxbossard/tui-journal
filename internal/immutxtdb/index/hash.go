package index

import (
	"crypto/sha512"
	"encoding/binary"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
)

func RotatingHasher(salt []byte, size int) func(int, []byte) ([]byte, error) {
	return func(pos int, b []byte) ([]byte, error) {
		hash := sha512.New()
		hash.Write(salt)
		fixedPos := int32(pos)
		err := binary.Write(hash, binary.BigEndian, fixedPos)
		if err != nil {
			return nil, err
		}
		if len(b) != size {
			b = idx.FixedSizeByteSlice(size, b)
		}
		hash.Write(b)
		hashed := hash.Sum(nil)
		fixedSizeHash := idx.FixedSizeByteSlice(size, hashed)
		return fixedSizeHash, nil
	}
}
