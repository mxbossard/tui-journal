package model

import (
	"encoding/binary"
	"iter"
	"time"
)

type Order int

const (
	TopToBottom Order = iota
	BottomToTop
)

type State []byte

func BuildState(size int, s string) State {
	data := make([]byte, size)
	_, err := binary.Encode(data, binary.BigEndian, []byte(s))
	if err != nil {
		panic(err)
	}
	return State(data)
}

type Index[K any, V any] interface {
	Add(key K, val V) error
	Paginate(key K, order Order, pageSize int) (*paginer[K, V], chan error)
	PaginateAll(order Order, pageSize int) (*paginer[K, V], chan error)
	All(order Order, errChan chan error) iter.Seq2[K, V]
	Count() (int, error)
}

type Labels map[string]string

type Metadata struct {
	version    int
	created    time.Time
	updated    time.Time
	labels     Labels
	commited   bool
	snapshoted bool
}

type Document struct {
	content  string
	metadata *Metadata
}

type LayerRef struct {
	blocsFilepath string
	blocId        int
	state         State
}

func NewLayerRef(blocsFilepath string, blocId int, state State) *LayerRef {
	return &LayerRef{blocsFilepath: blocsFilepath, blocId: blocId, state: state}
}

type Layer struct {
	content  string
	metadata *Metadata
}

type Bucket struct {
	//db     *DB
	uid    string
	layers []*LayerRef
	//layers *paginer[string, *layer]
}

func NewBucket(uid string, layers []*LayerRef) *Bucket {
	return &Bucket{
		uid:    uid,
		layers: layers,
	}
}

func (b Bucket) Project() (Document, error) {
	panic("not implemented yet")
}

func (b Bucket) Save(content string, labels Labels) error {
	panic("not implemented yet")
}

func (b Bucket) Commit() error {
	panic("not implemented yet")
}

func (b Bucket) Squash() error {
	panic("not implemented yet")
}

// type cursor[K comparable, V any] struct {
// 	pageSize int
// }

// // func (c cursor[K, V]) HasNext() bool {
// // 	panic("not implemented yet")
// // }

// // func (c *cursor[K, V]) Next() (page[K, V], error) {
// // 	panic("not implemented yet")
// // }

// func (c *cursor[K, V]) preLoad() error {
// 	panic("not implemented yet")
// }

// func (c *cursor[K, V]) All(errChan chan error) iter.Seq2[K, V] {
// 	panic("not implemented yet")
// }
