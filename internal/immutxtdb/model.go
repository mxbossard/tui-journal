package immutxtdb

import (
	"time"
)

type IdxOrder int

const (
	TopToBottom IdxOrder = iota
	BottomToTop
)

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

type layer struct {
	content  string
	metadata *Metadata
}

type Bucket struct {
	db     *DB
	uid    string
	layers []*layer
	//layers *paginer[string, *layer]
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

type idxEntry[K comparable, V any] interface {
	Key() K
	Val() V
}

type basicIdxEntry[K comparable, V any] struct {
	key K
	val V
	err error
}

func (e basicIdxEntry[K, V]) Key() K {
	return e.key
}

func (e basicIdxEntry[K, V]) Val() V {
	return e.val
}

func (e basicIdxEntry[K, V]) Error() error {
	return e.err
}

type bucketIdxEntry basicIdxEntry[string, bool]
type layerIdxEntry basicIdxEntry[string, layer]

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
