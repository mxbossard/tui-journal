package index

import (
	"crypto/sha512"
	"encoding/binary"
)

func RotatingHasher(salt []byte) func(int, []byte) ([]byte, error) {
	return func(pos int, b []byte) ([]byte, error) {
		hash := sha512.New()
		hash.Write(salt)
		err := binary.Write(hash, binary.BigEndian, pos)
		if err != nil {
			return nil, err
		}
		hash.Write(b)
		hashed := hash.Sum(nil)
		return hashed, nil
	}
}
