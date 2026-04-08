package idx

import (
	"errors"
	"iter"
)

type page[K comparable, V any] struct {
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

// Entries iterator return (position in page, entry[K, V])
func (p *page[K, V]) All() iter.Seq2[int, Entry[K, V]] {
	return func(yield func(int, Entry[K, V]) bool) {
		for pos, e := range p.entries {
			if !yield(pos, e) {
				return
			}
		}
	}
}

type Paginer[K comparable, V any] interface {
	Close()
	Reset() error
	Prev() (*page[K, V], bool, error)
	Next() (*page[K, V], bool, error)
	Pages() iter.Seq2[error, *page[K, V]]
	All() iter.Seq[Entry[K, V]]
}

type paginer[K comparable, V any] struct {
	Paginer[K, V]
	// errChan      chan error
	pageSize     int
	preloadCount int
	loaded       []*page[K, V]
	current      int
	pushed       chan Entry[K, V]
	closed       bool
	endReached   bool
}

// Build a page, attempt to complete it all (load page size items)
func (p *paginer[K, V]) buildPage(number int) *page[K, V] {
	if p.endReached {
		return nil
	}
	var entries []Entry[K, V]
	// fmt.Printf("building page #%d ...\n", number)
	var err error
	// fmt.Printf("pushed entries: %d\n", len(*p.pushed))
	for entry := range p.pushed {
		if entry.Error() != nil {
			err = entry.Error()
			break
		}
		// fmt.Printf("adding entry %s in page %d ...\n", entry, number)
		entries = append(entries, entry)
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
	// close(*p.pushed)
	// close(p.errChan)
	// fmt.Printf("paginer closed\n")
}

func (p *paginer[K, V]) Reset() error {
	p.current = -1
	// empty the channel
	// for len(p.pushed) > 0 {
	// 	<-p.pushed
	// }
	// c := make(chan Entry[K, V], p.pageSize*p.preloadCount)
	// p.pushed = &c
	return nil
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

var ErrNotExist = errors.New("page do not exists")

func (p *paginer[K, V]) Next() (*page[K, V], bool, error) {
	if p.current+1 >= len(p.loaded) {
		// panic("next page does not exists")
		// panic(fmt.Sprintf("next page #%d does not exists", p.current+1))
		// return nil, false
		return nil, false, ErrNotExist
	}
	p.current++

	// Load next page in advance
	next := p.buildPage(p.current + 1)
	if next != nil {
		p.loaded = append(p.loaded, next)
	}

	remaining := p.current+1 < len(p.loaded)
	current := p.loaded[p.current]
	// fmt.Printf("Next() current: %d ; loaded count: %d ; remaining: %v\n", p.current, len(p.loaded), remaining)
	// fmt.Printf("Returning page #%d ...\n", current.Number())
	return current, remaining, current.Err()

}

func (p *paginer[K, V]) Pages() iter.Seq2[error, *page[K, V]] {
	return func(yield func(error, *page[K, V]) bool) {
		p.Reset()
		for {
			page, ok, err := p.Next()
			if !yield(err, page) {
				return
			}
			if !ok {
				// No more pages
				return
			}
		}
	}
}

func (p *paginer[K, V]) All() iter.Seq[Entry[K, V]] {
	return func(yield func(Entry[K, V]) bool) {
		p.Reset()
		for {
			page, ok, err := p.Next()
			if err != nil {
				if !yield(NewErrEntry[K, V](err)) {
					return
				}
				// Stop page iteration on error
				return
			}
			for _, entry := range page.All() {
				// fmt.Printf("supplying entry %s in pagine %d ...\n", entry, page.number)
				if !yield(entry) {
					return
				}
			}
			if !ok {
				// No more pages
				return
			}
		}
	}
}

// Build a Paginer.
// Current implem preload all page items (pageSize & preloadPageCount are important).
// pusher func must be implemented to push each items to paginer using push function.
// if push function return false pusher func MUST stop.
func NewPaginer[K comparable, V any](pageSize, preloadPageCount int,
	pusher func(push func(e Entry[K, V]) bool)) Paginer[K, V] {
	p := &paginer[K, V]{
		pageSize:     pageSize,
		preloadCount: preloadPageCount,
		loaded:       make([]*page[K, V], 0),
		pushed:       make(chan Entry[K, V], pageSize*preloadPageCount),
		current:      -1,
	}

	go func() {
		pusher(func(e Entry[K, V]) bool {
			p.pushed <- e
			// fmt.Printf("pushed entry %s in paginer\n", e)
			return !p.closed && e.Error() == nil
		})
		// When all items were pushed close the channel
		close(p.pushed)
	}()

	// Build first page
	first := p.buildPage(0)
	p.loaded = append(p.loaded, first)

	return p
}

// Updatable Paginer may be updated with new data when pusher func is terminated.
func NewUpdatablePaginer[K comparable, V any](pageSize, preloadPageCount int, pusher func(push func(e Entry[K, V]) bool)) *paginer[K, V] {
	// - if rerunning pusher should keep a state to not repush everything ?
	// - change pusher ?
	// - use a new updater callback ?
	// HOW TO ?
	panic("not implemented yet")
}
