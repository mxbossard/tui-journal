package immutxtdb

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginer_Empty(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 2
	p := newPaginer(expectedPageSize, expectedPreloadCount, func(push func(int, string, error) bool) {
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

func TestPaginer_AllIterator(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 2
	expectedCount := 10

	k := 0
	var expectedMessages []string
	p := newPaginer(expectedPageSize, expectedPreloadCount, func(push func(int, string, error) bool) {
		for {
			msg := fmt.Sprintf("msg%d", k)
			expectedMessages = append(expectedMessages, msg)
			// fmt.Printf("pushing msg: [%s] ...\n", msg)
			if !push(k, msg, nil) {
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

	i := 0
	for err, page := range p.All() {
		assert.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, i, page.Number(), "bad page number")
		i++
	}
	assert.Equal(t, expectedCount/expectedPageSize+1, i, "bad page count")
}

func TestPaginer_NextPages(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 2
	expectedCount := 10

	k := 0
	var expectedMessages []string
	p := newPaginer(expectedPageSize, expectedPreloadCount, func(push func(int, string, error) bool) {
		for {
			msg := fmt.Sprintf("msg%d", k)
			expectedMessages = append(expectedMessages, msg)
			// fmt.Printf("pushing msg: [%s] ...\n", msg)
			if !push(k, msg, nil) {
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

		for i, item := range page.All() {
			assert.Equal(t, n, item.Key(), "bad item key (page: %d)", page.Number())
			assert.Equal(t, expectedMessages[n], item.Val(), "bad item value (page: %d)", page.Number())
			assert.Equal(t, n%expectedPageSize, i, "bad item order in page (page: %d)", page.Number())
			n++
		}

		i++
		if !ok {
			break
		}
	}
	assert.Len(t, expectedMessages, expectedCount, "bad produced msg count")
	assert.Equal(t, expectedCount/expectedPageSize+1, i, "bad page count")
	assert.Equal(t, expectedCount, k, "bad push call count")
	assert.Equal(t, expectedCount, n, "bad item iteration count")
}

func TestPaginer_PrevPages(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 2
	expectedCount := 10

	k := 0
	var expectedMessages []string
	p := newPaginer(expectedPageSize, expectedPreloadCount, func(push func(int, string, error) bool) {
		for {
			msg := fmt.Sprintf("msg%d", k)
			expectedMessages = append(expectedMessages, msg)
			// fmt.Printf("pushing msg: [%s] ...\n", msg)
			if !push(k, msg, nil) {
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

func TestPaginer_WithErrors(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 2
	expectedCountBeforeError := 7
	expectedError := fmt.Errorf("blocking error")
	k := 0
	var expectedMessages []string
	p := newPaginer(expectedPageSize, expectedPreloadCount, func(push func(int, string, error) bool) {
		for {
			var err error
			if k >= expectedCountBeforeError {
				err = expectedError
			}
			msg := fmt.Sprintf("msg%d", k)
			expectedMessages = append(expectedMessages, msg)
			// fmt.Printf("pushing msg: [%s] ...\n", msg)
			if !push(k, msg, err) {
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
	p := newPaginer(expectedPageSize, expectedPreloadCount, func(push func(int, string, error) bool) {
		for {
			msg := fmt.Sprintf("msg%d", k)
			expectedMessages = append(expectedMessages, msg)
			// fmt.Printf("pushing msg: [%s] ...\n", msg)
			if !push(k, msg, nil) {
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
