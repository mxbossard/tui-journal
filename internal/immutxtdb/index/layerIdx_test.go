package index

import (
	"os"
	"testing"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/mxbossard/utilz/filez"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLayerIndex_Add(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestLayerIndex_Add")
	defer os.RemoveAll(tmpDir)

	bIdx, err := NewLayerIndex(tmpDir, "test")
	assert.NoError(t, err)
	require.NotNil(t, bIdx)

	err = bIdx.Add(nil, []byte("foo"), model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)
}

func TestLayerIndex_Count(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestLayerIndex_Count")
	defer os.RemoveAll(tmpDir)

	bIdx, err := NewLayerIndex(tmpDir, "test")
	assert.NoError(t, err)
	require.NotNil(t, bIdx)

	count, err := bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	err = bIdx.Add(nil, []byte("foo"), model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	err = bIdx.Add(nil, []byte("bar"), model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)
	err = bIdx.Add(nil, []byte("baz"), model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	err = bIdx.Add(nil, []byte("foo"), model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 4, count)
}

func TestLayerIndex_PaginateAll(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestLayerIndex_PaginateAll")
	defer os.RemoveAll(tmpDir)

	bIdx, err := NewLayerIndex(tmpDir, "test")
	assert.NoError(t, err)
	require.NotNil(t, bIdx)
	err = bIdx.Add(nil, []byte("foo"), model.NewLayerRef("file1", 10, Dump))
	assert.NoError(t, err)
	err = bIdx.Add(nil, []byte("bar"), model.NewLayerRef("file2", 20, Dump))
	assert.NoError(t, err)
	err = bIdx.Add(nil, []byte("baz"), model.NewLayerRef("file3", 30, Dump))
	assert.NoError(t, err)
	err = bIdx.Add(nil, []byte("foo"), model.NewLayerRef("file4", 40, Dump))
	assert.NoError(t, err)

	p, errChan := bIdx.PaginateAll(idx.TopToBottom, 100)
	require.NotNil(t, p)
	require.NotNil(t, errChan)

	page, ok, err := p.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, 4, page.Len())
	require.True(t, page.Len() >= 4)

	entries := page.Entries()
	assert.Equal(t, idx.FixedSizeStringKey(layerIdxKeySize, "foo"), entries[0].Key())
	assert.Equal(t, model.NewLayerRef("file1", 10, Dump), entries[0].Val())
	assert.Equal(t, idx.FixedSizeStringKey(layerIdxKeySize, "bar"), entries[1].Key())
	assert.Equal(t, model.NewLayerRef("file2", 20, Dump), entries[1].Val())
	assert.Equal(t, idx.FixedSizeStringKey(layerIdxKeySize, "baz"), entries[2].Key())
	assert.Equal(t, model.NewLayerRef("file3", 30, Dump), entries[2].Val())
	assert.Equal(t, idx.FixedSizeStringKey(layerIdxKeySize, "foo"), entries[3].Key())
	assert.Equal(t, model.NewLayerRef("file4", 40, Dump), entries[3].Val())

	p2, errChan := bIdx.PaginateAll(idx.BottomToTop, 100)
	require.NotNil(t, p2)
	require.NotNil(t, errChan)

	page2, ok, err := p2.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page2)
	assert.Equal(t, 4, page2.Len())
	require.True(t, page.Len() >= 4)

	entries2 := page2.Entries()
	assert.Equal(t, idx.FixedSizeStringKey(layerIdxKeySize, "foo"), entries2[0].Key())
	assert.Equal(t, model.NewLayerRef("file4", 40, Dump), entries2[0].Val())
	assert.Equal(t, idx.FixedSizeStringKey(layerIdxKeySize, "baz"), entries2[1].Key())
	assert.Equal(t, model.NewLayerRef("file3", 30, Dump), entries2[1].Val())
	assert.Equal(t, idx.FixedSizeStringKey(layerIdxKeySize, "bar"), entries2[2].Key())
	assert.Equal(t, model.NewLayerRef("file2", 20, Dump), entries2[2].Val())
	assert.Equal(t, idx.FixedSizeStringKey(layerIdxKeySize, "foo"), entries2[3].Key())
	assert.Equal(t, model.NewLayerRef("file1", 10, Dump), entries2[3].Val())

}

func TestLayerIndex_Paginate(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestLayerIndex_Paginate")
	defer os.RemoveAll(tmpDir)

	bIdx, err := NewLayerIndex(tmpDir, "test")
	assert.NoError(t, err)
	require.NotNil(t, bIdx)
	err = bIdx.Add(nil, []byte("foo"), model.NewLayerRef("file1", 10, Dump))
	assert.NoError(t, err)
	err = bIdx.Add(nil, []byte("bar"), model.NewLayerRef("file2", 20, Dump))
	assert.NoError(t, err)
	err = bIdx.Add(nil, []byte("baz"), model.NewLayerRef("file3", 30, Dump))
	assert.NoError(t, err)
	err = bIdx.Add(nil, []byte("foo"), model.NewLayerRef("file4", 40, Dump))
	assert.NoError(t, err)

	p, errChan := bIdx.Paginate([]byte("foo"), idx.BottomToTop, 100)
	require.NotNil(t, p)
	require.NotNil(t, errChan)

	page, ok, err := p.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, 2, page.Len())
	require.True(t, page.Len() >= 2)

	entries := page.Entries()
	assert.Equal(t, idx.FixedSizeStringKey(layerIdxKeySize, "foo"), entries[0].Key())
	assert.Equal(t, model.NewLayerRef("file4", 40, Dump), entries[0].Val())
	assert.Equal(t, idx.FixedSizeStringKey(layerIdxKeySize, "foo"), entries[1].Key())
	assert.Equal(t, model.NewLayerRef("file1", 10, Dump), entries[1].Val())
}
