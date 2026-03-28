package idx

import (
	"encoding/binary"
	"strings"
)

type State []byte

var dummyState = BuildState(8, "dummy")

func BuildState(size int, s ...string) State {
	data := make([]byte, size)
	_, err := binary.Encode(data, binary.BigEndian, []byte(strings.Join(s, "")))
	if err != nil {
		panic(err)
	}
	// fmt.Printf("built state of size: %d with strings: %v => %v\n", size, s, data)
	return State(data)
}

func CatState(size int, states ...State) State {
	data := make([]byte, size)
	k := 0
	for _, s := range states {
		copy(data[k:], s)
		k += len(s)
	}
	// fmt.Printf("cat state of size: %d with states: %v => %v\n", size, states, data)
	return State(data)
}
