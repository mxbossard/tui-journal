package idx

import (
	"iter"
	"time"
)

const (
	LoaderPauseInterval = 100 * time.Millisecond
)

/*
Ideas:
- Wrapper around an iter.Seq[T]
  - caching data if eager
- LazyLoading, wait first range use to preloadCount
- EagerLoading, immediatly range and buffer preloadCount
*/

// Eagerly load wrapped iterator.
// count = 0 => wait first yield call to iterate (no chan)
// count = 1 => immediatly iterate over first item but wait for it's consumption (chan of size 0)
// count = 2 => immediatly iterate over first & second item then wait for there's consumption (chan of size 1)
func eagerLoader[V any](s iter.Seq[V], maxCount uint, pause func(v V) bool) iter.Seq[V] {
	var c chan V
	done := false
	if maxCount > 0 {
		c = make(chan V, maxCount-1)
		go func() {
			for v := range s {
				if done {
					break
				}
				c <- v
				// fmt.Printf("pushed %v\n", v)
				if pause != nil {
					for !done && pause(v) && len(c) > 0 {
						// If pause wait for chan to be empty
						// fmt.Printf("done: %v, v: %v, len(c): %v\n", done, v, len(c))
						time.Sleep(LoaderPauseInterval)
					}
				}
			}
			close(c)
			// fmt.Printf("chanel closed\n")
		}()
	}

	return func(yield func(V) bool) {
		if maxCount == 0 {
			// No chan use
			for v := range s {
				if !yield(v) {
					return
				}
			}
		} else {
			// Consume buffered chan
			for v := range c {
				if !yield(v) {
					done = true
					// fmt.Printf("breaked\n")
					return
				}
				// fmt.Printf("consumed %v\n", v)
			}
		}
	}
}

type entry[K, V any] struct {
	k K
	v V
}

func eagerLoader2[K, V any](s iter.Seq2[K, V], maxCount uint, pause func(k K, v V) bool) iter.Seq2[K, V] {
	var c chan entry[K, V]
	done := false
	if maxCount > 0 {
		c = make(chan entry[K, V], maxCount-1)
		go func() {
			for k, v := range s {
				if pause != nil {
					for pause(k, v) && len(c) >= 0 {
						// If fause wait for chan to be empty
						time.Sleep(LoaderPauseInterval)
					}
				}
				if done {
					return
				}
				c <- entry[K, V]{k: k, v: v}
			}
			close(c)
		}()
	}

	return func(yield func(K, V) bool) {
		if maxCount == 0 {
			// No chan use
			for k, v := range s {
				if !yield(k, v) {
					return
				}
			}
		} else {
			// Consume buffered chan
			for e := range c {
				if !yield(e.k, e.v) {
					done = true
					return
				}
			}
		}
	}
}

func PausingEagerLoader[T any](s iter.Seq[T], count uint, pause func(T) bool) iter.Seq[T] {
	return eagerLoader(s, count, pause)
}

func EagerLoader[T any](s iter.Seq[T], count uint) iter.Seq[T] {
	return PausingEagerLoader(s, count, nil)
}

// Lazy load wrapped iterator: trigger eager loading when iteration begins.
func PausingLazyLoader[T any](s iter.Seq[T], count uint, pause func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		preLoader := PausingEagerLoader(s, count, pause)
		for i := range preLoader {
			if !yield(i) {
				return
			}
		}
	}
}

func LazyLoader[T any](s iter.Seq[T], count uint) iter.Seq[T] {
	return PausingLazyLoader(s, count, nil)
}

func PausingLoader[T any](s iter.Seq[T], count uint, pause func(T) bool, lazy bool) iter.Seq[T] {
	if lazy {
		return PausingLazyLoader(s, count, pause)
	}
	return PausingEagerLoader(s, count, nil)
}

func Loader[T any](s iter.Seq[T], count uint, lazy bool) iter.Seq[T] {
	return PausingLoader(s, count, nil, lazy)
}

func PausingEagerLoader2[K, V any](s iter.Seq2[K, V], count uint, pause func(K, V) bool) iter.Seq2[K, V] {
	return eagerLoader2(s, count, pause)
}

func EagerLoader2[K, V any](s iter.Seq2[K, V], count uint) iter.Seq2[K, V] {
	return PausingEagerLoader2(s, count, nil)
}

// Lazy load wrapped iterator: trigger eager loading when iteration begins.
func PausingLazyLoader2[K, V any](s iter.Seq2[K, V], count uint, pause func(K, V) bool) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		preLoader := eagerLoader2(s, count, pause)
		for k, v := range preLoader {
			if !yield(k, v) {
				return
			}
		}
	}
}

func LazyLoader2[K, V any](s iter.Seq2[K, V], count uint) iter.Seq2[K, V] {
	return PausingLazyLoader2(s, count, nil)
}

func PausingLoader2[K, V any](s iter.Seq2[K, V], count uint, pause func(K, V) bool, lazy bool) iter.Seq2[K, V] {
	if lazy {
		return PausingLazyLoader2(s, count, pause)
	}
	return PausingEagerLoader2(s, count, pause)
}

func Loader2[K, V any](s iter.Seq2[K, V], count uint, lazy bool) iter.Seq2[K, V] {
	return PausingLoader2(s, count, nil, lazy)
}
