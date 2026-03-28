package idx

import (
	"fmt"
	"time"
)

type Entry[K comparable, V any] interface {
	Key() K
	BytesKey() []byte
	Val() V
	Seq() int
	Time() time.Time
	State() State
	Error() error
}

type BasicEntry[K comparable, V any] struct {
	key      *K
	val      *V
	seq      int
	time     time.Time
	state    State
	err      error
	bytesKey []byte
}

func NewEntry[K comparable, V any](key K, val V, seq int, time time.Time, state State, err error, bKey []byte) *BasicEntry[K, V] {
	e := &BasicEntry[K, V]{
		key:      &key,
		val:      &val,
		seq:      seq,
		time:     time,
		state:    state,
		err:      err,
		bytesKey: bKey,
	}
	return e
}

func NewErrEntry[K comparable, V any](err error) *BasicEntry[K, V] {
	e := &BasicEntry[K, V]{
		seq: -1,
		err: err,
	}
	return e
}

func (e BasicEntry[K, V]) String() string {
	var key K
	if e.key != nil {
		key = *e.key
	}
	return fmt.Sprintf("Entry(#%d)[%v, %s]", e.seq, key, e.val)
}

func (e BasicEntry[K, V]) Key() K {
	if e.key != nil {
		return *e.key
	}
	var k K
	return k
}

func (e BasicEntry[K, V]) BytesKey() []byte {
	return e.bytesKey
}

func (e BasicEntry[K, V]) Val() V {
	if e.val != nil {
		return *e.val
	}
	var v V
	return v
}

func (e BasicEntry[K, V]) Seq() int {
	return e.seq

}

func (e BasicEntry[K, V]) Time() time.Time {
	return e.time
}

func (e BasicEntry[K, V]) State() State {
	return e.state
}

func (e BasicEntry[K, V]) Error() error {
	return e.err
}
