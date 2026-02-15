package immutxtdb

import "iter"

type page[K comparable, V any] struct {
	size    int
	number  int
	entries []*basicIdxEntry[K, V]
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

func (p page[K, V]) Entries() []*basicIdxEntry[K, V] {
	return p.entries
}

type paginer[K comparable, V any] struct {
	errChan      chan error
	pageSize     int
	preloadCount int
	laoded       []*page[K, V]
	current      int
	pushed       chan *basicIdxEntry[K, V]
	closed       bool
}

func (p *paginer[K, V]) buildPage(number int) *page[K, V] {
	k := 0
	var entries []*basicIdxEntry[K, V]
	for item := range p.pushed {
		if k >= p.pageSize {
			break
		}
		entries = append(entries, item)
		k++
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

func (p *paginer[K, V]) Close() {
	p.closed = true
	close(p.pushed)
	close(p.errChan)
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
	next := p.buildPage(p.current + 1)
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

func newPaginer[K comparable, V any](pageSize, preloadCount int, errChan chan error, pusher func(func(K, V) bool)) *paginer[K, V] {

	p := &paginer[K, V]{
		errChan:      errChan,
		pageSize:     pageSize,
		preloadCount: preloadCount,
		laoded:       make([]*page[K, V], 0),
		pushed:       make(chan *basicIdxEntry[K, V], pageSize*preloadCount),
		current:      -1,
	}
	go func() {
		pusher(func(k K, v V) bool {
			e := basicIdxEntry[K, V]{key: k, val: v}
			p.pushed <- &e
			return p.closed
		})
		// When all items were pushed close the channel
		close(p.pushed)
	}()

	return p
}
