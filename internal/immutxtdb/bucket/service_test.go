package bucket

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/zip"
	"github.com/mxbossard/utilz/collectionz"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/iterz"
	"github.com/mxbossard/utilz/ztring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dayTime(value string) *time.Time {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		panic(err)
	}
	return &t
}

func Test_NewBucketService(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("Test_NewBucketService")
	defer os.RemoveAll(tmpDir)

	expectedSalt := "salt"
	svc, err := NewBucketService(tmpDir, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestBucketService_New(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketService_New")
	defer os.RemoveAll(tmpDir)

	expectedSalt := "salt"
	expectedName := "foo"

	svc, err := NewBucketService(tmpDir, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	bkt := svc.New(expectedName, nil)
	assert.NotNil(t, bkt)
}

func TestBucketService_Save(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketService_Save")
	defer os.RemoveAll(tmpDir)

	expectedPartition := "device"
	expectedSalt := "salt"
	expectedName := "foo"
	expectedMsg := ztring.LoremIpsumWords(10)
	expectedZipedMsg, err := zip.ZipString(expectedMsg)
	require.NoError(t, err)
	expectedLabels := Labels{
		"foo": "pif0",
		"bar": "paf0",
	}

	svc, err := NewBucketService(tmpDir, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	before1 := time.Now()

	bkt := svc.New(expectedName, expectedLabels)
	assert.NotNil(t, bkt)

	before2 := time.Now()

	err = bkt.UpdateText(expectedMsg)
	assert.NoError(t, err)

	before3 := time.Now()

	err = svc.Save(bkt, expectedPartition)
	assert.NoError(t, err)

	after := time.Now()

	// Check Header
	require.NotNil(t, bkt.Header)
	assert.Equal(t, expectedName, bkt.Header.Name)
	assert.Equal(t, expectedLabels, bkt.Header.Labels)
	assert.Less(t, before1, *bkt.Header.Created)
	assert.Less(t, before2, *bkt.Header.Created)
	assert.Less(t, before3, *bkt.Header.Created)
	assert.Greater(t, after, *bkt.Header.Created)

	// Check Metadata
	require.NotNil(t, bkt.Metadata)
	assert.Equal(t, len([]byte(expectedMsg)), bkt.Metadata.Size)
	assert.Equal(t, *bkt.Header.Created, *bkt.Metadata.Updated)
	assert.Greater(t, after, *bkt.Metadata.Updated)
	assert.Equal(t, FirstLayerVersion, bkt.Metadata.Version)

	// Check Root Layer
	layerIt, err := bkt.LayerIt(LatestVersion)
	assert.NoError(t, err)
	require.NotNil(t, layerIt)
	k := 0
	for err, l := range layerIt {
		k++
		assert.NoError(t, err)
		assert.NotNil(t, l)
		assert.Equal(t, expectedZipedMsg, l.Content)
		assert.Equalf(t, Version(k), l.Metadata.Version, "bad layer #%d version", k)
	}
	assert.Equal(t, 1, k)

	// assert.Equal(t, expectedMsg, bkt.data)
}

func TestBucketService_Get_Not_Existing(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketService_Get_Not_Existing")
	defer os.RemoveAll(tmpDir)

	expectedSalt := "salt"
	expectedBucketUid := BucketUid([]byte{42, 17, 42, 17, 42, 17, 42, 17, 42, 17, 42, 17, 42, 17, 42, 17})

	svc, err := NewBucketService(tmpDir, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	bkt, err := svc.Get(expectedBucketUid)
	assert.ErrorIs(t, err, ErrNotExist)
	assert.Nil(t, bkt)
}

func TestBucketService_Save_Then_Get(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir)

	expectedPartition := "device"
	expectedSalt := "salt"
	expectedName := "foo"
	expectedMsg := ztring.LoremIpsumWords(10)
	expectedZipedMsg, err := zip.ZipString(expectedMsg)
	require.NoError(t, err)
	expectedLabels := Labels{
		"foo": "pif",
		"bar": "paf",
	}

	svc, err := NewBucketService(tmpDir, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	bkt1 := svc.New(expectedName, expectedLabels)
	assert.NotNil(t, bkt1)

	err = bkt1.UpdateText(expectedMsg)
	assert.NoError(t, err)

	before := time.Now()
	err = svc.Save(bkt1, expectedPartition)
	assert.NoError(t, err)
	after := time.Now()

	bkt2, err := svc.Get(bkt1.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bkt2)

	// Check Header
	require.NotNil(t, bkt2.Header)
	// Use ExportedValues comparison because of time.Time comparison not working (some private fields are not the same)
	assert.EqualExportedValues(t, bkt1.Header, bkt2.Header)
	assert.Equal(t, expectedName, bkt2.Header.Name)
	assert.Equal(t, expectedLabels, bkt2.Header.Labels)

	// Check Metadata
	require.NotNil(t, bkt2.Metadata)
	// Use ExportedValues comparison because of time.Time comparison not working (some private fields are not the same)
	assert.EqualExportedValues(t, *bkt1.Metadata, *bkt2.Metadata)
	assert.Equal(t, len(expectedMsg), bkt2.Metadata.Size)
	require.NotNil(t, bkt2.Metadata.Updated)
	assert.Less(t, before, *bkt2.Metadata.Updated)
	assert.Greater(t, after, *bkt2.Metadata.Updated)
	assert.Equal(t, FirstLayerVersion, bkt2.Metadata.Version)

	// Check Root Layer
	layerIt1, err := bkt1.LayerIt(LatestVersion)
	assert.NoError(t, err)
	require.NotNil(t, layerIt1)
	var rootLayer1, rootLayer2 *Layer
	k := 0
	for err, l := range layerIt1 {
		k++
		assert.NoError(t, err)
		assert.NotNil(t, l)
		rootLayer1 = l
	}
	assert.Equal(t, 1, k)
	layerIt2, err := bkt2.LayerIt(LatestVersion)
	assert.NoError(t, err)
	require.NotNil(t, layerIt2)
	k = 0
	for err, l := range layerIt2 {
		k++
		assert.NoError(t, err)
		require.NotNil(t, l)
		rootLayer2 = l
	}
	assert.Equal(t, 1, k)

	require.NotNil(t, rootLayer2)
	assert.Equal(t, expectedZipedMsg, rootLayer2.Content)
	assert.Equal(t, FirstLayerVersion, rootLayer2.Metadata.Version)
	assert.Equal(t, len(expectedMsg), rootLayer2.Metadata.Size)
	require.NotNil(t, rootLayer2.Metadata.Updated)
	assert.Less(t, before, *rootLayer2.Metadata.Updated)
	assert.Greater(t, after, *rootLayer2.Metadata.Updated)

	// Use ExportedValues comparison because of time.Time comparison not working (some private fields are not the same)
	assert.EqualExportedValues(t, rootLayer1, rootLayer2)
}

func TestBucketService_Names(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir)

	expectedPartition := "device"
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

	svc, err := NewBucketService(tmpDir, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// Make some buckets
	bkt1 := svc.New(expectedName1, expectedLabels)
	assert.NotNil(t, bkt1)
	err = bkt1.UpdateText(expectedMsg1)
	assert.NoError(t, err)
	err = svc.Save(bkt1, expectedPartition)
	assert.NoError(t, err)

	bkt2 := svc.New(expectedName2, expectedLabels)
	assert.NotNil(t, bkt2)
	err = bkt2.UpdateText(expectedMsg2)
	assert.NoError(t, err)
	err = svc.Save(bkt2, expectedPartition)
	assert.NoError(t, err)

	bkt3 := svc.New(expectedName3, expectedLabels)
	assert.NotNil(t, bkt3)
	err = bkt3.UpdateText(expectedMsg3)
	assert.NoError(t, err)
	err = svc.Save(bkt3, expectedPartition)
	assert.NoError(t, err)

	names, err := svc.Names()
	assert.NoError(t, err)
	require.NotNil(t, names)
	assert.Len(t, names, 3)
	assert.Contains(t, collectionz.Keys(names), expectedName1)
	assert.Contains(t, collectionz.Keys(names), expectedName2)
	assert.Contains(t, collectionz.Keys(names), expectedName3)
}

func TestBucketService_Names_Multipart(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir)

	expectedPart1 := "part1"
	expectedPart2 := "part2"
	expectedPart3 := "part3"
	expectedPart4 := expectedPart1
	expectedSalt := "salt"
	expectedName1 := "foo"
	expectedMsg1 := ztring.LoremIpsumWords(10)
	expectedName2 := "bar"
	expectedMsg2 := ztring.LoremIpsumWords(15)
	expectedName3 := "baz"
	expectedMsg3 := ztring.LoremIpsumWords(20)
	expectedName4 := expectedName2
	expectedMsg4 := ztring.LoremIpsumWords(25)
	expectedLabels := Labels{
		"foo": "pif",
		"bar": "paf",
	}

	svc, err := NewBucketService(tmpDir, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// Make some buckets
	bkt1 := svc.New(expectedName1, expectedLabels)
	assert.NotNil(t, bkt1)
	err = bkt1.UpdateText(expectedMsg1)
	assert.NoError(t, err)
	err = svc.Save(bkt1, expectedPart1)
	assert.NoError(t, err)

	bkt2 := svc.New(expectedName2, expectedLabels)
	assert.NotNil(t, bkt2)
	err = bkt2.UpdateText(expectedMsg2)
	assert.NoError(t, err)
	err = svc.Save(bkt2, expectedPart2)
	assert.NoError(t, err)

	bkt3 := svc.New(expectedName3, expectedLabels)
	assert.NotNil(t, bkt3)
	err = bkt3.UpdateText(expectedMsg3)
	assert.NoError(t, err)
	err = svc.Save(bkt3, expectedPart3)
	assert.NoError(t, err)

	bkt4 := svc.New(expectedName4, expectedLabels)
	assert.NotNil(t, bkt4)
	err = bkt4.UpdateText(expectedMsg4)
	assert.NoError(t, err)
	err = svc.Save(bkt4, expectedPart4)
	assert.NoError(t, err)

	names, err := svc.Names()
	assert.NoError(t, err)
	require.NotNil(t, names)
	assert.Len(t, names, 3)
	assert.Contains(t, collectionz.Keys(names), expectedName1)
	assert.Contains(t, collectionz.Keys(names), expectedName2)
	assert.Contains(t, collectionz.Keys(names), expectedName3)
}

func TestBucketService_Save_Edit_Save_Get(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketService_Save_Edit_Save_Get")
	defer os.RemoveAll(tmpDir)

	expectedPartition := "device"
	expectedSalt := "salt"
	expectedName := "foo"
	expectedMsg1 := ztring.LoremIpsumWords(10)
	expectedMsg2 := ztring.LoremIpsumWords(15)
	expectedLabels := Labels{
		"foo": "pif",
		"bar": "paf",
	}

	svc, err := NewBucketService(tmpDir, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	bkt1 := svc.New(expectedName, expectedLabels)
	assert.NotNil(t, bkt1)

	// First write
	err = bkt1.UpdateText(expectedMsg1)
	assert.NoError(t, err)

	before1 := time.Now()
	err = svc.Save(bkt1, expectedPartition)
	assert.NoError(t, err)
	after1 := time.Now()

	// Check Metadata
	require.NotNil(t, bkt1.Metadata)
	assert.Equal(t, FirstLayerVersion, bkt1.Metadata.Version)
	assert.Equal(t, len(expectedMsg1), bkt1.Metadata.Size)
	// assert.Equal(t, nil, bkt1.metadata.Updated)
	assert.Less(t, before1, *bkt1.Metadata.Updated)
	assert.Greater(t, after1, *bkt1.Metadata.Updated)

	// Check text projection
	text, err := bkt1.ProjectText(LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsg1, text)

	err = bkt1.UpdateText(expectedMsg2)
	assert.NoError(t, err)

	// Second write
	before2 := time.Now()
	err = svc.Save(bkt1, expectedPartition)
	assert.NoError(t, err)
	after2 := time.Now()

	// Check Metadata
	require.NotNil(t, bkt1.Metadata)
	assert.Equal(t, FirstLayerVersion+1, bkt1.Metadata.Version)
	assert.Equal(t, len(expectedMsg2), bkt1.Metadata.Size)
	assert.Less(t, before2, *bkt1.Metadata.Updated)
	assert.Greater(t, after2, *bkt1.Metadata.Updated)

	// Check text projection
	text, err = bkt1.ProjectText(LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsg2, text)

	// Get the Bucket
	bkt2, err := svc.Get(bkt1.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bkt2)

	// Check Header
	require.NotNil(t, bkt2.Header)
	// Use ExportedValues comparison because of time.Time comparison not working (some private fields are not the same)
	assert.EqualExportedValues(t, bkt1.Header, bkt2.Header)
	assert.Equal(t, expectedName, bkt2.Header.Name)
	assert.Equal(t, expectedLabels, bkt2.Header.Labels)

	// Check Metadata
	require.NotNil(t, bkt2.Metadata)
	assert.Equal(t, FirstLayerVersion+1, bkt2.Metadata.Version)
	assert.Equal(t, len(expectedMsg2), bkt2.Metadata.Size)
	assert.Less(t, before2, *bkt2.Metadata.Updated)
	assert.Greater(t, after2, *bkt2.Metadata.Updated)

	// Check Layers
	layerIt1, err := bkt2.LayerIt(LatestVersion)
	assert.NoError(t, err)
	require.NotNil(t, layerIt1)
	var bkt1Layers, bkt2Layers []*Layer
	k := 0
	for err, l := range layerIt1 {
		assert.NoError(t, err)
		assert.NotNil(t, l)
		// if k == 0 {
		// 	assert.Equal(t, []byte(expectedMsg2), l.Content)
		// }
		// if k == 1 {
		// 	assert.Equal(t, []byte(expectedMsg1), l.Content)
		// }
		bkt1Layers = append(bkt1Layers, l)
		k++
	}
	assert.Equal(t, 2, k)
	k = 0
	layerIt2, err := bkt2.LayerIt(LatestVersion)
	assert.NoError(t, err)
	require.NotNil(t, layerIt2)
	for err, l := range layerIt2 {

		assert.NoError(t, err)
		require.NotNil(t, l)
		// if k == 0 {
		// 	assert.Equal(t, []byte(expectedMsg2), l.Content)
		// }
		// if k == 1 {
		// 	assert.Equal(t, []byte(expectedMsg1), l.Content)
		// }
		bkt2Layers = append(bkt2Layers, l)
		k++
	}
	assert.Equal(t, 2, k)

	require.True(t, len(bkt1Layers) >= 1)
	// assert.Equal(t, []byte(expectedMsg2), bkt1Layers[0].Content)
	require.NotNil(t, bkt1Layers[0].Metadata)
	assert.Equal(t, Version(1), bkt1Layers[0].Metadata.Version)
	assert.Equal(t, len(expectedMsg1), bkt1Layers[0].Metadata.Size)
	require.NotNil(t, bkt1Layers[0].Metadata.Updated)
	assert.Less(t, before1, *bkt1Layers[0].Metadata.Updated)
	assert.Greater(t, after1, *bkt1Layers[0].Metadata.Updated)

	require.True(t, len(bkt1Layers) >= 2)
	// assert.Equal(t, []byte(expectedMsg1), bkt1Layers[1].Content)
	require.NotNil(t, bkt1Layers[1].Metadata)
	assert.Equal(t, Version(2), bkt1Layers[1].Metadata.Version)
	assert.Equal(t, len(expectedMsg2), bkt1Layers[1].Metadata.Size)
	require.NotNil(t, bkt1Layers[1].Metadata.Updated)
	assert.Less(t, before2, *bkt1Layers[1].Metadata.Updated)
	assert.Greater(t, after2, *bkt1Layers[1].Metadata.Updated)

	require.True(t, len(bkt2Layers) >= 2)
	// Use ExportedValues comparison because of time.Time comparison not working (some private fields are not the same)
	assert.EqualExportedValues(t, bkt1Layers[0], bkt2Layers[0])
	assert.EqualExportedValues(t, bkt1Layers[1], bkt2Layers[1])

	// Check text projections
	text, err = bkt1.ProjectText(LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsg2, text)

	text, err = bkt2.ProjectText(LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsg2, text)
}

func TestBucketService_ProjectText(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketService_ProjectText")
	defer os.RemoveAll(tmpDir)

	expectedPartition := "device"
	expectedSalt := "salt"
	expectedName := "foo"
	expectedMsg1 := ztring.LoremIpsumWords(2)
	expectedMsg2 := ztring.LoremIpsumWords(4)
	expectedMsg3 := ztring.LoremIpsumWords(6)
	expectedMsg4 := ztring.LoremIpsumWords(5)

	svc, err := NewBucketService(tmpDir, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	bkt1 := svc.New(expectedName, nil)
	assert.NotNil(t, bkt1)

	// First write
	err = bkt1.UpdateText(expectedMsg1)
	assert.NoError(t, err)
	err = svc.Save(bkt1, expectedPartition)
	assert.NoError(t, err)

	err = bkt1.UpdateText(expectedMsg2)
	assert.NoError(t, err)
	err = svc.Save(bkt1, expectedPartition)
	assert.NoError(t, err)

	err = bkt1.UpdateText(expectedMsg3)
	assert.NoError(t, err)
	err = svc.Save(bkt1, expectedPartition)
	assert.NoError(t, err)

	err = bkt1.UpdateText(expectedMsg4)
	assert.NoError(t, err)
	err = svc.Save(bkt1, expectedPartition)
	assert.NoError(t, err)

	textLatest, err := bkt1.ProjectText(LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsg4, textLatest)
	text4, err := bkt1.ProjectText(4)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsg4, text4)
	text3, err := bkt1.ProjectText(3)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsg3, text3)
	text2, err := bkt1.ProjectText(2)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsg2, text2)
	text1, err := bkt1.ProjectText(1)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsg1, text1)
}

func TestBucketService_EditMultiParts_Then_Get(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir)

	expectedPartition1 := "device1"
	expectedPartition2 := "device2"
	expectedPartition3 := "device3"
	expectedSalt := "salt"
	expectedName := "foo"
	expectedMsg1 := ztring.LoremIpsumWords(5)
	expectedMsg2 := ztring.LoremIpsumWords(15)
	expectedMsg3 := ztring.LoremIpsumWords(10)
	// expectedZipedMsg1, err := zip.ZipString(expectedMsg1)
	// require.NoError(t, err)
	// expectedZipedMsg2, err := zip.ZipString(expectedMsg2)
	// require.NoError(t, err)
	// expectedZipedMsg3, err := zip.ZipString(expectedMsg3)
	// require.NoError(t, err)
	expectedLabels := Labels{
		"foo": "pif",
		"bar": "paf",
	}

	svc, err := NewBucketService(tmpDir, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	bkt1 := svc.New(expectedName, expectedLabels)
	assert.NotNil(t, bkt1)

	assert.Nil(t, bkt1.Metadata)

	// First Save in bkt1
	err = bkt1.UpdateText(expectedMsg1)
	assert.NoError(t, err)

	assert.Nil(t, bkt1.Metadata)

	before1 := time.Now()
	err = svc.Save(bkt1, expectedPartition1)
	assert.NoError(t, err)
	after1 := time.Now()

	assert.Equal(t, Version(1), bkt1.Metadata.Version)

	// Second save in same bkt1
	err = bkt1.UpdateText(expectedMsg2)
	assert.NoError(t, err)
	before2 := time.Now()

	assert.Equal(t, Version(1), bkt1.Metadata.Version)

	err = svc.Save(bkt1, expectedPartition2)
	assert.NoError(t, err)
	after2 := time.Now()

	// ----- Check bkt1 Metadata -----
	assert.Equal(t, Version(2), bkt1.Metadata.Version)

	// ----- Check bkt1 Layers -----
	layerIt1, err := bkt1.LayerIt(LatestVersion)
	assert.NoError(t, err)
	require.NotNil(t, layerIt1)
	k := 0
	for err, l := range layerIt1 {
		k++
		assert.NoError(t, err)
		assert.NotNil(t, l)
		require.NotNil(t, l.Metadata)
		assert.Equal(t, Version(k), l.Metadata.Version)
		switch k {
		case 1:
			assert.Equal(t, len(expectedMsg1), l.Metadata.Size)
			require.NotNil(t, l.Metadata.Updated)
			assert.Less(t, before1, *l.Metadata.Updated)
			assert.Greater(t, after1, *l.Metadata.Updated)
			txt, err := bkt1.ProjectText(Version(k))
			assert.NoError(t, err)
			assert.Equal(t, expectedMsg1, txt)
		case 2:
			assert.Equal(t, len(expectedMsg2), l.Metadata.Size)
			require.NotNil(t, l.Metadata.Updated)
			assert.Less(t, before2, *l.Metadata.Updated)
			assert.Greater(t, after2, *l.Metadata.Updated)
			txt, err := bkt1.ProjectText(Version(k))
			assert.NoError(t, err)
			assert.Equal(t, expectedMsg2, txt)
		default:
			assert.Fail(t, "should have only 2 layers")
		}
	}
	assert.Equal(t, 2, k)

	// Third save in another bkt2
	bkt2, err := svc.Get(bkt1.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bkt2)
	err = bkt2.UpdateText(expectedMsg3)
	assert.NoError(t, err)
	before3 := time.Now()
	err = svc.Save(bkt2, expectedPartition3)
	assert.NoError(t, err)
	after3 := time.Now()

	// ----- Check bkt2 Metadata -----
	assert.Equal(t, Version(3), bkt2.Metadata.Version)

	// ----- Check bkt2 Layers -----
	layerIt2, err := bkt2.LayerIt(LatestVersion)
	assert.NoError(t, err)
	require.NotNil(t, layerIt2)
	k = 0
	for err, l := range layerIt2 {
		k++
		assert.NoError(t, err)
		assert.NotNil(t, l)
		require.NotNil(t, l.Metadata)
		assert.Equal(t, Version(k), l.Metadata.Version)
		switch k {
		case 1:
			assert.Equal(t, len(expectedMsg1), l.Metadata.Size)
			require.NotNil(t, l.Metadata.Updated)
			assert.Less(t, before1, *l.Metadata.Updated)
			assert.Greater(t, after1, *l.Metadata.Updated)
			txt, err := bkt2.ProjectText(Version(k))
			assert.NoError(t, err)
			assert.Equal(t, expectedMsg1, txt)
		case 2:
			assert.Equal(t, len(expectedMsg2), l.Metadata.Size)
			require.NotNil(t, l.Metadata.Updated)
			assert.Less(t, before2, *l.Metadata.Updated)
			assert.Greater(t, after2, *l.Metadata.Updated)
			txt, err := bkt2.ProjectText(Version(k))
			assert.NoError(t, err)
			assert.Equal(t, expectedMsg2, txt)
		case 3:
			assert.Equal(t, len(expectedMsg3), l.Metadata.Size)
			require.NotNil(t, l.Metadata.Updated)
			assert.Less(t, before3, *l.Metadata.Updated)
			assert.Greater(t, after3, *l.Metadata.Updated)
			txt, err := bkt2.ProjectText(Version(k))
			assert.NoError(t, err)
			assert.Equal(t, expectedMsg3, txt)
		default:
			assert.Fail(t, "should have only 3 layers")
		}
	}
	assert.Equal(t, 3, k)

	// Check retrieval of a third bkt3
	bkt3, err := svc.Get(bkt1.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bkt3)

	// Check Header
	require.NotNil(t, bkt3.Header)
	// Use ExportedValues comparison because of time.Time comparison not working (some private fields are not the same)
	assert.EqualExportedValues(t, bkt1.Header, bkt3.Header)
	assert.Equal(t, expectedName, bkt3.Header.Name)
	assert.Equal(t, expectedLabels, bkt3.Header.Labels)

	// Check Metadata
	require.NotNil(t, bkt3.Metadata)
	// Use ExportedValues comparison because of time.Time comparison not working (some private fields are not the same)
	assert.Equal(t, len(expectedMsg3), bkt3.Metadata.Size)
	require.NotNil(t, bkt3.Metadata.Updated)
	assert.Less(t, before3, *bkt3.Metadata.Updated)
	assert.Greater(t, after3, *bkt3.Metadata.Updated)
	assert.Equal(t, Version(3), bkt3.Metadata.Version)

	// ----- Check bkt3 Metadata -----
	assert.Equal(t, Version(3), bkt3.Metadata.Version)

	// ----- Check bkt3 Layers -----
	layerIt3, err := bkt3.LayerIt(LatestVersion)
	assert.NoError(t, err)
	require.NotNil(t, layerIt3)
	k = 0
	for err, l := range layerIt3 {
		k++
		assert.NoError(t, err)
		assert.NotNil(t, l)
		require.NotNil(t, l.Metadata)
		assert.Equal(t, Version(k), l.Metadata.Version)
		switch k {
		case 1:
			assert.Equal(t, len(expectedMsg1), l.Metadata.Size)
			require.NotNil(t, l.Metadata.Updated)
			assert.Less(t, before1, *l.Metadata.Updated)
			assert.Greater(t, after1, *l.Metadata.Updated)
			txt, err := bkt3.ProjectText(Version(k))
			assert.NoError(t, err)
			assert.Equal(t, expectedMsg1, txt)
		case 2:
			assert.Equal(t, len(expectedMsg2), l.Metadata.Size)
			require.NotNil(t, l.Metadata.Updated)
			assert.Less(t, before2, *l.Metadata.Updated)
			assert.Greater(t, after2, *l.Metadata.Updated)
			txt, err := bkt3.ProjectText(Version(k))
			assert.NoError(t, err)
			assert.Equal(t, expectedMsg2, txt)
		case 3:
			assert.Equal(t, len(expectedMsg3), l.Metadata.Size)
			require.NotNil(t, l.Metadata.Updated)
			assert.Less(t, before3, *l.Metadata.Updated)
			assert.Greater(t, after3, *l.Metadata.Updated)
			txt, err := bkt3.ProjectText(Version(k))
			assert.NoError(t, err)
			assert.Equal(t, expectedMsg3, txt)
		default:
			assert.Fail(t, "should have only 3 layers")
		}
	}
	assert.Equal(t, 3, k, "bad count of layers loaded")

	// Not existing version
	layerIt4, err := bkt3.LayerIt(Version(10))
	assert.Equal(t, ErrNotExist, err)
	assert.Nil(t, layerIt4)
}

func TestBucketService_Filter(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir)

	expectedPartition := "device"
	expectedSalt := "salt"
	expectedName1 := "foo1"
	expectedName2 := "foo2"
	expectedName3 := "bar3"
	expectedName4 := "bar4"
	expectedName5 := "foo5"
	expectedMsg1 := ztring.LoremIpsumWords(5)
	expectedMsg2 := ztring.LoremIpsumWords(10)
	expectedMsg3 := ztring.LoremIpsumWords(15)
	expectedMsg4 := ztring.LoremIpsumWords(20)
	expectedMsg5 := ztring.LoremIpsumWords(25)

	svc, err := NewBucketService(tmpDir, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// Day 1: create bkt1
	bkt1, err := svc.NewText(expectedName1, nil, expectedMsg1)
	assert.NoError(t, err)
	assert.NotNil(t, bkt1)
	day1 := dayTime("2026-03-01")
	// bkt1.Header.Created = day1
	err = svc.Save(bkt1, expectedPartition, *day1)
	assert.NoError(t, err)

	// Day 2: create bkt2
	bkt2, err := svc.NewText(expectedName2, nil, expectedMsg2)
	assert.NoError(t, err)
	assert.NotNil(t, bkt2)
	day2 := dayTime("2026-03-02")
	// bkt2.Header.Created = day2
	err = svc.Save(bkt2, expectedPartition, *day2)
	assert.NoError(t, err)

	// Day 3: create bkt3
	bkt3, err := svc.NewText(expectedName3, nil, expectedMsg3)
	assert.NoError(t, err)
	assert.NotNil(t, bkt3)
	day3 := dayTime("2026-03-03")
	// bkt3.Header.Created = day3
	err = svc.Save(bkt3, expectedPartition, *day3)
	assert.NoError(t, err)

	// Update bkt1
	bkt1.UpdateText(expectedMsg1 + "updated")
	err = svc.Save(bkt1, expectedPartition, *day3)
	assert.NoError(t, err)

	// Day 4: create bkt4
	bkt4, err := svc.NewText(expectedName4, nil, expectedMsg4)
	assert.NoError(t, err)
	assert.NotNil(t, bkt4)
	day4 := dayTime("2026-03-04")
	// bkt4.Header.Created = day4
	err = svc.Save(bkt4, expectedPartition, *day4)
	assert.NoError(t, err)

	// Day 5: create bkt5
	bkt5, err := svc.NewText(expectedName5, nil, expectedMsg5)
	assert.NoError(t, err)
	assert.NotNil(t, bkt5)
	day5 := dayTime("2026-03-05")
	// bkt5.Header.Created = day5
	err = svc.Save(bkt5, expectedPartition, *day5)
	assert.NoError(t, err)

	// Update bkt4
	bkt4.UpdateText(expectedMsg4 + "updated")
	err = svc.Save(bkt4, expectedPartition, *day5)
	assert.NoError(t, err)

	// Between day1 and day5 there is bkt2, bkt3 & bkt4 created
	pgnr, err := svc.Filter(OlderFirst, CreatedBetweenCriterion(*day1, *day5), 1, 1)
	assert.NoError(t, err)
	require.NotNil(t, pgnr)

	k := 0
	for b := range pgnr.All() {
		switch k {
		case 0:
			assert.Equal(t, expectedName2, b.Val().Header.Name)
		case 1:
			assert.Equal(t, expectedName3, b.Val().Header.Name)
		case 2:
			assert.Equal(t, expectedName4, b.Val().Header.Name)
		}
		k++
	}
	assert.Equal(t, 3, k)

	// Between day2 and day5 there is bkt3 & bkt4 created
	pgnr2, err := svc.Filter(OlderFirst, CreatedBetweenCriterion(*day2, *day5), 1, 1)
	assert.NoError(t, err)
	require.NotNil(t, pgnr2)
	entries := iterz.Flatten(pgnr2.All())
	require.NotNil(t, entries)

	// Check first bkt is bkt3
	require.True(t, len(entries) > 0)
	assert.Equal(t, expectedName3, entries[0].Val().Header.Name)
	txt, err := entries[0].Val().ProjectText(LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsg3, txt)

	// Check second bkt is updated bkt4
	require.True(t, len(entries) > 1)
	assert.Equal(t, expectedName4, entries[1].Val().Header.Name)
	txt, err = entries[1].Val().ProjectText(LatestVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsg4+"updated", txt)

	// Check no third bkt
	require.True(t, len(entries) < 3)
}

func TestBucketService_FilterMultipart(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir)

	expectedPart1 := "part1"
	expectedPart2 := "part2"
	expectedPart3 := "part3"
	expectedSalt := "salt"
	expectedName1 := "foo1"
	expectedName2 := "foo2"
	expectedName3 := "bar3"
	expectedName4 := "bar4"
	expectedName5 := "foo5"
	expectedMsg1 := ztring.LoremIpsumWords(5)
	expectedMsg2 := ztring.LoremIpsumWords(10)
	expectedMsg3 := ztring.LoremIpsumWords(15)
	expectedMsg4 := ztring.LoremIpsumWords(20)
	expectedMsg5 := ztring.LoremIpsumWords(25)

	svc, err := NewBucketService(tmpDir, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// Day1: create bkt1
	bkt1, err := svc.NewText(expectedName1, nil, expectedMsg1)
	assert.NoError(t, err)
	assert.NotNil(t, bkt1)
	day1 := dayTime("2026-03-01")
	// bkt1.Header.Created = day1
	err = svc.Save(bkt1, expectedPart1, *day1)
	assert.NoError(t, err)

	// Day2: create bkt2
	bkt2, err := svc.NewText(expectedName2, nil, expectedMsg2)
	assert.NoError(t, err)
	assert.NotNil(t, bkt2)
	day2 := dayTime("2026-03-02")
	// bkt2.Header.Created = day2
	err = svc.Save(bkt2, expectedPart2, *day2)
	assert.NoError(t, err)

	// Day3: create bkt3
	bkt3, err := svc.NewText(expectedName3, nil, expectedMsg3)
	assert.NoError(t, err)
	assert.NotNil(t, bkt3)
	day3 := dayTime("2026-03-03")
	// bkt3.Header.Created = day3
	err = svc.Save(bkt3, expectedPart3, *day3)
	assert.NoError(t, err)

	// Update bkt1
	bkt1.UpdateText(expectedMsg1 + "updated")
	err = svc.Save(bkt1, expectedPart1, *day3)
	assert.NoError(t, err)

	// Day4: create bkt4
	bkt4, err := svc.NewText(expectedName4, nil, expectedMsg4)
	assert.NoError(t, err)
	assert.NotNil(t, bkt4)
	day4 := dayTime("2026-03-04")
	// bkt4.Header.Created = day4
	err = svc.Save(bkt4, expectedPart1, *day4)
	assert.NoError(t, err)

	// Day5: create bkt5
	bkt5, err := svc.NewText(expectedName5, nil, expectedMsg5)
	assert.NoError(t, err)
	assert.NotNil(t, bkt5)
	day5 := dayTime("2026-03-05")
	// bkt5.Header.Created = day5
	err = svc.Save(bkt5, expectedPart2, *day5)
	assert.NoError(t, err)

	// Update bkt4
	bkt4.UpdateText(expectedMsg4 + "updated")
	err = svc.Save(bkt4, expectedPart3, *day5)
	assert.NoError(t, err)

	// Between day1 and day5 there is bkt2, bkt3 & bkt4 created, bkt1 updated
	pgnr, err := svc.Filter(OlderFirst, UpdatedBetweenCriterion(*day1, *day5), 1, 1)
	assert.NoError(t, err)
	require.NotNil(t, pgnr)

	k := 0
	for b := range pgnr.All() {
		switch k {
		case 0:
			assert.Equal(t, expectedName2, b.Val().Header.Name)
		case 1:
			assert.Equal(t, expectedName1, b.Val().Header.Name)
		case 2:
			assert.Equal(t, expectedName3, b.Val().Header.Name)
		case 3:
			assert.Equal(t, expectedName4, b.Val().Header.Name)
		}
		k++
	}
	assert.Equal(t, 4, k)

	// Between day1 and day5 there is bkt2, bkt3 & bkt4 created, bkt1 updated
	pgnr2, err := svc.Filter(OlderFirst, UpdatedBetweenCriterion(*day1, *day5), 1, 1)
	assert.NoError(t, err)
	require.NotNil(t, pgnr2)

	k = 0
	for b := range pgnr2.All() {
		switch k {
		case 0:
			assert.Equal(t, expectedName2, b.Val().Header.Name)
		case 1:
			assert.Equal(t, expectedName1, b.Val().Header.Name)
		case 2:
			assert.Equal(t, expectedName3, b.Val().Header.Name)
		case 3:
			assert.Equal(t, expectedName4, b.Val().Header.Name)
		}
		k++
	}
	assert.Equal(t, 4, k)

	fmt.Printf("-------START--------\n")
	// Between day2 and day5 there is bkt3 & bkt4 created, bkt1 updated
	pgnr3, err := svc.Filter(OlderFirst, UpdatedBetweenCriterion(*day2, *day5), 1, 1)
	assert.NoError(t, err)
	require.NotNil(t, pgnr3)

	k = 0
	for b := range pgnr3.All() {
		switch k {
		case 0:
			assert.Equal(t, expectedName1, b.Val().Header.Name)
		case 1:
			assert.Equal(t, expectedName3, b.Val().Header.Name)
		case 2:
			assert.Equal(t, expectedName4, b.Val().Header.Name)
		}
		k++
	}
	assert.Equal(t, 3, k)
	fmt.Printf("-------END--------\n")
}
