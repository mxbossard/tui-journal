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
