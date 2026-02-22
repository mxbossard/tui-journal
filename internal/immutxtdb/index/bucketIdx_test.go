package index

import (
	"os"
	"testing"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/utilz/filez"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBucketIndex_Add(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketIndex_Add")
	defer os.RemoveAll(tmpDir)

	bIdx, err := NewBucketIndex(tmpDir, "test")
	assert.NoError(t, err)
	require.NotNil(t, bIdx)

	err = bIdx.Add(Document, nil, "foo")
	assert.NoError(t, err)
}

func TestBucketIndex_Count(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketIndex_Count")
	defer os.RemoveAll(tmpDir)

	bIdx, err := NewBucketIndex(tmpDir, "test")
	assert.NoError(t, err)
	require.NotNil(t, bIdx)

	count, err := bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	err = bIdx.Add(Document, nil, "foo")
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	err = bIdx.Add(Document, nil, "bar")
	assert.NoError(t, err)
	err = bIdx.Add(Document, nil, "baz")
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	err = bIdx.Add(Document, nil, "foo")
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 4, count)
}

func TestBucketIndex_PaginateAll(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketIndex_PaginateAll")
	defer os.RemoveAll(tmpDir)

	bIdx, err := NewBucketIndex(tmpDir, "test")
	assert.NoError(t, err)
	require.NotNil(t, bIdx)
	err = bIdx.Add(Document, nil, "foo")
	assert.NoError(t, err)
	err = bIdx.Add(Document, nil, "bar")
	assert.NoError(t, err)
	err = bIdx.Add(Document, nil, "baz")
	assert.NoError(t, err)
	err = bIdx.Add(Document, nil, "foo")
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
	assert.Equal(t, "foo", entries[0].Val())
	assert.Equal(t, "bar", entries[1].Val())
	assert.Equal(t, "baz", entries[2].Val())
	assert.Equal(t, "foo", entries[3].Val())

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
	assert.Equal(t, "foo", entries2[0].Val())
	assert.Equal(t, "baz", entries2[1].Val())
	assert.Equal(t, "bar", entries2[2].Val())
	assert.Equal(t, "foo", entries2[3].Val())

}
