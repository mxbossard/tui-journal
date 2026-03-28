package bucket

import (
	"os"
	"testing"
	"time"

	"github.com/mxbossard/utilz/collectionz"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/ztring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewBucketService(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("Test_NewBucketService")
	defer os.RemoveAll(tmpDir)

	expectedDevice := "device"
	expectedSalt := "salt"
	svc, err := NewBucketService(tmpDir, expectedDevice, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestBucketService_New(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketService_New")
	defer os.RemoveAll(tmpDir)

	expectedDevice := "device"
	expectedSalt := "salt"
	expectedName := "foo"

	svc, err := NewBucketService(tmpDir, expectedDevice, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	bkt := svc.New(expectedName, nil)
	assert.NotNil(t, bkt)
}

func TestBucketService_Save(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketService_Save")
	defer os.RemoveAll(tmpDir)

	expectedDevice := "device"
	expectedSalt := "salt"
	expectedName := "foo"
	expectedMsg := ztring.LoremIpsumWords(10)
	expectedLabels := Labels{
		"foo": "pif0",
		"bar": "paf0",
	}

	svc, err := NewBucketService(tmpDir, expectedDevice, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	before1 := time.Now()

	bkt := svc.New(expectedName, expectedLabels)
	assert.NotNil(t, bkt)

	before2 := time.Now()

	err = bkt.WriteString(expectedMsg)
	assert.NoError(t, err)

	before3 := time.Now()

	err = bkt.Save()
	assert.NoError(t, err)

	after := time.Now()

	// Check Header
	require.NotNil(t, bkt.header)
	assert.Equal(t, expectedName, bkt.header.Name)
	assert.Equal(t, expectedLabels, bkt.header.Labels)
	assert.Less(t, before1, *bkt.header.Created)
	assert.Less(t, before2, *bkt.header.Created)
	assert.Less(t, before3, *bkt.header.Created)
	assert.Greater(t, after, *bkt.header.Created)

	// Check Metadata
	require.NotNil(t, bkt.metadata)
	assert.Equal(t, len([]byte(expectedMsg)), bkt.metadata.Size)
	assert.Equal(t, *bkt.header.Created, *bkt.metadata.Updated)
	assert.Greater(t, after, *bkt.metadata.Updated)
	assert.Equal(t, 0, bkt.metadata.Version)

	// Check Root Layer
	require.NotNil(t, bkt.layerIt)
	k := 0
	for err, l := range bkt.layerIt {
		k++
		assert.NoError(t, err)
		assert.NotNil(t, l)
		assert.Equal(t, []byte(expectedMsg), l.Content)
	}
	assert.Equal(t, 1, k)

	// assert.Equal(t, expectedMsg, bkt.data)
}

func TestBucketService_Save_And_Get(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketService_Save_And_Get")
	defer os.RemoveAll(tmpDir)

	expectedDevice := "device"
	expectedSalt := "salt"
	expectedName := "foo"
	expectedMsg := ztring.LoremIpsumWords(10)
	expectedLabels := Labels{
		"foo": "pif",
		"bar": "paf",
	}

	svc, err := NewBucketService(tmpDir, expectedDevice, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	bkt1 := svc.New(expectedName, expectedLabels)
	assert.NotNil(t, bkt1)

	err = bkt1.WriteString(expectedMsg)
	assert.NoError(t, err)

	before := time.Now()
	err = bkt1.Save()
	assert.NoError(t, err)
	after := time.Now()

	bkt2, err := svc.Get(bkt1.header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bkt2)

	// Check Header
	require.NotNil(t, bkt2.header)
	// Use ExportedValues comparison because of time.Time comparison not working (some private fields are not the same)
	assert.EqualExportedValues(t, bkt1.header, bkt2.header)
	assert.Equal(t, expectedName, bkt2.header.Name)
	assert.Equal(t, expectedLabels, bkt2.header.Labels)

	// Check Metadata
	require.NotNil(t, bkt2.metadata)
	// Use ExportedValues comparison because of time.Time comparison not working (some private fields are not the same)
	assert.EqualExportedValues(t, *bkt1.metadata, *bkt2.metadata)
	assert.Equal(t, len(expectedMsg), bkt2.metadata.Size)
	require.NotNil(t, bkt2.metadata.Updated)
	assert.Less(t, before, *bkt2.metadata.Updated)
	assert.Greater(t, after, *bkt2.metadata.Updated)
	assert.Equal(t, 0, bkt2.metadata.Version)

	// Check Root Layer
	require.NotNil(t, bkt2.layerIt)
	var rootLayer1, rootLayer2 *Layer
	k := 0
	for err, l := range bkt1.layerIt {
		k++
		assert.NoError(t, err)
		assert.NotNil(t, l)
		rootLayer1 = l
	}
	assert.Equal(t, 1, k)
	k = 0
	for err, l := range bkt2.layerIt {
		k++
		assert.NoError(t, err)
		require.NotNil(t, l)
		rootLayer2 = l
	}
	assert.Equal(t, 1, k)

	assert.Equal(t, []byte(expectedMsg), rootLayer2.Content)
	assert.Equal(t, 0, rootLayer2.Metadata.Version)
	assert.Equal(t, len([]byte(expectedMsg)), rootLayer2.Metadata.Size)
	require.NotNil(t, rootLayer2.Metadata.Updated)
	assert.Less(t, before, *rootLayer2.Metadata.Updated)
	assert.Greater(t, after, *rootLayer2.Metadata.Updated)

	// Use ExportedValues comparison because of time.Time comparison not working (some private fields are not the same)
	assert.EqualExportedValues(t, rootLayer1, rootLayer2)
}

func TestBucketService_Names(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketService_Names")
	defer os.RemoveAll(tmpDir)

	expectedDevice := "device"
	expectedSalt := "salt"
	expectedName1 := "foo"
	expectedMsg1 := ztring.LoremIpsumWords(10)
	expectedName2 := "bar"
	expectedMsg2 := ztring.LoremIpsumWords(15)
	expectedName3 := "baz"
	expectedMsg3 := ztring.LoremIpsumWords(20)
	expectedLabels := Labels{
		"foo": "pif",
		"bar": "paf",
	}

	svc, err := NewBucketService(tmpDir, expectedDevice, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// Make some buckets
	bkt1 := svc.New(expectedName1, expectedLabels)
	assert.NotNil(t, bkt1)
	err = bkt1.WriteString(expectedMsg1)
	assert.NoError(t, err)
	err = bkt1.Save()
	assert.NoError(t, err)

	bkt2 := svc.New(expectedName2, expectedLabels)
	assert.NotNil(t, bkt2)
	err = bkt2.WriteString(expectedMsg2)
	assert.NoError(t, err)
	err = bkt2.Save()
	assert.NoError(t, err)

	bkt3 := svc.New(expectedName3, expectedLabels)
	assert.NotNil(t, bkt3)
	err = bkt3.WriteString(expectedMsg3)
	assert.NoError(t, err)
	err = bkt3.Save()
	assert.NoError(t, err)

	names, err := svc.Names()
	assert.NoError(t, err)
	require.NotNil(t, names)
	assert.Contains(t, collectionz.Keys(names), expectedName1)
	assert.Contains(t, collectionz.Keys(names), expectedName2)
	assert.Contains(t, collectionz.Keys(names), expectedName3)
}

func TestBucketService_Save_Edit_Save_Get(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketService_Save_Edit_Save_Get")
	defer os.RemoveAll(tmpDir)

	expectedDevice := "device"
	expectedSalt := "salt"
	expectedName := "foo"
	expectedMsg1 := ztring.LoremIpsumWords(10)
	expectedMsg2 := ztring.LoremIpsumWords(15)
	expectedLabels := Labels{
		"foo": "pif",
		"bar": "paf",
	}

	svc, err := NewBucketService(tmpDir, expectedDevice, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	bkt1 := svc.New(expectedName, expectedLabels)
	assert.NotNil(t, bkt1)

	err = bkt1.WriteString(expectedMsg1)
	assert.NoError(t, err)

	before1 := time.Now()
	err = bkt1.Save()
	assert.NoError(t, err)
	after1 := time.Now()

	err = bkt1.WriteString(expectedMsg2)
	assert.NoError(t, err)

	before2 := time.Now()
	err = bkt1.Save()
	assert.NoError(t, err)
	after2 := time.Now()

	bkt2, err := svc.Get(bkt1.header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bkt2)

	// Check Header
	require.NotNil(t, bkt2.header)
	// Use ExportedValues comparison because of time.Time comparison not working (some private fields are not the same)
	assert.EqualExportedValues(t, bkt1.header, bkt2.header)
	assert.Equal(t, expectedName, bkt2.header.Name)
	assert.Equal(t, expectedLabels, bkt2.header.Labels)

	// Check Metadata
	require.NotNil(t, bkt1.metadata)
	require.NotNil(t, bkt2.metadata)
	// Use ExportedValues comparison because of time.Time comparison not working (some private fields are not the same)
	assert.EqualExportedValues(t, *bkt1.metadata, *bkt2.metadata)
	assert.Equal(t, len(expectedMsg2), bkt2.metadata.Size)
	require.NotNil(t, bkt2.metadata.Updated)
	assert.Less(t, before2, *bkt2.metadata.Updated)
	assert.Greater(t, after2, *bkt2.metadata.Updated)
	assert.Equal(t, 0, bkt2.metadata.Version)

	// Check Root Layer
	require.NotNil(t, bkt2.layerIt)
	var bkt1Layers, bkt2Layers []*Layer
	k := 0
	for err, l := range bkt1.layerIt {
		k++
		assert.NoError(t, err)
		assert.NotNil(t, l)
		if k == 0 {
			assert.Equal(t, []byte(expectedMsg2), l.Content)
		}
		if k == 1 {
			assert.Equal(t, []byte(expectedMsg1), l.Content)
		}
		bkt1Layers = append(bkt1Layers, l)
	}
	assert.Equal(t, 2, k)
	k = 0
	for err, l := range bkt2.layerIt {
		k++
		assert.NoError(t, err)
		require.NotNil(t, l)
		if k == 0 {
			assert.Equal(t, []byte(expectedMsg2), l.Content)
		}
		if k == 1 {
			assert.Equal(t, []byte(expectedMsg1), l.Content)
		}
		bkt2Layers = append(bkt2Layers, l)
	}
	assert.Equal(t, 2, k)

	require.True(t, len(bkt1Layers) >= 1)
	assert.Equal(t, []byte(expectedMsg2), bkt1Layers[0].Content)
	assert.Equal(t, 0, bkt1Layers[0].Metadata.Version)
	assert.Equal(t, len([]byte(expectedMsg2)), bkt1Layers[0].Metadata.Size)
	require.NotNil(t, bkt1Layers[0].Metadata.Updated)
	assert.Less(t, before2, *bkt1Layers[0].Metadata.Updated)
	assert.Greater(t, after2, *bkt1Layers[0].Metadata.Updated)

	require.True(t, len(bkt1Layers) >= 2)
	assert.Equal(t, []byte(expectedMsg1), bkt1Layers[1].Content)
	assert.Equal(t, 0, bkt1Layers[1].Metadata.Version)
	assert.Equal(t, len([]byte(expectedMsg1)), bkt1Layers[1].Metadata.Size)
	require.NotNil(t, bkt1Layers[1].Metadata.Updated)
	assert.Less(t, before1, *bkt1Layers[1].Metadata.Updated)
	assert.Greater(t, after1, *bkt1Layers[1].Metadata.Updated)

	// Use ExportedValues comparison because of time.Time comparison not working (some private fields are not the same)
	assert.EqualExportedValues(t, bkt1Layers[0], bkt2Layers[0])
	assert.EqualExportedValues(t, bkt1Layers[1], bkt2Layers[1])
}

func TestBucketService_Save_Get_Edit_Save_Get(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketService_Save_Get_Edit_Save_Get")
	defer os.RemoveAll(tmpDir)

	// TODO
}

func TestBucketService_Project(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketService_Project")
	defer os.RemoveAll(tmpDir)

	// TODO
}
