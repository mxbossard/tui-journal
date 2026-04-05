package idx

import (
	"fmt"
	"testing"
	"time"

	"github.com/mxbossard/utilz/ptrz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginer_Empty(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 2
	p := NewPaginer(expectedPageSize, expectedPreloadCount, func(push func(Entry[int, string]) bool) {
		// Nothing to paginate
	})
	require.NotNil(t, p)

	page, ok, err := p.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, expectedPageSize, page.Size())
	assert.Equal(t, 0, page.Len())
	assert.Empty(t, page.Entries())
}

func TestPaginer_Pages(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 2
	var expectedCount *int
	expectedCount = ptrz.IntPtr(10)

	k := 0
	var expectedMessages []string
	p := NewPaginer(expectedPageSize, expectedPreloadCount, func(push func(Entry[int, string]) bool) {
		for {
			msg := fmt.Sprintf("msg%d", k)
			expectedMessages = append(expectedMessages, msg)
			// fmt.Printf("pushing msg: [%s] ...\n", msg)
			now := time.Now()
			e := NewEntry(k, msg, k, now, nil, nil, nil)
			if !push(e) {
				// fmt.Printf("breaked!\n")S
				break
			}
			k++
			if k >= *expectedCount {
				// End source
				// fmt.Printf("source end reached\n")
				break
			}
		}
	})
	require.NotNil(t, p)

	i := 0
	for err, page := range p.Pages() {
		assert.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, i, page.Number(), "bad page number")
		i++
	}
	assert.Equal(t, *expectedCount/expectedPageSize+1, i, "bad page count")

	// 2 consecutive operations should works
	i = 0
	for err, page := range p.Pages() {
		assert.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, i, page.Number(), "bad page number")
		i++
	}
	assert.Equal(t, *expectedCount/expectedPageSize+1, i, "bad page count")

	// 3 add data to paginer NOT IMPLEMENTED
	// expectedCount = ptrz.IntPtr(*expectedCount + expectedPageSize)
	// i = 0
	// for err, page := range p.Pages() {
	// 	assert.NoError(t, err)
	// 	require.NotNil(t, p)
	// 	assert.Equal(t, i, page.Number(), "bad page number")
	// 	i++
	// }
	// assert.Equal(t, *expectedCount/expectedPageSize+1, i, "bad page count")
}

func TestPaginer_Next(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 2
	expectedCount := 10

	a := 5
	b := &a
	assert.Equal(t, a, *b, "int ptr is a bad idea")

	k := 0
	var expectedMessages []string
	p := NewPaginer(expectedPageSize, expectedPreloadCount, func(push func(Entry[int, string]) bool) {
		for {
			msg := fmt.Sprintf("msg%d", k)
			expectedMessages = append(expectedMessages, msg)
			now := time.Now()
			e := NewEntry(k, msg, k, now, []byte("foobarba"), nil, nil)
			// fmt.Printf("pushing entry: [%s] ...\n", e)
			if !push(e) {
				// fmt.Printf("breaked!\n")
				break
			}
			k++
			if k >= expectedCount {
				// End source
				// fmt.Printf("source end reached\n")
				break
			}
		}
	})
	require.NotNil(t, p)
	// assert.Equal(t, expectedPageSize*expectedPreloadCount, k)

	i := 0
	n := 0
	for {
		page, ok, err := p.Next()
		assert.NoError(t, err)
		require.NotNil(t, page)
		assert.Equal(t, i, page.Number(), "bad page number  (i: %d, n: %d)", i, n)
		if i < expectedCount/expectedPageSize {
			assert.True(t, ok, "Next() MUST return true if remaining pages (i: %d, n: %d)", i, n)
		} else {
			assert.False(t, ok, "Next() MUST return false for last page (i: %d, n: %d)", i, n)
		}

		j := 0
		for k, item := range page.All() {
			assert.Equal(t, j, k, "bad item pos in page (page: %d)", page.Number())
			assert.Equal(t, n, item.Key(), "bad item key (page: %d)", page.Number())
			assert.Equal(t, expectedMessages[n], item.Val(), "bad item value (page: %d)", page.Number())
			n++
			j++
		}

		i++
		if !ok {
			_, next, err := p.Next()
			assert.False(t, next)
			assert.Error(t, err)
			assert.Equal(t, ErrNotExist, err)
			break
		}
	}
	assert.Len(t, expectedMessages, expectedCount, "bad produced msg count")
	assert.Equal(t, expectedCount/expectedPageSize+1, i, "bad page count")
	assert.Equal(t, expectedCount, k, "bad push call count")
	assert.Equal(t, expectedCount, n, "bad item iteration count")
}

func TestPaginer_Prev(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 2
	expectedCount := 10

	k := 0
	var expectedMessages []string
	p := NewPaginer(expectedPageSize, expectedPreloadCount, func(push func(Entry[int, string]) bool) {
		for {
			msg := fmt.Sprintf("msg%d", k)
			expectedMessages = append(expectedMessages, msg)
			// fmt.Printf("pushing msg: [%s] ...\n", msg)
			now := time.Now()
			e := NewEntry(k, msg, k, now, nil, nil, nil)
			if !push(e) {
				// fmt.Printf("breaked!\n")
				break
			}
			k++
			if k >= expectedCount {
				// End source
				// fmt.Printf("source end reached\n")
				break
			}
		}
	})
	require.NotNil(t, p)

	assert.Panics(t, func() {
		p.Prev()
	})

	page, ok, err := p.Next()
	assert.NoError(t, err)
	require.NotNil(t, page)
	assert.True(t, ok)
	assert.Equal(t, expectedPageSize, page.Len())
	assert.Equal(t, 0, page.Number())

	assert.Panics(t, func() {
		p.Prev()
	})

	page, ok, err = p.Next()
	assert.NoError(t, err)
	require.NotNil(t, page)
	assert.True(t, ok)
	assert.Equal(t, expectedPageSize, page.Len())
	assert.Equal(t, 1, page.Number())

	page, ok, err = p.Prev()
	assert.NoError(t, err)
	require.NotNil(t, page)
	assert.False(t, ok)
	assert.Equal(t, expectedPageSize, page.Len())
	assert.Equal(t, 0, page.Number())

	assert.Panics(t, func() {
		p.Prev()
	})
}

func TestPaginer_All(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 2
	expectedCount := 10

	a := 5
	b := &a
	assert.Equal(t, a, *b, "int ptr is a bad idea")

	k := 0
	var expectedMessages []string
	p := NewPaginer(expectedPageSize, expectedPreloadCount, func(push func(Entry[int, string]) bool) {
		for {
			msg := fmt.Sprintf("msg%d", k)
			expectedMessages = append(expectedMessages, msg)
			now := time.Now()
			e := NewEntry(k, msg, k, now, []byte("foobarba"), nil, nil)
			// fmt.Printf("pushing entry: [%s] ...\n", e)
			if !push(e) {
				// fmt.Printf("breaked!\n")
				break
			}
			k++
			if k >= expectedCount {
				// End source
				// fmt.Printf("source end reached\n")
				break
			}
		}
	})
	require.NotNil(t, p)
	// assert.Equal(t, expectedPageSize*expectedPreloadCount, k)

	n := 0
	for entry := range p.All() {
		assert.NoError(t, entry.Error())
		assert.Equal(t, expectedMessages[n], entry.Val(), "bad entry value")
		n++
	}
	assert.Len(t, expectedMessages, expectedCount, "bad produced msg count")
	assert.Equal(t, expectedCount, n, "bad entry iteration count")

	// 2 consecutive iteration should works
	n = 0
	for entry := range p.All() {
		assert.NoError(t, entry.Error())
		assert.Equal(t, expectedMessages[n], entry.Val(), "bad entry value")
		n++
	}
	assert.Len(t, expectedMessages, expectedCount, "bad produced msg count")
	assert.Equal(t, expectedCount, n, "bad entry iteration count")
}

func TestPaginer_All_Empty(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 0
	expectedCount := 0

	p := NewPaginer(expectedPageSize, expectedPreloadCount, func(push func(Entry[int, string]) bool) {
		// push nothing
	})
	require.NotNil(t, p)

	n := 0
	for entry := range p.All() {
		assert.NotNil(t, entry)
		n++
	}
	assert.Equal(t, expectedCount, n, "bad entry iteration count")
}

func TestPaginer_All_OneItem(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 0
	expectedCount := 1

	k := 0
	var expectedMessages []string
	p := NewPaginer(expectedPageSize, expectedPreloadCount, func(push func(Entry[int, string]) bool) {
		msg := fmt.Sprintf("msg%d", k)
		expectedMessages = append(expectedMessages, msg)
		now := time.Now()
		e := NewEntry(k, msg, k, now, []byte("foobarba"), nil, nil)
		// fmt.Printf("pushing entry: [%s] ...\n", e)
		push(e)
	})
	require.NotNil(t, p)

	n := 0
	for entry := range p.All() {
		assert.NotNil(t, entry)
		n++
	}
	assert.Len(t, expectedMessages, expectedCount, "bad produced msg count")
	assert.Equal(t, expectedCount, n, "bad entry iteration count")
}

func TestPaginer_WithErrors(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 2
	expectedCountBeforeError := 7
	expectedError := fmt.Errorf("blocking error")
	k := 0
	var expectedMessages []string
	p := NewPaginer(expectedPageSize, expectedPreloadCount, func(push func(Entry[int, string]) bool) {
		for {
			var err error
			if k >= expectedCountBeforeError {
				err = expectedError
			}
			msg := fmt.Sprintf("msg%d", k)
			expectedMessages = append(expectedMessages, msg)
			// fmt.Printf("pushing msg: [%s] ...\n", msg)
			now := time.Now()
			e := NewEntry(k, msg, k, now, nil, err, nil)
			if !push(e) {
				// fmt.Printf("breaked!\n")
				break
			}
			k++
		}
	})
	require.NotNil(t, p)

	page, ok, err := p.Next()
	assert.NotNil(t, page)
	assert.True(t, ok)
	assert.NoError(t, err)

	page, ok, err = p.Next()
	assert.NotNil(t, page)
	assert.True(t, ok)
	assert.NoError(t, err)

	page, ok, err = p.Next()
	assert.NotNil(t, page)
	assert.False(t, ok)
	require.Error(t, err)
	assert.ErrorIs(t, err, expectedError)

	assert.Equal(t, expectedCountBeforeError, k)
}

func TestPaginer_Preloading(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 2
	expectedCount := 14

	k := 0
	var expectedMessages []string
	p := NewPaginer(expectedPageSize, expectedPreloadCount, func(push func(Entry[int, string]) bool) {
		for {
			msg := fmt.Sprintf("msg%d", k)
			expectedMessages = append(expectedMessages, msg)
			// fmt.Printf("pushing msg: [%s] ...\n", msg)
			now := time.Now()
			e := NewEntry(k, msg, k, now, nil, nil, nil)
			if !push(e) {
				// fmt.Printf("breaked!\n")
				break
			}
			k++
			if k >= expectedCount {
				// End source
				// fmt.Printf("source end reached\n")
				break
			}
		}
	})
	require.NotNil(t, p)

	time.Sleep(10 * time.Millisecond)
	// Preloading should preload 3 pages (frst page + 2 in advance)
	assert.Equal(t, (expectedPreloadCount+1)*expectedPageSize, k)

	p.Next()
	time.Sleep(10 * time.Millisecond)
	// Preloading should preload a 4° page
	assert.Equal(t, (expectedPreloadCount+2)*expectedPageSize, k)

	p.Next()
	time.Sleep(10 * time.Millisecond)
	// Preloading should preload last page
	assert.Equal(t, expectedCount, k)

	p.Next()
	time.Sleep(10 * time.Millisecond)
	// Preloading should be ended
	assert.Equal(t, expectedCount, k)
}
