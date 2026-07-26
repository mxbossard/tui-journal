package idx

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx/idxrepo"
)

type State = idxrepo.State

var dummyState = BuildStringState(8, "dummy")

func BuildStringState(size int, s ...string) State {
	data := make([]byte, size)
	_, err := binary.Encode(data, binary.BigEndian, []byte(strings.Join(s, "")))
	if err != nil {
		panic(err)
	}
	// fmt.Printf("built state of size: %d with strings: %v => %v\n", size, s, data)
	return State(data)
}

func BuildState(flags ...byte) State {
	m := make(map[byte]bool)
	data := make([]byte, 1)
	for _, f := range flags {
		if f != 1 && f != 2 && f != 4 && f != 8 && f != 16 && f != 32 && f != 64 && f != 128 {
			panic(fmt.Sprintf("bad flag %x supplied", f))
		}
		if m[f] {
			panic(fmt.Sprintf("flag %x supplied twice", f))
		}
		m[f] = true
		data[0] += f
	}
	return State(data)
}

func MatchFlag(state byte, f byte) bool {
	if f != 1 && f != 2 && f != 4 && f != 8 && f != 16 && f != 32 && f != 64 && f != 128 {
		panic(fmt.Sprintf("bad flag %x supplied", f))
	}

	return state&f == f
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
