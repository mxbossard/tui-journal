package index

import (
	"os"
	"testing"

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

	err = bIdx.Add([]byte("foo"), model.NewLayerRef("file", 0, Dump))
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

	err = bIdx.Add([]byte("foo"), model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	err = bIdx.Add([]byte("bar"), model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)
	err = bIdx.Add([]byte("baz"), model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	err = bIdx.Add([]byte("foo"), model.NewLayerRef("file", 0, Dump))
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
	err = bIdx.Add([]byte("foo"), model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)
	err = bIdx.Add([]byte("bar"), model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)
	err = bIdx.Add([]byte("baz"), model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)
	err = bIdx.Add([]byte("foo"), model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)

	p, errChan := bIdx.PaginateAll(model.TopToBottom, 100)
	require.NotNil(t, p)
	require.NotNil(t, errChan)

	page, ok, err := p.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, 4, page.Len())
	require.True(t, page.Len() >= 4)

	entries := page.Entries()
	assert.Equal(t, "foo", entries[0].Key())
	assert.Equal(t, "bar", entries[1].Key())
	assert.Equal(t, "baz", entries[2].Key())
	assert.Equal(t, "foo", entries[3].Key())

	p2, errChan := bIdx.PaginateAll(model.BottomToTop, 100)
	require.NotNil(t, p2)
	require.NotNil(t, errChan)

	page2, ok, err := p2.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page2)
	assert.Equal(t, 4, page2.Len())
	require.True(t, page.Len() >= 4)

	entries2 := page2.Entries()
	assert.Equal(t, "foo", entries2[0].Key())
	assert.Equal(t, "baz", entries2[1].Key())
	assert.Equal(t, "bar", entries2[2].Key())
	assert.Equal(t, "foo", entries2[3].Key())

}
