package idx

import (
	"os"
	"testing"

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
	expectedKeySize := 8
	ser := serialize.AsciiSerializer{}
	enc := NewAsciiEncoder(0, len(expectedState), expectedKeySize, 100)
	bIdx, err := NewBasicIndex(tmpDir, "foo", "bar", ser, enc, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)

	err = bIdx.Add(expectedState, "k1", "foo")
	assert.NoError(t, err)
}

func TestBasicIndex_Count(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBasicIndex_Count")
	defer os.RemoveAll(tmpDir)

	expectedPageSize := 10
	expectedState := dummyState
	expectedKeySize := 8
	ser := serialize.AsciiSerializer{}
	enc := NewAsciiEncoder(0, len(expectedState), expectedKeySize, 100)
	bIdx, err := NewBasicIndex(tmpDir, "foo", "bar", ser, enc, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)

	count, err := bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	err = bIdx.Add(expectedState, "k1", "foo")
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	err = bIdx.Add(expectedState, "k2", "bar")
	assert.NoError(t, err)
	err = bIdx.Add(expectedState, "k3", "baz")
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	err = bIdx.Add(expectedState, "k1", "pif")
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 4, count)
}

func TestBasicIndex_PaginateAll(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBasicIndex_PaginateAll")
	defer os.RemoveAll(tmpDir)

	expectedPageSize := 10
	expectedState := dummyState
	expectedKeySize := 16
	ser := serialize.AsciiSerializer{}
	enc := NewAsciiEncoder(0, len(expectedState), expectedKeySize, 100)
	bIdx, err := NewBasicIndex(tmpDir, "foo", "bar", ser, enc, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)
	err = bIdx.Add(expectedState, "k1", "foo")
	assert.NoError(t, err)
	err = bIdx.Add(expectedState, "k2", "bar")
	assert.NoError(t, err)
	err = bIdx.Add(expectedState, "k3", "baz")
	assert.NoError(t, err)
	err = bIdx.Add(expectedState, "k1", "pif")
	assert.NoError(t, err)

	p, errChan := bIdx.PaginateAll(TopToBottom, 100)
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

	p2, errChan := bIdx.PaginateAll(BottomToTop, 100)
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

func TestBasicIndex_Paginate(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBasicIndex_PaginateAll")
	defer os.RemoveAll(tmpDir)

	expectedPageSize := 10
	expectedState := dummyState
	expectedKeySize := 16
	ser := serialize.AsciiSerializer{}
	enc := NewAsciiEncoder(0, len(expectedState), expectedKeySize, 100)
	bIdx, err := NewBasicIndex(tmpDir, "foo", "bar", ser, enc, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)
	err = bIdx.Add(expectedState, "k1", "foo")
	assert.NoError(t, err)
	err = bIdx.Add(expectedState, "k2", "bar")
	assert.NoError(t, err)
	err = bIdx.Add(expectedState, "k3", "baz")
	assert.NoError(t, err)
	err = bIdx.Add(expectedState, "k1", "pif")
	assert.NoError(t, err)

	pk1, errChan := bIdx.Paginate("k1", TopToBottom, 100)
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

	pk2, errChan := bIdx.Paginate("k1", BottomToTop, 100)
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

	pk3, errChan := bIdx.Paginate("k3", BottomToTop, 100)
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
