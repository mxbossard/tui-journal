package bucket

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/mxbossard/utilz/timez"
	"github.com/stretchr/testify/require"
)

func TestCriteria_CreationTimeCriteria(t *testing.T) {
	time1 := timez.ParseYyyyMmDd("2026-04-01")
	time2 := timez.ParseYyyyMmDd("2026-04-02")
	time3 := timez.ParseYyyyMmDd("2026-04-03")
	time4 := timez.ParseYyyyMmDd("2026-04-04")
	time5 := timez.ParseYyyyMmDd("2026-04-05")

	c1 := basicCriterion{
		creationTimeAfter: &time3,
	}
	m1 := c1.CreationTimeMatcher()
	require.NotNil(t, m1)
	assert.True(t, m1(time4))
	assert.False(t, m1(time2))

	c2 := basicCriterion{
		creationTimeBefore: &time3,
	}
	m2 := c2.CreationTimeMatcher()
	require.NotNil(t, m2)
	assert.False(t, m2(time4))
	assert.True(t, m2(time2))

	c3 := basicCriterion{
		creationTimeAfter:  &time2,
		creationTimeBefore: &time4,
	}
	m3 := c3.CreationTimeMatcher()
	require.NotNil(t, m3)
	assert.False(t, m3(time1))
	assert.False(t, m3(time2))
	assert.True(t, m3(time3))
	assert.False(t, m3(time4))
	assert.False(t, m3(time5))
}

func TestCriteria_CreationTimeCriteria_Union(t *testing.T) {
	time1 := timez.ParseYyyyMmDd("2026-04-01")
	time2 := timez.ParseYyyyMmDd("2026-04-02")
	time3 := timez.ParseYyyyMmDd("2026-04-03")
	time4 := timez.ParseYyyyMmDd("2026-04-04")
	time5 := timez.ParseYyyyMmDd("2026-04-05")

	c1 := basicCriterion{
		creationTimeAfter: &time2,
	}
	m1 := c1.CreationTimeMatcher()
	require.NotNil(t, m1)

	c2 := basicCriterion{
		creationTimeBefore: &time4,
	}
	m2 := c2.CreationTimeMatcher()
	require.NotNil(t, m2)

	uOr := unionCriteria{
		logicalOr: true,
		criteria:  []basicCriterion{c1, c2},
	}
	mOr := uOr.CreationTimeMatcher()
	assert.True(t, mOr(time1))
	assert.True(t, mOr(time2))
	assert.True(t, mOr(time3))
	assert.True(t, mOr(time4))
	assert.True(t, mOr(time5))

	uAnd := unionCriteria{
		logicalOr: false,
		criteria:  []basicCriterion{c1, c2},
	}
	mAnd := uAnd.CreationTimeMatcher()
	assert.False(t, mAnd(time1))
	assert.False(t, mAnd(time2))
	assert.True(t, mAnd(time3))
	assert.False(t, mAnd(time4))
	assert.False(t, mAnd(time5))
}
