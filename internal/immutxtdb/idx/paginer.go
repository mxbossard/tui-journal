package idx

import (
	"iter"
)

type page[K any, V any] struct {
	size    int
	number  int
	entries []Entry[K, V]
	err     error
}

// The page size (max item count in the page)
func (p page[K, V]) Size() int {
	return p.size
}

// Number of item in the page
func (p page[K, V]) Len() int {
	return len(p.entries)
}

// First page is the 0 page.
func (p page[K, V]) Number() int {
	return p.number
}

// Return all entries
func (p page[K, V]) Entries() []Entry[K, V] {
	return p.entries
}

// Return errors
func (p page[K, V]) Err() error {
	return p.err
}

// Entries iterator
func (p *page[K, V]) All() iter.Seq2[int, Entry[K, V]] {
	return func(yield func(int, Entry[K, V]) bool) {
		for pos, e := range p.entries {
			if !yield(pos, e) {
				return
			}
		}
	}
}

type Paginer[K any, V any] interface {
	Close()
	Prev() (*page[K, V], bool, error)
	Next() (*page[K, V], bool, error)
	All() iter.Seq2[error, *page[K, V]]
}

type paginer[K any, V any] struct {
	Paginer[K, V]
	errChan      chan error
	pageSize     int
	preloadCount int
	loaded       []*page[K, V]
	current      int
	pushed       chan Entry[K, V]
	closed       bool
	endReached   bool
}

func (p *paginer[K, V]) buildPage(number int) *page[K, V] {
	if p.endReached {
		return nil
	}
	var entries []Entry[K, V]
	// fmt.Printf("building page #%d ...\n", number)
	var err error
	for item := range p.pushed {
		if item.Error() != nil {
			err = item.Error()
			break
		}
		entries = append(entries, item)
		if len(entries) >= p.pageSize {
			break
		}
	}
	page := &page[K, V]{
		size:    p.pageSize,
		number:  number,
		entries: entries,
		err:     err,
	}
	p.endReached = len(entries) < p.pageSize
	// fmt.Printf("built page #%d (count: %d)\n", number, len(entries))
	return page
}

func (p *paginer[K, V]) Close() {
	p.closed = true
	close(p.pushed)
	close(p.errChan)
	// fmt.Printf("paginer closed\n")
}

func (p *paginer[K, V]) Prev() (*page[K, V], bool, error) {
	if p.current <= 0 {
		panic("previous page does not exists")
		// return nil, false
	}
	p.current--
	current := p.loaded[p.current]
	return current, p.current > 0, current.Err()
}

func (p *paginer[K, V]) Next() (*page[K, V], bool, error) {
	if p.current >= len(p.loaded) {
		panic("next page does not exists")
		// return nil, false
	}
	p.current++

	// Load next page in advance
	next := p.buildPage(p.current + 1)
	if next != nil {
		p.loaded = append(p.loaded, next)
	}

	remaining := p.current < (len(p.loaded) - 1)
	current := p.loaded[p.current]
	// fmt.Printf("Next() current: %d ; loaded size: %d ; remaining: %v\n", p.current, len(p.loaded), remaining)
	// fmt.Printf("Returning page #%d ...\n", current.Number())
	return current, remaining, current.Err()

}

func (p *paginer[K, V]) All() iter.Seq2[error, *page[K, V]] {
	return func(yield func(error, *page[K, V]) bool) {
		for {
			page, ok, err := p.Next()
			if !yield(err, page) {
				return
			}
			if !ok {
				return
			}
		}
	}
}

func NewPaginer[K any, V any](pageSize, preloadPageCount int, pusher func(func(State, K, V, error) bool)) *paginer[K, V] {
	p := &paginer[K, V]{
		//errChan:      errChan,
		pageSize:     pageSize,
		preloadCount: preloadPageCount,
		loaded:       make([]*page[K, V], 0),
		pushed:       make(chan Entry[K, V], pageSize*preloadPageCount),
		current:      -1,
	}
	go func() {
		pusher(func(s State, k K, v V, err error) bool {
			e := BasicEntry[K, V]{key: k, val: v, state: s, err: err}
			p.pushed <- &e
			return !p.closed && err == nil
		})
		// When all items were pushed close the channel
		close(p.pushed)
	}()

	// Build first page
	first := p.buildPage(0)
	p.loaded = append(p.loaded, first)

	return p
}
