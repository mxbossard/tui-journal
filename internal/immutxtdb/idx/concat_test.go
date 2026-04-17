package idx

import (
	"os"
	"testing"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/timez"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	expectedPageSize = 10
	partitionA       = "qA"
	partitionB       = "qB"
	partitionC       = "qC"
	partitionD       = "qD"
	expectedMsg1     = "msg1"
	expectedMsg2     = "msg2"
	expectedMsg3     = "msg3"
	expectedMsg4     = "msg4"
	expectedMsg5     = "msg5"
	expectedMsg6     = "msg6"
	expectedMsg7     = "msg4"
	expectedMsg8     = "msg5"
	expectedMsg9     = "msg6"
	expectedKeySize  = 16
)

var (
	expectedState = dummyState
	time1         = timez.ParseOrPanic(YYYYMMDD, "2026-03-15")
	time2         = timez.ParseOrPanic(YYYYMMDD, "2026-03-16")
	time3         = timez.ParseOrPanic(YYYYMMDD, "2026-03-17")
	time4         = timez.ParseOrPanic(YYYYMMDD, "2026-03-18")
	time5         = timez.ParseOrPanic(YYYYMMDD, "2026-03-19")
	time6         = timez.ParseOrPanic(YYYYMMDD, "2026-03-20")
	time7         = timez.ParseOrPanic(YYYYMMDD, "2026-03-21")
	time8         = timez.ParseOrPanic(YYYYMMDD, "2026-03-22")
	time9         = timez.ParseOrPanic(YYYYMMDD, "2026-03-23")
)

func testBuildCatIdx(t *testing.T) (string, *BasicIndexAggregate[string, string]) {
	tmpDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir)

	keySer := serialize.AsciiSerializer{}
	valSer := serialize.AsciiSerializer{}
	enc := NewAsciiEncoder(0, len(expectedState), expectedKeySize, 100)

	idxA, err := NewBasicIndex(tmpDir, "foo", partitionA, keySer, valSer, nil, nil, enc, expectedPageSize, 0)
	assert.NoError(t, err)
	require.NotNil(t, idxA)
	_, err = idxA.Add(expectedState, time1, "k1", expectedMsg1)
	assert.NoError(t, err)
	_, err = idxA.Add(expectedState, time5, "k2", expectedMsg5)
	assert.NoError(t, err)
	_, err = idxA.Add(expectedState, time9, "k3", expectedMsg9)
	assert.NoError(t, err)

	idxB, err := NewBasicIndex(tmpDir, "foo", partitionB, keySer, valSer, nil, nil, enc, expectedPageSize, 0)
	assert.NoError(t, err)
	require.NotNil(t, idxB)
	_, err = idxB.Add(expectedState, time4, "k3", expectedMsg4)
	assert.NoError(t, err)
	_, err = idxB.Add(expectedState, time2, "k2", expectedMsg2)
	assert.NoError(t, err)
	_, err = idxB.Add(expectedState, time6, "k1", expectedMsg6)
	assert.NoError(t, err)

	idxC, err := NewBasicIndex(tmpDir, "foo", partitionC, keySer, valSer, nil, nil, enc, expectedPageSize, 0)
	assert.NoError(t, err)
	require.NotNil(t, idxC)
	_, err = idxC.Add(expectedState, time7, "k1", expectedMsg7)
	assert.NoError(t, err)
	_, err = idxC.Add(expectedState, time8, "k2", expectedMsg8)
	assert.NoError(t, err)
	_, err = idxC.Add(expectedState, time3, "k3", expectedMsg3)
	assert.NoError(t, err)

	idxD, err := NewBasicIndex(tmpDir, "foo", partitionD, keySer, valSer, nil, nil, enc, expectedPageSize, 0)
	assert.NoError(t, err)
	require.NotNil(t, idxD)

	return tmpDir, Aggregate(1, 0,
		map[string]Index[string, string]{
			partitionA: idxA,
			partitionB: idxB,
			partitionC: idxC,
		},
		map[string]Index[string, string]{
			partitionD: idxD,
		},
	)
}

func TestConcat_Add(t *testing.T) {
	tmpDir, cat := testBuildCatIdx(t)
	defer os.RemoveAll(tmpDir)

	p, err := cat.Paginate("k10", BottomToTop)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	// Search for not added yet entry
	k := 0
	for range p.All() {
		k++
	}
	assert.Equal(t, 0, k)

	// Adding in ReadOnly idx should fail
	entry, err := cat.Add(partitionA, dummyState, time1, "k10", "v10")
	assert.Error(t, err)
	assert.Nil(t, entry)

	// Adding in ReadWrite idx should be ok
	entry, err = cat.Add(partitionD, dummyState, time1, "k10", "v10")
	assert.NoError(t, err)
	assert.NotNil(t, entry)

	// NOT IMPLEMENTED YET
	// Search for added entry not rebuilding paginer
	// k = 0
	// err = p.Reset()
	// assert.NoError(t, err)
	// for e := range p.All() {
	// 	k++
	// 	assert.Equal(t, "k10", e.Key())
	// 	assert.Equal(t, "v10", e.Val())
	// }
	// assert.Equal(t, 1, k)

	// Search for added entry with rebuilded paginer
	p, err = cat.Paginate("k10", BottomToTop)
	assert.NoError(t, err)
	assert.NotNil(t, p)
	k = 0
	p.Reset()
	for e := range p.All() {
		k++
		assert.Equal(t, "k10", e.Key())
		assert.Equal(t, "v10", e.Val())
	}
	assert.Equal(t, 1, k)
}

func TestConcat_Filter(t *testing.T) {
	tmpDir, cat := testBuildCatIdx(t)
	defer os.RemoveAll(tmpDir)

	p, err := cat.Filter("k1", TopToBottom, nil)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	var k int

	// First pagination use
	k = 0
	for e := range p.All() {
		k++
		assert.NotNil(t, e)
		switch k {
		case 1:
			assert.Equal(t, expectedMsg1, e.Val())
		case 2:
			assert.Equal(t, expectedMsg6, e.Val())
		case 3:
			assert.Equal(t, expectedMsg7, e.Val())
		}
	}
	assert.Equal(t, 3, k)

	// Second pagination use
	p.Reset()
	k = 0
	for e := range p.All() {
		k++
		assert.NotNil(t, e)
		switch k {
		case 1:
			assert.Equal(t, expectedMsg1, e.Val())
		case 2:
			assert.Equal(t, expectedMsg6, e.Val())
		case 3:
			assert.Equal(t, expectedMsg7, e.Val())
		}
	}
	assert.Equal(t, 3, k)

	// Change order
	r, err := cat.Filter("k1", BottomToTop, nil)
	assert.NoError(t, err)
	assert.NotNil(t, r)

	k = 0
	for e := range r.All() {
		k++
		assert.NotNil(t, e)
		switch k {
		case 1:
			assert.Equal(t, expectedMsg7, e.Val())
		case 2:
			assert.Equal(t, expectedMsg6, e.Val())
		case 3:
			assert.Equal(t, expectedMsg1, e.Val())
		}
	}
	assert.Equal(t, 3, k)

	// Change key
	p, err = cat.Filter("k3", TopToBottom, nil)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	k = 0
	for e := range p.All() {
		k++
		assert.NotNil(t, e)
		switch k {
		case 1:
			assert.Equal(t, expectedMsg3, e.Val())
		case 2:
			assert.Equal(t, expectedMsg4, e.Val())
		case 3:
			assert.Equal(t, expectedMsg9, e.Val())
		}
	}
	assert.Equal(t, 3, k)

	//Change key
	p, err = cat.Filter("k2", TopToBottom, nil)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	k = 0
	for e := range p.All() {
		k++
		assert.NotNil(t, e)
		switch k {
		case 1:
			assert.Equal(t, expectedMsg2, e.Val())
		case 2:
			assert.Equal(t, expectedMsg5, e.Val())
		case 3:
			assert.Equal(t, expectedMsg8, e.Val())
		}
	}
	assert.Equal(t, 3, k)
}

// func TestConcat_HashedFilter(t *testing.T) {
// 	tmpDir, cat := testBuildCatIdx(t)
// 	defer os.RemoveAll(tmpDir)

// 	p, err := cat.HashedFilter("k1", TopToBottom, nil)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, p)

// 	var k int

// 	// First pagination use
// 	k = 0
// 	for e := range p.All() {
// 		k++
// 		assert.NotNil(t, e)
// 		switch k {
// 		case 1:
// 			assert.Equal(t, expectedMsg1, e.Val())
// 		case 2:
// 			assert.Equal(t, expectedMsg6, e.Val())
// 		case 3:
// 			assert.Equal(t, expectedMsg7, e.Val())
// 		}
// 	}
// 	assert.Equal(t, 3, k)

// 	// Second pagination use
// 	p.Reset()
// 	k = 0
// 	for e := range p.All() {
// 		k++
// 		assert.NotNil(t, e)
// 		switch k {
// 		case 1:
// 			assert.Equal(t, expectedMsg1, e.Val())
// 		case 2:
// 			assert.Equal(t, expectedMsg6, e.Val())
// 		case 3:
// 			assert.Equal(t, expectedMsg7, e.Val())
// 		}
// 	}
// 	assert.Equal(t, 3, k)
// }

func TestConcat_FilterAll(t *testing.T) {
	tmpDir, cat := testBuildCatIdx(t)
	defer os.RemoveAll(tmpDir)

	p, err := cat.FilterAll(TopToBottom, nil)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	var k int

	// First pagination use
	k = 0
	for e := range p.All() {
		k++
		assert.NotNil(t, e)
		switch k {
		case 1:
			assert.Equal(t, expectedMsg1, e.Val())
		case 2:
			assert.Equal(t, expectedMsg4, e.Val())
		case 3:
			assert.Equal(t, expectedMsg2, e.Val())
		case 4:
			assert.Equal(t, expectedMsg5, e.Val())
		case 5:
			assert.Equal(t, expectedMsg6, e.Val())
		case 6:
			assert.Equal(t, expectedMsg7, e.Val())
		case 7:
			assert.Equal(t, expectedMsg8, e.Val())
		case 8:
			assert.Equal(t, expectedMsg3, e.Val())
		case 9:
			assert.Equal(t, expectedMsg9, e.Val())
		}
	}
	assert.Equal(t, 9, k)
}

func TestConcat_Paginate(t *testing.T) {
	tmpDir, cat := testBuildCatIdx(t)
	defer os.RemoveAll(tmpDir)

	p, err := cat.Paginate("k1", TopToBottom)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	var k int

	// First pagination use
	k = 0
	for e := range p.All() {
		k++
		assert.NotNil(t, e)
		switch k {
		case 1:
			assert.Equal(t, expectedMsg1, e.Val())
		case 2:
			assert.Equal(t, expectedMsg6, e.Val())
		case 3:
			assert.Equal(t, expectedMsg7, e.Val())
		}
	}
	assert.Equal(t, 3, k)
}

// func TestConcat_HashedPaginate(t *testing.T) {
// 	tmpDir, cat := testBuildCatIdx(t)
// 	defer os.RemoveAll(tmpDir)

// 	p, err := cat.HashedPaginate("k1", TopToBottom)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, p)

// 	var k int

// 	// First pagination use
// 	k = 0
// 	for e := range p.All() {
// 		k++
// 		assert.NotNil(t, e)
// 		switch k {
// 		case 1:
// 			assert.Equal(t, expectedMsg1, e.Val())
// 		case 2:
// 			assert.Equal(t, expectedMsg6, e.Val())
// 		case 3:
// 			assert.Equal(t, expectedMsg7, e.Val())
// 		}
// 	}
// 	assert.Equal(t, 3, k)
// }

func TestConcat_PaginateAll(t *testing.T) {
	tmpDir, cat := testBuildCatIdx(t)
	defer os.RemoveAll(tmpDir)

	p, err := cat.PaginateAll(TopToBottom)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	var k int

	// First pagination use
	k = 0
	for e := range p.All() {
		k++
		assert.NotNil(t, e)
		switch k {
		case 1:
			assert.Equal(t, expectedMsg1, e.Val())
		case 2:
			assert.Equal(t, expectedMsg4, e.Val())
		case 3:
			assert.Equal(t, expectedMsg2, e.Val())
		case 4:
			assert.Equal(t, expectedMsg5, e.Val())
		case 5:
			assert.Equal(t, expectedMsg6, e.Val())
		case 6:
			assert.Equal(t, expectedMsg7, e.Val())
		case 7:
			assert.Equal(t, expectedMsg8, e.Val())
		case 8:
			assert.Equal(t, expectedMsg3, e.Val())
		case 9:
			assert.Equal(t, expectedMsg9, e.Val())
		}
	}
	assert.Equal(t, 9, k)
}

func TestConcat_All(t *testing.T) {
	tmpDir, cat := testBuildCatIdx(t)
	defer os.RemoveAll(tmpDir)

	entries, err := cat.All(TopToBottom)
	assert.NoError(t, err)
	assert.NotNil(t, entries)

	var k int

	// First pagination use
	k = 0
	for e := range entries {
		k++
		assert.NotNil(t, e)
		switch k {
		case 1:
			assert.Equal(t, expectedMsg1, e.Val())
		case 2:
			assert.Equal(t, expectedMsg4, e.Val())
		case 3:
			assert.Equal(t, expectedMsg2, e.Val())
		case 4:
			assert.Equal(t, expectedMsg5, e.Val())
		case 5:
			assert.Equal(t, expectedMsg6, e.Val())
		case 6:
			assert.Equal(t, expectedMsg7, e.Val())
		case 7:
			assert.Equal(t, expectedMsg8, e.Val())
		case 8:
			assert.Equal(t, expectedMsg3, e.Val())
		case 9:
			assert.Equal(t, expectedMsg9, e.Val())
		}
	}
	assert.Equal(t, 9, k)
}

func TestSharedVar(t *testing.T) {
	a := 0
	m := make(map[string]int)
	p := &m
	fa := func() {
		a = 1
		assert.Equal(t, 1, a)
		m["foo"] = 1
		(*p)["foo"] = 3
	}
	assert.NotEqual(t, 1, a)

	fb := func() {
		a = 2
		assert.Equal(t, 2, a)
		assert.NotEqual(t, 1, m["foo"])
		assert.Equal(t, 3, (*p)["foo"])
		assert.Len(t, m, 1)
		assert.Len(t, *p, 1)
	}

	fa()
	fb()
	assert.Equal(t, 2, a)
	assert.NotEqual(t, 1, m["foo"])
	assert.Equal(t, 3, (*p)["foo"])
	assert.Len(t, m, 1)
	assert.Len(t, *p, 1)
}
