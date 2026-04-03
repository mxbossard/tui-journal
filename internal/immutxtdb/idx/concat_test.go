package idx

import (
	"os"
	"testing"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
	"github.com/mxbossard/utilz/filez"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcat_Add(t *testing.T) {
	// TODO
}

func TestConcat_Filter(t *testing.T) {
	// TODO
	tmpDir := filez.MkdirTempOrPanic("TestConcat_Filter")
	defer os.RemoveAll(tmpDir)

	expectedPageSize := 10
	expectedState := dummyState
	qualifierA := "qA"
	qualifierB := "qB"
	qualifierC := "qC"
	expectedMsg1 := "msg1"
	expectedMsg2 := "msg2"
	expectedMsg3 := "msg3"
	expectedMsg4 := "msg4"
	expectedMsg5 := "msg5"
	expectedMsg6 := "msg6"
	expectedMsg7 := "msg4"
	expectedMsg8 := "msg5"
	expectedMsg9 := "msg6"
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
	time7, err := time.Parse(YYYYMMDD, "2026-03-21")
	require.NoError(t, err)
	time8, err := time.Parse(YYYYMMDD, "2026-03-22")
	require.NoError(t, err)
	time9, err := time.Parse(YYYYMMDD, "2026-03-23")
	require.NoError(t, err)

	expectedKeySize := 16
	keySer := serialize.AsciiSerializer{}
	valSer := serialize.AsciiSerializer{}
	enc := NewAsciiEncoder(0, len(expectedState), expectedKeySize, 100)

	idxA, err := NewBasicIndex(tmpDir, "foo", qualifierA, keySer, valSer, nil, nil, enc, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, idxA)
	_, err = idxA.Add(expectedState, time1, "k1", expectedMsg1)
	assert.NoError(t, err)
	_, err = idxA.Add(expectedState, time5, "k2", expectedMsg5)
	assert.NoError(t, err)
	_, err = idxA.Add(expectedState, time9, "k3", expectedMsg9)
	assert.NoError(t, err)

	idxB, err := NewBasicIndex(tmpDir, "foo", qualifierB, keySer, valSer, nil, nil, enc, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, idxB)
	_, err = idxB.Add(expectedState, time4, "k3", expectedMsg4)
	assert.NoError(t, err)
	_, err = idxB.Add(expectedState, time2, "k2", expectedMsg2)
	assert.NoError(t, err)
	_, err = idxB.Add(expectedState, time6, "k1", expectedMsg6)
	assert.NoError(t, err)

	idxC, err := NewBasicIndex(tmpDir, "foo", qualifierC, keySer, valSer, nil, nil, enc, expectedPageSize)
	assert.NoError(t, err)
	require.NotNil(t, idxC)
	_, err = idxC.Add(expectedState, time7, "k1", expectedMsg7)
	assert.NoError(t, err)
	_, err = idxC.Add(expectedState, time8, "k2", expectedMsg8)
	assert.NoError(t, err)
	_, err = idxC.Add(expectedState, time3, "k3", expectedMsg3)
	assert.NoError(t, err)

	cat := BasicIndexCat[string, string]{
		pageSize:         1,
		preloadPageCount: 0,
		roIndexes: map[string]Index[string, string]{
			qualifierA: idxA,
			qualifierB: idxB,
			qualifierC: idxC,
		},
	}

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
