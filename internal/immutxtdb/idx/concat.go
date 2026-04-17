package idx

import (
	"cmp"
	"fmt"
	"iter"
	"sort"
	"sync"
	"time"

	"github.com/mxbossard/utilz/collectionz"
	"github.com/mxbossard/utilz/errorz"
)

type BasicIndexAggregate[K comparable, V any] struct {
	pageSize         int
	preloadPageCount int
	roIndexes        map[string]Index[K, V]
	rwIndexes        map[string]Index[K, V]
}

func Aggregate[K comparable, V any](pageSize, preloadPageCount int,
	roIndexes map[string]Index[K, V], rwIndexes map[string]Index[K, V]) *BasicIndexAggregate[K, V] {
	return &BasicIndexAggregate[K, V]{
		pageSize:         pageSize,
		preloadPageCount: preloadPageCount,
		roIndexes:        roIndexes,
		rwIndexes:        rwIndexes,
	}
}

// Add a KV entry
func (c BasicIndexAggregate[K, V]) Add(partition string, s State, t time.Time, key K, val V) (Entry[K, V], error) {
	idx, ok := c.rwIndexes[partition]
	if !ok {
		return nil, fmt.Errorf("index partition: %s not referenced", partition)
	}
	return idx.Add(s, t, key, val)
}

// Return last entry seq
// func (c BasicIndexCat[K, V]) LastSeq() (int, error) {

// }

func (c BasicIndexAggregate[K, V]) concat(compare func(a, b Entry[K, V]) int, f func(i Index[K, V]) (Paginer[K, V], error)) (Paginer[K, V], error) {
	var paginers []Paginer[K, V]
	for _, idx := range c.rwIndexes {
		p, err := f(idx)
		if err != nil {
			return nil, err
		}
		paginers = append(paginers, p)
	}
	for _, idx := range c.roIndexes {
		p, err := f(idx)
		if err != nil {
			return nil, err
		}
		paginers = append(paginers, p)
	}

	return CatPaginers(compare, c.pageSize, c.preloadPageCount, paginers...), nil

}

// FIXME: using a "Time Order" not "Positional Order (ex: TopToBottom)"
func EntryTimeCompare[K comparable, V any](order Order) func(a, b Entry[K, V]) int {
	return func(a, b Entry[K, V]) int {
		switch order {
		case TopToBottom:
			return cmp.Compare(a.Time().UnixMilli(), b.Time().UnixMilli())
		case BottomToTop:
			return cmp.Compare(b.Time().UnixMilli(), a.Time().UnixMilli())
		default:
			panic(fmt.Sprintf("no supported order: %s", order))
		}
	}
}

// ----- Browsing Methods -----
// Paginate all KV entries matching supplied key & Filter
func (c BasicIndexAggregate[K, V]) Filter(key K, order Order, f Filter) (Paginer[K, V], error) {
	return c.concat(EntryTimeCompare[K, V](order), func(i Index[K, V]) (Paginer[K, V], error) {
		return i.Filter(key, order, f)
	})
}

// Paginate all KV entries matching supplied key & Filter
// func (c BasicIndexAggregate[K, V]) HashedFilter(key K, order Order, f Filter) (Paginer[K, V], error) {
// 	return c.concat(EntryTimeCompare[K, V](order), func(i Index[K, V]) (Paginer[K, V], error) {
// 		return i.HashedFilter(key, order, f)
// 	})
// }

// Paginate all KV entries matching supplied Filter
func (c BasicIndexAggregate[K, V]) FilterAll(order Order, f Filter) (Paginer[K, V], error) {
	return c.concat(EntryTimeCompare[K, V](order), func(i Index[K, V]) (Paginer[K, V], error) {
		return i.FilterAll(order, f)
	})
}

// Paginate all KV entries exactly matching supplied key
func (c BasicIndexAggregate[K, V]) Paginate(key K, order Order) (Paginer[K, V], error) {
	return c.concat(EntryTimeCompare[K, V](order), func(i Index[K, V]) (Paginer[K, V], error) {
		return i.Paginate(key, order)
	})
}

// Paginate all KV entries matching supplied key which will be rotating hashed
// func (c BasicIndexAggregate[K, V]) HashedPaginate(key K, order Order) (Paginer[K, V], error) {
// 	return c.concat(EntryTimeCompare[K, V](order), func(i Index[K, V]) (Paginer[K, V], error) {
// 		return i.HashedPaginate(key, order)
// 	})
// }

// Paginate all KV entries
func (c BasicIndexAggregate[K, V]) PaginateAll(order Order) (Paginer[K, V], error) {
	return c.concat(EntryTimeCompare[K, V](order), func(i Index[K, V]) (Paginer[K, V], error) {
		return i.PaginateAll(order)
	})
}

// Return an iterator of all KV entries
func (c BasicIndexAggregate[K, V]) All(order Order) (iter.Seq[Entry[K, V]], error) {
	paginer, err := c.PaginateAll(order)
	if err != nil {
		return nil, err
	}
	return paginer.All(), nil
}

// call supplied f paginer function on all concatenated indexes
func (c BasicIndexAggregate[K, V]) concat0(order Order, f func(i Index[K, V]) (Paginer[K, V], error)) (Paginer[K, V], error) {
	stop := false
	chansVal := make(map[string]chan Entry[K, V])
	chans := &chansVal
	nextEntriesVal := make(map[string]Entry[K, V])
	nextEntries := &nextEntriesVal
	mutex := &sync.Mutex{}

	init := func() error {
		stop = false
		m := &sync.Mutex{}
		errs := make(chan error)
		wg := sync.WaitGroup{}
		paginers := make(map[string]Paginer[K, V])

		for q, idx := range c.rwIndexes {
			wg.Add(1)
			go func() {
				defer wg.Done()
				p, err := f(idx)
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
				p, err := f(idx)
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
			return agg
		}

		// Launch goroutines to stream entries
		for q, p := range paginers {
			// Make a buffered chan of size 1
			c := make(chan Entry[K, V], 1)
			mutex.Lock()
			(*chans)[q] = c
			mutex.Unlock()
			go func() {
				// Continuously push next entries
				defer close(c)
				for entry := range p.All() {
					c <- entry
					if stop {
						break
					}
				}
			}()
		}

		// Init nextEntries
		for q, c := range *chans {
			e, ok := <-c
			if ok {
				(*nextEntries)[q] = e
			} else {
				mutex.Lock()
				delete(*chans, q)
				mutex.Unlock()
			}
		}

		// fmt.Printf("cat paginer initialized chans: %d\n", len(*chans))
		return nil
	}

	err := init()
	if err != nil {
		return nil, err
	}

	// TODO implements paginers browsing into a new one order by time
	// Launch one goroutine by paginer to feed a chan
	// Consume one element from each chan
	// Keep elements pushing the one with the good order

	paginer := NewPaginer(c.pageSize, c.preloadPageCount, func(push func(e Entry[K, V]) bool) {
		// TODO updates idx if needed
		defer func() {
			stop = true
		}()
		// fmt.Printf("paginer chans: %d\n", len(*chans))
		partitions := collectionz.Keys(*nextEntries)
	End:
		for len(partitions) > 0 {
			var refTime *time.Time
			selectedPartition := ""
			sort.Strings(partitions)
			for _, q := range partitions {
				e, ok := (*nextEntries)[q]
				if !ok {
					// no more entries in chan
					delete(*nextEntries, q)
					selectedPartition = ""
				}
				// Select next entry to push partition (by time)
				switch order {
				case TopToBottom:
					if refTime == nil || refTime.After(e.Time()) {
						t := e.Time()
						refTime = &t
						selectedPartition = q
					}
				case BottomToTop:
					if refTime == nil || refTime.Before(e.Time()) {
						t := e.Time()
						refTime = &t
						selectedPartition = q
					}
				default:
					panic(fmt.Sprintf("no supported order: %s", order))
				}
			}

			if nextEntry, ok := (*nextEntries)[selectedPartition]; ok {
				if !push(nextEntry) {
					break End
				}
				delete(*nextEntries, selectedPartition)
			}

			mutex.Lock()
			c, ok := (*chans)[selectedPartition]
			mutex.Unlock()
			if ok {
				e, ok := <-c
				if ok {
					(*nextEntries)[selectedPartition] = e
				} else {
					mutex.Lock()
					delete(*chans, selectedPartition)
					mutex.Unlock()
				}
			}

			partitions = collectionz.Keys(*nextEntries)
		}
	})

	return paginer, err
}
