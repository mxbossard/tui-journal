package idx

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
	"github.com/mxbossard/utilz/filez"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicIndex_Add(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBasicIndex_Add")
	defer os.RemoveAll(tmpDir)

	expectedPageSize := 10
	expectedState := dummyState
	expectedTime := time.Now()
	expectedKeySize := 8
	keySer := serialize.AsciiSerializer{}
	valSer := serialize.AsciiSerializer{}
	enc := NewAsciiEncoder(0, len(expectedState), expectedKeySize, 100)
	bIdx, err := NewBasicIndex(tmpDir, "foo", "bar", keySer, valSer, nil, nil, enc, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)

	assert.Implements(t, (*Index[string, string])(nil), bIdx)

	_, err = bIdx.Add(expectedState, expectedTime, "k1", "foo")
	assert.NoError(t, err)
}

func TestBasicIndex_Count(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBasicIndex_Count")
	defer os.RemoveAll(tmpDir)

	expectedPageSize := 10
	expectedState := dummyState
	expectedTime := time.Now()
	expectedKeySize := 8
	keySer := serialize.AsciiSerializer{}
	valSer := serialize.AsciiSerializer{}
	enc := NewAsciiEncoder(0, len(expectedState), expectedKeySize, 100)
	bIdx, err := NewBasicIndex(tmpDir, "foo", "bar", keySer, valSer, nil, nil, enc, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)

	count, err := bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	_, err = bIdx.Add(expectedState, expectedTime, "k1", "foo")
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	_, err = bIdx.Add(expectedState, expectedTime, "k2", "bar")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState, expectedTime, "k3", "baz")
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	_, err = bIdx.Add(expectedState, expectedTime, "k1", "pif")
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 4, count)
}

func TestBasicIndex_CountReopen(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBasicIndex_CountReopen")
	defer os.RemoveAll(tmpDir)

	expectedPageSize := 10
	expectedState := dummyState
	expectedTime := time.Now()
	expectedKeySize := 8
	keySer := serialize.AsciiSerializer{}
	valSer := serialize.AsciiSerializer{}
	enc := NewAsciiEncoder(0, len(expectedState), expectedKeySize, 100)
	bIdx, err := NewBasicIndex(tmpDir, "foo", "bar", keySer, valSer, nil, nil, enc, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)

	count, err := bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	_, err = bIdx.Add(expectedState, expectedTime, "k1", "foo")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState, expectedTime, "k2", "bar")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState, expectedTime, "k3", "baz")
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	enc2 := NewAsciiEncoder(0, len(expectedState), expectedKeySize, 100)
	bIdx2, err := NewBasicIndex(tmpDir, "foo", "bar", keySer, valSer, nil, nil, enc2, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)

	count, err = bIdx2.Count()
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestBasicIndex_PaginateAll(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBasicIndex_PaginateAll")
	defer os.RemoveAll(tmpDir)

	expectedPageSize := 10
	expectedState := dummyState
	expectedTime := time.Now()
	expectedKeySize := 16
	keySer := serialize.AsciiSerializer{}
	valSer := serialize.AsciiSerializer{}
	enc := NewAsciiEncoder(0, len(expectedState), expectedKeySize, 100)
	bIdx, err := NewBasicIndex(tmpDir, "foo", "bar", keySer, valSer, nil, nil, enc, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)
	_, err = bIdx.Add(expectedState, expectedTime, "k1", "foo")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState, expectedTime, "k2", "bar")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState, expectedTime, "k3", "baz")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState, expectedTime, "k1", "pif")
	assert.NoError(t, err)

	p, errChan := bIdx.PaginateAll(TopToBottom)
	require.NotNil(t, p)
	require.NotNil(t, errChan)

	page, ok, err := p.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, 4, page.Len())
	require.True(t, page.Len() >= 4)

	entries := page.Entries()
	assert.Equal(t, "k1", entries[0].Key())
	assert.Equal(t, "foo", entries[0].Val())
	assert.Equal(t, "k2", entries[1].Key())
	assert.Equal(t, "bar", entries[1].Val())
	assert.Equal(t, "k3", entries[2].Key())
	assert.Equal(t, "baz", entries[2].Val())
	assert.Equal(t, "k1", entries[3].Key())
	assert.Equal(t, "pif", entries[3].Val())

	p2, errChan := bIdx.PaginateAll(BottomToTop)
	require.NotNil(t, p2)
	require.NotNil(t, errChan)

	page2, ok, err := p2.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page2)
	assert.Equal(t, 4, page2.Len())
	require.True(t, page.Len() >= 4)

	entries2 := page2.Entries()
	assert.Equal(t, "k1", entries2[0].Key())
	assert.Equal(t, "pif", entries2[0].Val())
	assert.Equal(t, "k3", entries2[1].Key())
	assert.Equal(t, "baz", entries2[1].Val())
	assert.Equal(t, "k2", entries2[2].Key())
	assert.Equal(t, "bar", entries2[2].Val())
	assert.Equal(t, "k1", entries2[3].Key())
	assert.Equal(t, "foo", entries2[3].Val())
}

func TestBasicIndex_PaginateAllReopen(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBasicIndex_PaginateAllReopen")
	defer os.RemoveAll(tmpDir)

	expectedPageSize := 10
	expectedState := dummyState
	expectedTime := time.Now()
	expectedKeySize := 16
	expectedQualifier := "foo"
	expectedDevice := "bar"
	keySer := serialize.AsciiSerializer{}
	valSer := serialize.AsciiSerializer{}

	// Open a first Idx
	enc := NewAsciiEncoder(0, len(expectedState), expectedKeySize, 100)
	bIdx, err := NewBasicIndex(tmpDir, expectedQualifier, expectedDevice, keySer, valSer, nil, nil, enc, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)
	_, err = bIdx.Add(expectedState, expectedTime, "k1", "foo")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState, expectedTime, "k2", "bar")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState, expectedTime, "k3", "baz")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState, expectedTime, "k1", "pif")
	assert.NoError(t, err)

	// Open a second Idx
	enc2 := NewAsciiEncoder(0, len(expectedState), expectedKeySize, 100)
	bIdx2, err := NewBasicIndex(tmpDir, expectedQualifier, expectedDevice, keySer, valSer, nil, nil, enc2, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)

	p, errChan := bIdx2.PaginateAll(TopToBottom)
	require.NotNil(t, p)
	require.NotNil(t, errChan)

	page, ok, err := p.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, 4, page.Len())
	require.True(t, page.Len() >= 4)

	entries := page.Entries()
	assert.Equal(t, "k1", entries[0].Key())
	assert.Equal(t, "foo", entries[0].Val())
	assert.Equal(t, "k2", entries[1].Key())
	assert.Equal(t, "bar", entries[1].Val())
	assert.Equal(t, "k3", entries[2].Key())
	assert.Equal(t, "baz", entries[2].Val())
	assert.Equal(t, "k1", entries[3].Key())
	assert.Equal(t, "pif", entries[3].Val())
}

func TestBasicIndex_Paginate(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBasicIndex_Paginate")
	defer os.RemoveAll(tmpDir)

	expectedPageSize := 10
	expectedState := dummyState
	expectedTime := time.Now()
	expectedKeySize := 16
	keySer := serialize.AsciiSerializer{}
	valSer := serialize.AsciiSerializer{}
	enc := NewAsciiEncoder(0, len(expectedState), expectedKeySize, 100)
	bIdx, err := NewBasicIndex(tmpDir, "foo", "bar", keySer, valSer, nil, nil, enc, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)
	_, err = bIdx.Add(expectedState, expectedTime, "k1", "foo")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState, expectedTime, "k2", "bar")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState, expectedTime, "k3", "baz")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState, expectedTime, "k1", "pif")
	assert.NoError(t, err)

	pk1, errChan := bIdx.Paginate("k1", TopToBottom)
	require.NotNil(t, pk1)
	require.NotNil(t, errChan)

	page, ok, err := pk1.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, 2, page.Len())
	require.True(t, page.Len() >= 2)

	entries := page.Entries()
	assert.Equal(t, "k1", entries[0].Key())
	assert.Equal(t, "foo", entries[0].Val())
	assert.Equal(t, "k1", entries[1].Key())
	assert.Equal(t, "pif", entries[1].Val())

	pk2, errChan := bIdx.Paginate("k1", BottomToTop)
	require.NotNil(t, pk2)
	require.NotNil(t, errChan)

	page2, ok, err := pk2.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page2)
	assert.Equal(t, 2, page2.Len())
	require.True(t, page.Len() >= 2)

	entries2 := page2.Entries()
	assert.Equal(t, "k1", entries2[0].Key())
	assert.Equal(t, "pif", entries2[0].Val())
	assert.Equal(t, "k1", entries2[1].Key())
	assert.Equal(t, "foo", entries2[1].Val())

	pk3, errChan := bIdx.Paginate("k3", BottomToTop)
	require.NotNil(t, pk3)
	require.NotNil(t, errChan)

	page3, ok, err := pk3.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page3)
	assert.Equal(t, 1, page3.Len())
	require.True(t, page.Len() >= 1)

	entries3 := page3.Entries()
	assert.Equal(t, "k3", entries3[0].Key())
	assert.Equal(t, "baz", entries3[0].Val())
}

func TestBasicIndex_Filter(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBasicIndex_Filter")
	defer os.RemoveAll(tmpDir)

	expectedPageSize := 10
	expectedStateLen := 8
	expectedState1 := BuildState(expectedStateLen, "state1")
	expectedState2 := BuildState(expectedStateLen, "state2")
	expectedState3 := BuildState(expectedStateLen, "state3")
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
	time6, err := time.Parse(YYYYMMDD, "2026-03-20")
	require.NoError(t, err)
	_ = time1
	_ = time6

	expectedKeySize := 16
	keySer := serialize.AsciiSerializer{}
	valSer := serialize.AsciiSerializer{}
	enc := NewAsciiEncoder(0, expectedStateLen, expectedKeySize, 100)
	bIdx, err := NewBasicIndex(tmpDir, "foo", "bar", keySer, valSer, nil, nil, enc, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)
	_, err = bIdx.Add(expectedState1, time2, "k1", "foo")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState1, time3, "k2", "bar")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState2, time4, "k3", "baz")
	assert.NoError(t, err)
	_, err = bIdx.Add(expectedState3, time5, "k1", "pif")
	assert.NoError(t, err)

	p1, errChan := bIdx.FilterAll(TopToBottom, BeforeFilter(time4))
	require.NotNil(t, p1)
	require.NotNil(t, errChan)

	page, ok, err := p1.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, 2, page.Len())
	require.True(t, page.Len() >= 2)

	entries := page.Entries()
	assert.Equal(t, "k1", entries[0].Key())
	assert.Equal(t, "foo", entries[0].Val())
	assert.Equal(t, "k2", entries[1].Key())
	assert.Equal(t, "bar", entries[1].Val())

	p2, errChan := bIdx.FilterAll(TopToBottom, AfterFilter(time4))
	require.NotNil(t, p2)
	require.NotNil(t, errChan)

	page, ok, err = p2.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, 1, page.Len())
	require.True(t, page.Len() >= 1)

	entries = page.Entries()
	assert.Equal(t, "k1", entries[0].Key())
	assert.Equal(t, "pif", entries[0].Val())

	p3, errChan := bIdx.FilterAll(TopToBottom, BetweenFilter(time2, time4))
	require.NotNil(t, p3)
	require.NotNil(t, errChan)

	page, ok, err = p3.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, 1, page.Len())
	require.True(t, page.Len() >= 1)

	entries = page.Entries()
	assert.Equal(t, "k2", entries[0].Key())
	assert.Equal(t, "bar", entries[0].Val())

	p4, errChan := bIdx.Filter("k1", TopToBottom, AfterFilter(time2))
	require.NotNil(t, p4)
	require.NotNil(t, errChan)

	page, ok, err = p4.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, 1, page.Len())
	require.True(t, page.Len() >= 1)

	entries = page.Entries()
	assert.Equal(t, "k1", entries[0].Key())
	assert.Equal(t, "pif", entries[0].Val())

	p5, errChan := bIdx.FilterAll(TopToBottom, StateFilter(func(s State) (ok bool, loop bool) {
		loop = true
		ok = bytes.Equal(s, expectedState1)
		return
	}))
	require.NotNil(t, p5)
	require.NotNil(t, errChan)

	page, ok, err = p5.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, 2, page.Len())
	require.True(t, page.Len() >= 2)

	entries = page.Entries()
	assert.Equal(t, "k1", entries[0].Key())
	assert.Equal(t, "foo", entries[0].Val())
	assert.Equal(t, "k2", entries[1].Key())
	assert.Equal(t, "bar", entries[1].Val())

	p6, errChan := bIdx.FilterAll(TopToBottom, StateFilter(func(s State) (ok bool, loop bool) {
		loop = true
		ok = bytes.Equal(s, expectedState2)
		return
	}))
	require.NotNil(t, p6)
	require.NotNil(t, errChan)

	page, ok, err = p6.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, 1, page.Len())
	require.True(t, page.Len() >= 1)

	entries = page.Entries()
	assert.Equal(t, "k3", entries[0].Key())
	assert.Equal(t, "baz", entries[0].Val())
}
