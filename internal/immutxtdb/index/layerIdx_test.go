package index

import (
	"os"
	"testing"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/mxbossard/utilz/filez"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLayerIndex_Add(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestLayerIndex_Add")
	defer os.RemoveAll(tmpDir)

	expectedTime := time.Now()
	expectedUid0 := StringToBucketUid("foo")

	bIdx, err := NewLayerIndex(tmpDir, "test", "salt")
	assert.NoError(t, err)
	require.NotNil(t, bIdx)

	_, err = bIdx.Add(nil, expectedTime, &expectedUid0, model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)
}

func TestLayerIndex_Count(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestLayerIndex_Count")
	defer os.RemoveAll(tmpDir)

	expectedTime := time.Now()
	expectedUid0 := StringToBucketUid("foo")
	expectedUid1 := StringToBucketUid("bar")
	expectedUid2 := StringToBucketUid("baz")
	expectedUid3 := StringToBucketUid("foo")

	bIdx, err := NewLayerIndex(tmpDir, "test", "salt")
	assert.NoError(t, err)
	require.NotNil(t, bIdx)

	count, err := bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 0, count)

	_, err = bIdx.Add(nil, expectedTime, &expectedUid0, model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	_, err = bIdx.Add(nil, expectedTime, &expectedUid1, model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)
	_, err = bIdx.Add(nil, expectedTime, &expectedUid2, model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 3, count)

	_, err = bIdx.Add(nil, expectedTime, &expectedUid3, model.NewLayerRef("file", 0, Dump))
	assert.NoError(t, err)

	count, err = bIdx.Count()
	assert.NoError(t, err)
	assert.Equal(t, 4, count)
}

func TestLayerIndex_PaginateAll(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestLayerIndex_PaginateAll")
	defer os.RemoveAll(tmpDir)

	expectedSalt := "salt"
	expectedTime := time.Now()
	expectedBucket0 := StringToBucketUid("foo")
	expectedBucket1 := StringToBucketUid("bar")
	expectedBucket2 := StringToBucketUid("baz")
	expectedBucket3 := StringToBucketUid("foo")

	bIdx, err := NewLayerIndex(tmpDir, "test", expectedSalt)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)
	_, err = bIdx.Add(nil, expectedTime, &expectedBucket0, model.NewLayerRef("file1", 10, Dump))
	assert.NoError(t, err)
	_, err = bIdx.Add(nil, expectedTime, &expectedBucket1, model.NewLayerRef("file2", 20, Dump))
	assert.NoError(t, err)
	_, err = bIdx.Add(nil, expectedTime, &expectedBucket2, model.NewLayerRef("file3", 30, Dump))
	assert.NoError(t, err)
	_, err = bIdx.Add(nil, expectedTime, &expectedBucket3, model.NewLayerRef("file4", 40, Dump))
	assert.NoError(t, err)

	p, errChan := bIdx.PaginateAll(idx.TopToBottom)
	require.NotNil(t, p)
	require.NotNil(t, errChan)

	page, ok, err := p.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, 4, page.Len())
	require.True(t, page.Len() >= 4)

	hasher := RotatingHasher([]byte(expectedSalt), LayerIdxKeySize)

	entries := page.Entries()
	// hashedKey0, err := hasher(0, idx.FixedSizeString(LayerIdxKeySize, "foo"))
	// hashedKey0, err := hasher(0, []byte("foo"))
	hashedKey0, err := hasher(0, expectedBucket0[:])
	assert.NoError(t, err)
	assert.Equal(t, hashedKey0, entries[0].Key()[:])
	assert.Equal(t, model.NewLayerRef("file1", 10, Dump), entries[0].Val())
	hashedKey1, err := hasher(1, expectedBucket1[:])
	assert.NoError(t, err)
	assert.Equal(t, hashedKey1, entries[1].Key()[:])
	assert.Equal(t, model.NewLayerRef("file2", 20, Dump), entries[1].Val())
	hashedKey2, err := hasher(2, expectedBucket2[:])
	assert.NoError(t, err)
	assert.Equal(t, hashedKey2, entries[2].Key()[:])
	assert.Equal(t, model.NewLayerRef("file3", 30, Dump), entries[2].Val())
	hashedKey3, err := hasher(3, expectedBucket3[:])
	assert.NoError(t, err)
	assert.Equal(t, hashedKey3, entries[3].Key()[:])
	assert.Equal(t, model.NewLayerRef("file4", 40, Dump), entries[3].Val())

	p2, errChan := bIdx.PaginateAll(idx.BottomToTop)
	require.NotNil(t, p2)
	require.NotNil(t, errChan)

	page2, ok, err := p2.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page2)
	assert.Equal(t, 4, page2.Len())
	require.True(t, page.Len() >= 4)

	entries2 := page2.Entries()
	hashedKey3, err = hasher(3, expectedBucket3[:])
	assert.NoError(t, err)
	assert.Equal(t, hashedKey3, entries2[0].Key()[:])
	assert.Equal(t, model.NewLayerRef("file4", 40, Dump), entries2[0].Val())
	hashedKey2, err = hasher(2, expectedBucket2[:])
	assert.NoError(t, err)
	assert.Equal(t, hashedKey2, entries2[1].Key()[:])
	assert.Equal(t, model.NewLayerRef("file3", 30, Dump), entries2[1].Val())
	hashedKey1, err = hasher(1, expectedBucket1[:])
	assert.NoError(t, err)
	assert.Equal(t, hashedKey1, entries2[2].Key()[:])
	assert.Equal(t, model.NewLayerRef("file2", 20, Dump), entries2[2].Val())
	hashedKey0, err = hasher(0, expectedBucket0[:])
	assert.NoError(t, err)
	assert.Equal(t, hashedKey0, entries2[3].Key()[:])
	assert.Equal(t, model.NewLayerRef("file1", 10, Dump), entries2[3].Val())

}

func TestLayerIndex_Paginate(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestLayerIndex_Paginate")
	defer os.RemoveAll(tmpDir)

	expectedSalt := "salt"
	expectedTime := time.Now()
	expectedBucket0 := StringToBucketUid("foo")
	expectedBucket1 := StringToBucketUid("bar")
	expectedBucket2 := StringToBucketUid("baz")
	expectedBucket3 := StringToBucketUid("foo")

	bIdx, err := NewLayerIndex(tmpDir, "test", expectedSalt)
	assert.NoError(t, err)
	require.NotNil(t, bIdx)
	_, err = bIdx.Add(nil, expectedTime, &expectedBucket0, model.NewLayerRef("file1", 10, Dump))
	assert.NoError(t, err)
	_, err = bIdx.Add(nil, expectedTime, &expectedBucket1, model.NewLayerRef("file2", 20, Dump))
	assert.NoError(t, err)
	_, err = bIdx.Add(nil, expectedTime, &expectedBucket2, model.NewLayerRef("file3", 30, Dump))
	assert.NoError(t, err)
	_, err = bIdx.Add(nil, expectedTime, &expectedBucket3, model.NewLayerRef("file4", 40, Dump))
	assert.NoError(t, err)

	p, errChan := bIdx.HashedPaginate(&expectedBucket0, idx.BottomToTop)
	require.NotNil(t, p)
	require.NotNil(t, errChan)

	page, ok, err := p.Next()
	assert.NoError(t, err)
	assert.False(t, ok)
	require.NotNil(t, page)
	assert.Equal(t, 2, page.Len())

	hasher := RotatingHasher([]byte(expectedSalt), LayerIdxKeySize)

	entries := page.Entries()
	require.True(t, page.Len() >= 1)
	hashedKey3, err := hasher(3, expectedBucket3[:])
	assert.NoError(t, err)
	assert.Equal(t, hashedKey3, entries[0].Key()[:])
	assert.Equal(t, model.NewLayerRef("file4", 40, Dump), entries[0].Val())
	require.True(t, page.Len() >= 2)
	hashedKey0, err := hasher(0, expectedBucket0[:])
	assert.NoError(t, err)
	assert.Equal(t, hashedKey0, entries[1].Key()[:])
	assert.Equal(t, model.NewLayerRef("file1", 10, Dump), entries[1].Val())
}
