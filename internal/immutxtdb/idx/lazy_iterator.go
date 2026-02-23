package idx

import "iter"

/*
Ideas:
- Wrapper around an iter.Seq[T]
  - caching data if eager
*/
func LazyLoader[T any](preloadCount int) iter.Seq[T] {

}
