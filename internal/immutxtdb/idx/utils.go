package idx

import "fmt"

func FixedSizeString(s int, k string) []byte {
	b := make([]byte, s)
	n := copy(b, []byte(k))
	if n > s {
		panic(fmt.Sprintf("string too long for fixed size: %d", s))
	}
	return b
}

func FixedSizeByteSlice(s int, k []byte) []byte {
	if len(k) == s {
		return k
	}
	b := make([]byte, s)
	n := copy(b, k)
	if n > s {
		panic(fmt.Sprintf("byte slice too long for fixed size: %d", s))
	}
	return b
}
