package immutxtdb

import (
	"fmt"
	"testing"

	"github.com/smartystreets/assertions/assert"
	"github.com/stretchr/testify/assert"
)

func TestPaginer(t *testing.T) {
	expectedPageSize := 3
	expectedPreloadCount := 2
	errChan := make(chan error)

	k := 0
	p := newPaginer(expectedPageSize, expectedPreloadCount, errChan, func(push func(int, string) bool) {
		for {
			if !push(k, fmt.Sprintf("msg%d", k)) {
				break
			}
		}
	})
	assert.NotNil(t, p)

	assert.Equal(t, expectedPageSize*expectedPreloadCount, k)
}
