package idx

import (
	"iter"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intIterator(min, max int, callback func(int)) iter.Seq[int] {
	return func(yield func(int) bool) {
		for k := min; k < max; k++ {
			callback(k)
			if !yield(k) {
				return
			}
		}
	}
}

func TestLoader_EagerLoader_0(t *testing.T) {
	var k uint
	it := intIterator(0, 10, func(i int) {
		k++
	})

	expectedPreloadCount := uint(0)

	// Nothing should be preloaded yet
	assert.Equal(t, uint(0), k)

	elIt := EagerLoader(it, expectedPreloadCount)
	require.NotNil(t, elIt)

	time.Sleep(10 * time.Millisecond)
	// Iterator should be preloaded
	assert.Equal(t, uint(0), k)

	elIt(func(n int) bool {
		assert.Equal(t, 0, n)
		return false
	})

	time.Sleep(10 * time.Millisecond)
	// Iterator should be preloaded
	assert.Equal(t, expectedPreloadCount+1, k)
}

func TestLoader_EagerLoader_1(t *testing.T) {
	var k uint
	it := intIterator(0, 10, func(i int) {
		k++
	})

	expectedPreloadCount := uint(1)

	// Nothing should be preloaded yet
	assert.Equal(t, uint(0), k)

	elIt := EagerLoader(it, expectedPreloadCount)
	require.NotNil(t, elIt)

	time.Sleep(10 * time.Millisecond)
	// Iterator should be preloaded
	assert.Equal(t, uint(1), k)

	elIt(func(n int) bool {
		assert.Equal(t, 0, n)
		return false
	})

	time.Sleep(10 * time.Millisecond)
	// Iterator should be preloaded
	assert.Equal(t, expectedPreloadCount+1, k)
}

func TestLoader_EagerLoader_5(t *testing.T) {
	var k uint
	it := intIterator(0, 10, func(i int) {
		k++
	})

	expectedPreloadCount := uint(5)

	// Nothing should be preloaded yet
	assert.Equal(t, uint(0), k)

	elIt := EagerLoader(it, expectedPreloadCount)
	require.NotNil(t, elIt)

	time.Sleep(10 * time.Millisecond)
	// Iterator should be preloaded
	assert.Equal(t, expectedPreloadCount, k)

	elIt(func(n int) bool {
		assert.Equal(t, 0, n)
		return false
	})

	time.Sleep(10 * time.Millisecond)
	// Iterator should be preloaded
	assert.Equal(t, expectedPreloadCount+1, k)
}

func TestLoader_PausingEagerLoader(t *testing.T) {
	var k uint
	it := intIterator(0, 10, func(i int) {
		k++
	})

	expectedPreloadCount := uint(3)

	// Nothing should be preloaded yet
	assert.Equal(t, uint(0), k)

	pause := func(v int) bool {
		return v != 0 && uint(v)%(expectedPreloadCount) == 2
	}
	elIt := PausingEagerLoader(it, 5, pause)
	require.NotNil(t, elIt)

	time.Sleep(10 * time.Millisecond)
	// Iterator should be preloaded
	assert.Equal(t, expectedPreloadCount, k)

	i := uint(0)
	elIt(func(n int) bool {
		assert.Equal(t, min(10, expectedPreloadCount*(1+i/expectedPreloadCount)), k, "for n:%d", n)
		assert.Equal(t, i, uint(n), "for n:%d", n)
		i++
		time.Sleep(10 * time.Millisecond)
		return true
	})

}

func TestLoader_LazyLoader_0(t *testing.T) {
	var k uint
	it := intIterator(0, 10, func(i int) {
		k++
	})

	expectedPreloadCount := uint(0)

	// Nothing should be preloaded yet
	assert.Equal(t, uint(0), k)

	elIt := LazyLoader(it, expectedPreloadCount)
	require.NotNil(t, elIt)

	time.Sleep(10 * time.Millisecond)
	// Iterator should not be preloaded
	assert.Equal(t, uint(0), k)

	elIt(func(n int) bool {
		assert.Equal(t, 0, n)
		return false
	})

	time.Sleep(10 * time.Millisecond)
	// Iterator should not be preloaded
	assert.Equal(t, uint(1), k)
}

func TestLoader_LazyLoader_1(t *testing.T) {
	var k uint
	it := intIterator(0, 10, func(i int) {
		k++
	})

	expectedPreloadCount := uint(1)

	// Nothing should be preloaded yet
	assert.Equal(t, uint(0), k)

	elIt := LazyLoader(it, expectedPreloadCount)
	require.NotNil(t, elIt)

	time.Sleep(10 * time.Millisecond)
	// Iterator should not be preloaded
	assert.Equal(t, uint(0), k)

	elIt(func(n int) bool {
		assert.Equal(t, 0, n)
		return false
	})

	time.Sleep(10 * time.Millisecond)
	// Iterator should be preloaded
	assert.Equal(t, expectedPreloadCount+1, k)
}

func TestLoader_LazyLoader_5(t *testing.T) {
	var k uint
	it := intIterator(0, 10, func(i int) {
		k++
	})

	expectedPreloadCount := uint(5)

	// Nothing should be preloaded yet
	assert.Equal(t, uint(0), k)

	elIt := LazyLoader(it, expectedPreloadCount)
	require.NotNil(t, elIt)

	time.Sleep(10 * time.Millisecond)
	// Iterator should not be preloaded
	assert.Equal(t, uint(0), k)

	elIt(func(n int) bool {
		assert.Equal(t, 0, n)
		return false
	})

	time.Sleep(10 * time.Millisecond)
	// Iterator should be preloaded
	assert.Equal(t, expectedPreloadCount+1, k)
}
