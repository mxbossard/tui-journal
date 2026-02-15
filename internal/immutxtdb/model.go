package immutxtdb

import (
	"iter"
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
	layers *paginer[string, *layer]
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

type idxEntry[K any, V any] interface {
	Key() K
	Val() V
}

type basicIdxEntry[K any, V any] struct {
	key K
	val V
}

func (e basicIdxEntry[K, V]) Key() K {
	return e.key
}

func (e basicIdxEntry[K, V]) Val() V {
	return e.val
}

type bucketIdxEntry basicIdxEntry[string, bool]
type layerIdxEntry basicIdxEntry[string, layer]

type page[K comparable, V any] struct {
	size    int
	number  int
	entries []idxEntry[K, V]
}

func (p page[K, V]) Size() int {
	return p.size
}

func (p page[K, V]) Len() int {
	return len(p.entries)
}

func (p page[K, V]) Offset() int {
	return p.number
}

func (p page[K, V]) Entries() []idxEntry[K, V] {
	return p.entries
}

type paginer[K comparable, V any] struct {
	errChan      chan error
	pageSize     int
	preloadCount int
	laoded       []*page[K, V]
	current      int
	loader       func(offset int) ([]idxEntry[K, V], error)
}

func (p *paginer[K, V]) loadPage(number int) *page[K, V] {
	entries, err := p.loader(number)
	if err != nil {
		p.errChan <- err
	}
	if len(entries) == 0 {
		return nil
	}

	page := &page[K, V]{
		size:    p.pageSize,
		number:  number,
		entries: entries,
	}
	return page
}

func (p *paginer[K, V]) Prev() (*page[K, V], bool) {
	if p.current <= 0 {
		return nil, false
	}
	p.current--
	return p.laoded[p.current], p.current > 0
}

func (p *paginer[K, V]) Next() (*page[K, V], bool) {
	// Load next page in advance
	next := p.loadPage(p.current + 1)
	if next != nil {
		p.laoded = append(p.laoded, next)
	}
	if p.current < len(p.laoded)-1 {
		p.current++
	}
	return p.laoded[p.current], p.current < len(p.laoded)-1
}

func (p *paginer[K, V]) All() iter.Seq[*page[K, V]] {
	panic("not implemented yet")
}

func newPaginer[K comparable, V any](pageSize, preloadCount int, errChan chan error, loader func(offset int) ([]idxEntry[K, V], error)) *paginer[K, V] {
	p := &paginer[K, V]{
		errChan:      errChan,
		pageSize:     pageSize,
		preloadCount: preloadCount,
		laoded:       make([]*page[K, V], 0),
		current:      -1,
		loader:       loader,
	}

	// TODO: preloading in a goroutine
	p.laoded = append(p.laoded, p.loadPage(0))
	return p
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
