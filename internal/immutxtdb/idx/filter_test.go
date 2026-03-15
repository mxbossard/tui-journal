package idx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const YYYYMMDD = "2006-01-02"

func TestAfterFilter(t *testing.T) {
	time1, err := time.Parse(YYYYMMDD, "2026-03-15")
	require.NoError(t, err)
	time2, err := time.Parse(YYYYMMDD, "2026-03-16")
	require.NoError(t, err)
	time3, err := time.Parse(YYYYMMDD, "2026-03-17")
	require.NoError(t, err)

	f := AfterFilter(time2)
	ok1, loop1 := f.TimeFilter()(time1)
	assert.False(t, ok1)
	assert.True(t, loop1)

	ok3, loop3 := f.TimeFilter()(time3)
	assert.True(t, ok3)
	assert.True(t, loop3)
}

func TestBeforeFilter(t *testing.T) {
	time1, err := time.Parse(YYYYMMDD, "2026-03-15")
	require.NoError(t, err)
	time2, err := time.Parse(YYYYMMDD, "2026-03-16")
	require.NoError(t, err)
	time3, err := time.Parse(YYYYMMDD, "2026-03-17")
	require.NoError(t, err)

	f := BeforeFilter(time2)
	ok1, loop1 := f.TimeFilter()(time1)
	assert.True(t, ok1)
	assert.True(t, loop1)

	ok3, loop3 := f.TimeFilter()(time3)
	assert.False(t, ok3)
	assert.True(t, loop3)
}

func TestBetweenFilter(t *testing.T) {
	time1, err := time.Parse(YYYYMMDD, "2026-03-15")
	require.NoError(t, err)
	time2, err := time.Parse(YYYYMMDD, "2026-03-16")
	require.NoError(t, err)
	time3, err := time.Parse(YYYYMMDD, "2026-03-17")
	require.NoError(t, err)
	time4, err := time.Parse(YYYYMMDD, "2026-03-18")
	require.NoError(t, err)
	time5, err := time.Parse(YYYYMMDD, "2026-03-19")
	require.NoError(t, err)

	f := BetweenFilter(time2, time4)
	ok1, loop1 := f.TimeFilter()(time1)
	assert.False(t, ok1)
	assert.True(t, loop1)

	ok3, loop3 := f.TimeFilter()(time3)
	assert.True(t, ok3)
	assert.True(t, loop3)

	ok5, loop5 := f.TimeFilter()(time5)
	assert.False(t, ok5)
	assert.True(t, loop5)
}
