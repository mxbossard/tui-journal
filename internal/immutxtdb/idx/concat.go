package idx

import (
	"fmt"
	"iter"
	"sort"
	"sync"
	"time"

	"github.com/mxbossard/utilz/collectionz"
	"github.com/mxbossard/utilz/errorz"
)

type BasicIndexCat[K comparable, V any] struct {
	pageSize         int
	preloadPageCount int
	roIndexes        map[string]Index[K, V]
	rwIndexes        map[string]Index[K, V]
}

// Add a KV entry
func (c BasicIndexCat[K, V]) Add(qualifier string, s State, t time.Time, key K, val V) (Entry[K, V], error) {
	idx, ok := c.rwIndexes[qualifier]
	if !ok {
		return nil, fmt.Errorf("index qualifier: %s not referenced", qualifier)
	}
	return idx.Add(s, t, key, val)
}

// Return last entry seq
// func (c BasicIndexCat[K, V]) LastSeq() (int, error) {

// }

// ----- Browsing Methods -----
// Paginate all KV entries matching supplied key & Filter
func (c BasicIndexCat[K, V]) Filter(key K, order Order, f Filter) (Paginer[K, V], error) {
	paginers := make(map[string]Paginer[K, V])
	m := &sync.Mutex{}
	errs := make(chan error)
	wg := sync.WaitGroup{}
	for q, idx := range c.rwIndexes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p, err := idx.Filter(key, order, f)
			if err != nil {
				errs <- err
			} else {
				m.Lock()
				paginers[q] = p
				m.Unlock()
			}
		}()
	}
	for q, idx := range c.roIndexes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p, err := idx.Filter(key, order, f)
			if err != nil {
				errs <- err
			} else {
				m.Lock()
				paginers[q] = p
				m.Unlock()
			}
		}()
	}
	wg.Wait()

	if agg := errorz.ChanCollect(errs); agg.Got() {
		return nil, agg
	}

	// TODO implements paginers browsing into a new one order by time
	// Launch one goroutine by paginer to feed a chan
	// Consume one element from each chan
	// Keep elements pushing the one with the good order

	stop := false
	chans := make(map[string]chan Entry[K, V])
	for q, p := range paginers {
		// Make a buffered chan of size 1
		c := make(chan Entry[K, V], 1)
		chans[q] = c
		go func() {
			defer close(c)
			for entry := range p.All() {
				chans[q] <- entry
				if stop {
					break
				}
			}
		}()
	}

	paginer := NewPaginer(c.pageSize, c.preloadPageCount, func(push func(e Entry[K, V]) bool) {
		nextEntries := make(map[string]Entry[K, V])
		// Init nextEntries
		for q, c := range chans {
			e, ok := <-c
			if ok {
				nextEntries[q] = e
			} else {
				delete(chans, q)
			}
		}

		qualifiers := collectionz.Keys(nextEntries)
	End:
		for len(qualifiers) > 0 {
			var refTime *time.Time
			selectedQualifier := ""
			sort.Strings(qualifiers)
			for _, q := range qualifiers {
				e, ok := nextEntries[q]
				if !ok {
					// no more entries in chan
					delete(nextEntries, q)
					selectedQualifier = ""
				}
				// Select next entry to push qualifier (by time)
				switch order {
				case TopToBottom:
					if refTime == nil || refTime.After(e.Time()) {
						t := e.Time()
						refTime = &t
						selectedQualifier = q
					}
				case BottomToTop:
					if refTime == nil || refTime.Before(e.Time()) {
						t := e.Time()
						refTime = &t
						selectedQualifier = q
					}
				default:
					panic(fmt.Sprintf("no supported order: %s", order))
				}
			}

			if nextEntry, ok := nextEntries[selectedQualifier]; ok {
				if !push(nextEntry) {
					break End
				}
				delete(nextEntries, selectedQualifier)
			}

			if c, ok := chans[selectedQualifier]; ok {
				e, ok := <-c
				if ok {
					nextEntries[selectedQualifier] = e
				} else {
					delete(chans, selectedQualifier)
				}
			}
			qualifiers = collectionz.Keys(nextEntries)
		}
	})

	return paginer, nil
}

// Paginate all KV entries matching supplied key & Filter
func (c BasicIndexCat[K, V]) HashedFilter(key K, order Order, f Filter) (Paginer[K, V], error) {
	panic("not implemented yet")
}

// Paginate all KV entries matching supplied Filter
func (c BasicIndexCat[K, V]) FilterAll(order Order, f Filter) (Paginer[K, V], error) {
	panic("not implemented yet")
}

// Paginate all KV entries exactly matching supplied key
func (c BasicIndexCat[K, V]) Paginate(key K, order Order) (Paginer[K, V], error) {
	panic("not implemented yet")
}

// Paginate all KV entries matching supplied key which will be rotating hashed
func (c BasicIndexCat[K, V]) HashedPaginate(key K, order Order) (Paginer[K, V], error) {
	panic("not implemented yet")
}

// Paginate all KV entries
func (c BasicIndexCat[K, V]) PaginateAll(order Order) (Paginer[K, V], error) {
	panic("not implemented yet")
}

// Return an iterator of all KV entries
func (c BasicIndexCat[K, V]) All(order Order) (iter.Seq2[error, Entry[K, V]], error) {
	panic("not implemented yet")
}
