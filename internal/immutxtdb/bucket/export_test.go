package bucket

import (
	"os"
	"testing"

	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/ztring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExport_Export_Then_Import(t *testing.T) {
	tmpDir1 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir1)
	tmpDir2 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir2)

	expectedPartition1 := "device1"
	expectedPartition2 := "device2"
	expectedSalt1 := "salt1"
	expectedSalt2 := "salt2"
	expectedNameA := "foo"
	expectedMsgA1 := ztring.LoremIpsumWords(10)
	expectedMsgA2 := ztring.LoremIpsumWords(15)
	expectedNameB := "bar"
	expectedMsgB1 := ztring.LoremIpsumWords(20)
	expectedLabels := Labels{
		"foo": "pif",
		"bar": "paf",
	}

	svc1, err := NewBucketService(tmpDir1, expectedSalt1)
	assert.NoError(t, err)
	assert.NotNil(t, svc1)

	// Make bucket A
	bktA := svc1.New(expectedNameA, expectedLabels)
	assert.NotNil(t, bktA)
	err = bktA.SetText(expectedMsgA1)
	assert.NoError(t, err)
	err = svc1.Save(bktA, expectedPartition1)
	assert.NoError(t, err)

	// Update bucket A
	err = bktA.SetText(expectedMsgA2)
	assert.NoError(t, err)
	err = svc1.Save(bktA, expectedPartition1)
	assert.NoError(t, err)

	// Make bucket B
	bktB := svc1.New(expectedNameB, expectedLabels)
	assert.NotNil(t, bktB)
	err = bktB.SetText(expectedMsgB1)
	assert.NoError(t, err)
	err = svc1.Save(bktB, expectedPartition1)
	assert.NoError(t, err)

	// Export bktA
	expA, err := svc1.Export(bktA.Header.Uid, false, false, false)
	require.NotNil(t, expA)
	assert.NoError(t, err)

	svc2, err := NewBucketService(tmpDir2, expectedSalt2)
	assert.NoError(t, err)
	require.NotNil(t, svc2)

	// Import bktA
	err = svc2.Import(expA, expectedPartition2)
	assert.NoError(t, err)

	// Check bktA exists in svc2
	bktC, err := svc2.Get(bktA.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bktC)

	assert.EqualExportedValues(t, bktA.Header, bktC.Header)
	assert.EqualExportedValues(t, *bktA.Metadata, *bktC.Metadata)

	v1a, err := bktA.ProjectText(1)
	assert.NoError(t, err)
	v1c, err := bktC.ProjectText(1)
	assert.NoError(t, err)
	assert.Equal(t, v1a, v1c)

	v2a, err := bktA.ProjectText(2)
	assert.NoError(t, err)
	v2c, err := bktC.ProjectText(2)
	assert.NoError(t, err)
	assert.Equal(t, v2a, v2c)
}

func TestExport_ImportTwice(t *testing.T) {
	tmpDir1 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir1)
	tmpDir2 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir2)

	expectedPartition1 := "device1"
	expectedPartition2 := "device2"
	expectedPartition3 := "device3"
	expectedSalt1 := "salt1"
	expectedSalt2 := "salt2"
	expectedNameA := "foo"
	expectedMsgA1 := ztring.LoremIpsumWords(10)
	expectedMsgA2 := ztring.LoremIpsumWords(15)
	expectedNameB := "bar"
	expectedMsgB1 := ztring.LoremIpsumWords(20)
	expectedLabels := Labels{
		"foo": "pif",
		"bar": "paf",
	}

	svc1, err := NewBucketService(tmpDir1, expectedSalt1)
	assert.NoError(t, err)
	assert.NotNil(t, svc1)

	// Make bucket A
	bktA := svc1.New(expectedNameA, expectedLabels)
	assert.NotNil(t, bktA)
	err = bktA.SetText(expectedMsgA1)
	assert.NoError(t, err)
	err = svc1.Save(bktA, expectedPartition1)
	assert.NoError(t, err)

	// Update bucket A
	err = bktA.SetText(expectedMsgA2)
	assert.NoError(t, err)
	err = svc1.Save(bktA, expectedPartition1)
	assert.NoError(t, err)

	// Make bucket B
	bktB := svc1.New(expectedNameB, expectedLabels)
	assert.NotNil(t, bktB)
	err = bktB.SetText(expectedMsgB1)
	assert.NoError(t, err)
	err = svc1.Save(bktB, expectedPartition1)
	assert.NoError(t, err)

	// Export bktA from svc1
	expA, err := svc1.Export(bktA.Header.Uid, false, false, false)
	require.NotNil(t, expA)
	assert.NoError(t, err)

	svc2, err := NewBucketService(tmpDir2, expectedSalt2)
	assert.NoError(t, err)
	require.NotNil(t, svc2)

	// Import bktA in svc2
	err = svc2.Import(expA, expectedPartition2)
	assert.NoError(t, err)

	// Check bktA exists in svc2
	bktA2, err := svc2.Get(bktA.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bktA2)

	assert.EqualExportedValues(t, bktA.Header, bktA2.Header)
	assert.EqualExportedValues(t, *bktA.Metadata, *bktA2.Metadata)

	txtA1, err := bktA.ProjectText(1)
	assert.NoError(t, err)
	txtA2, err := bktA2.ProjectText(1)
	assert.NoError(t, err)
	assert.Equal(t, txtA1, txtA2)

	txtA1, err = bktA.ProjectText(2)
	assert.NoError(t, err)
	txtA2, err = bktA2.ProjectText(2)
	assert.NoError(t, err)
	assert.Equal(t, txtA1, txtA2)

	// ReImport bktA in svc2 in same partition
	err = svc2.Import(expA, expectedPartition2)
	assert.NoError(t, err)

	// ReImport bktA in svc2 in different partition
	err = svc2.Import(expA, expectedPartition3)
	assert.NoError(t, err)
}

func TestExport_ImportSuccessiveSaves(t *testing.T) {
	tmpDir1 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir1)
	tmpDir2 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir2)

	expectedPartition1 := "device1"
	expectedSalt1 := "salt1"
	expectedSalt2 := "salt2"
	expectedNameA := "foo"
	expectedMsgA1 := ztring.LoremIpsumWords(10)
	expectedMsgA2 := ztring.LoremIpsumWords(12)
	expectedMsgA3 := ztring.LoremIpsumWords(14)
	expectedLabels := Labels{
		"foo": "pif",
		"bar": "paf",
	}

	svc1, err := NewBucketService(tmpDir1, expectedSalt1)
	assert.NoError(t, err)
	assert.NotNil(t, svc1)

	// Make bucket A
	bktA := svc1.New(expectedNameA, expectedLabels)
	assert.NotNil(t, bktA)
	err = bktA.SetText(expectedMsgA1)
	assert.NoError(t, err)
	err = svc1.Save(bktA, expectedPartition1)
	assert.NoError(t, err)

	// Update bucket A
	err = bktA.SetText(expectedMsgA2)
	assert.NoError(t, err)
	err = svc1.Save(bktA, expectedPartition1)
	assert.NoError(t, err)

	// Export bktA from svc1
	expA, err := svc1.Export(bktA.Header.Uid, false, false, false)
	require.NotNil(t, expA)
	assert.NoError(t, err)

	svc2, err := NewBucketService(tmpDir2, expectedSalt2)
	assert.NoError(t, err)
	require.NotNil(t, svc2)

	// Import bktA in svc2
	err = svc2.Import(expA, expectedPartition1)
	assert.NoError(t, err)

	// Check bktA exists in svc2
	bktA2, err := svc2.Get(bktA.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bktA2)
	assert.EqualExportedValues(t, bktA.Header, bktA2.Header)
	assert.EqualExportedValues(t, *bktA.Metadata, *bktA2.Metadata)
	txtA2, err := bktA2.ProjectText()
	assert.NoError(t, err)
	assert.Equal(t, expectedMsgA2, txtA2)
	txtA2, err = bktA2.ProjectText(FirstLayerVersion + 1)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsgA2, txtA2)
	txtA2, err = bktA2.ProjectText(FirstLayerVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsgA1, txtA2)
	txtA2, err = bktA2.ProjectText(FirstLayerVersion + 2)
	assert.Error(t, err)

	// Update BucketA in svc1
	err = bktA.SetText(expectedMsgA3)
	assert.NoError(t, err)
	err = svc1.Save(bktA, expectedPartition1)
	assert.NoError(t, err)

	// Import updated bktA in svc2
	expA2, err := svc1.Export(bktA.Header.Uid, false, false, false)
	require.NotNil(t, expA2)
	assert.NoError(t, err)
	err = svc2.Import(expA2, expectedPartition1)
	assert.NoError(t, err)

	// Check bktA exists in svc2
	bktA3, err := svc2.Get(bktA.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bktA3)
	assert.EqualExportedValues(t, bktA.Header, bktA3.Header)
	assert.EqualExportedValues(t, *bktA.Metadata, *bktA3.Metadata)
	txtA3, err := bktA2.ProjectText()
	assert.NoError(t, err)
	assert.Equal(t, expectedMsgA3, txtA3)
	txtA3, err = bktA2.ProjectText(FirstLayerVersion + 2)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsgA3, txtA3)
	txtA3, err = bktA2.ProjectText(FirstLayerVersion + 1)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsgA2, txtA3)
	txtA3, err = bktA2.ProjectText(FirstLayerVersion)
	assert.NoError(t, err)
	assert.Equal(t, expectedMsgA1, txtA3)
	txtA3, err = bktA2.ProjectText(FirstLayerVersion + 3)
	assert.Error(t, err)
}

func TestExport_ImportOlderVersion(t *testing.T) {
	tmpDir1 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir1)
	tmpDir2 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir2)

	expectedPartition1 := "device1"
	expectedSalt1 := "salt1"
	expectedSalt2 := "salt2"
	expectedNameA := "foo"
	expectedMsgA1 := ztring.LoremIpsumWords(10)
	expectedMsgA2 := ztring.LoremIpsumWords(15)
	expectedLabels := Labels{
		"foo": "pif",
		"bar": "paf",
	}

	svc1, err := NewBucketService(tmpDir1, expectedSalt1)
	assert.NoError(t, err)
	require.NotNil(t, svc1)
	svc2, err := NewBucketService(tmpDir2, expectedSalt2)
	assert.NoError(t, err)
	require.NotNil(t, svc2)

	// Make bucket A
	bktA := svc1.New(expectedNameA, expectedLabels)
	assert.NotNil(t, bktA)
	err = bktA.SetText(expectedMsgA1)
	assert.NoError(t, err)
	err = svc1.Save(bktA, expectedPartition1)
	assert.NoError(t, err)

	// Export bucket A
	expA, err := svc1.Export(bktA.Header.Uid, false, false, false)
	require.NotNil(t, expA)
	assert.NoError(t, err)

	// Get a new bucket A => A2
	bktA2, err := svc1.Get(bktA.Header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bktA2)

	// Update bucket A2
	err = bktA2.SetText(expectedMsgA2)
	assert.NoError(t, err)
	err = svc1.Save(bktA2, expectedPartition1)
	assert.NoError(t, err)

	// Export / Import bucket A2 in svc2
	expA2, err := svc1.Export(bktA.Header.Uid, false, false, false)
	require.NotNil(t, expA2)
	assert.NoError(t, err)
	err = svc2.Import(expA2, expectedPartition1)
	assert.NoError(t, err)

	// Import bucket A in svc2
	err = svc2.Import(expA, expectedPartition1)
	assert.Error(t, err)
	assert.Equal(t, ErrVersionMissmatch, err)
}

func TestExport_ImportNotConsistentLayers(t *testing.T) {
	tmpDir1 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir1)
	tmpDir2 := filez.MkdirTempOrPanic(t.Name())
	defer os.RemoveAll(tmpDir2)

	expectedPartition1 := "device1"
	expectedSalt1 := "salt1"
	expectedSalt2 := "salt2"
	expectedNameA := "foo"
	expectedMsgA1 := ztring.LoremIpsumWords(10)
	expectedMsgB1 := ztring.LoremIpsumWords(20)
	expectedMsgB2 := ztring.LoremIpsumWords(22)
	expectedLabels := Labels{
		"foo": "pif",
		"bar": "paf",
	}

	svc1, err := NewBucketService(tmpDir1, expectedSalt1)
	assert.NoError(t, err)
	require.NotNil(t, svc1)
	svc2, err := NewBucketService(tmpDir2, expectedSalt2)
	assert.NoError(t, err)
	require.NotNil(t, svc2)

	// Make bucket A (1 layer)
	bktA := svc1.New(expectedNameA, expectedLabels)
	assert.NotNil(t, bktA)
	err = bktA.SetText(expectedMsgA1)
	assert.NoError(t, err)
	err = svc1.Save(bktA, expectedPartition1)
	assert.NoError(t, err)

	// Export bucket A
	expA, err := svc1.Export(bktA.Header.Uid, false, false, false)
	require.NotNil(t, expA)
	assert.NoError(t, err)

	// Make bucket B (2 layers)
	bktB := svc1.New(expectedNameA, expectedLabels)
	assert.NotNil(t, bktB)
	err = bktB.SetText(expectedMsgB1)
	assert.NoError(t, err)
	err = svc1.Save(bktB, expectedPartition1)
	assert.NoError(t, err)
	err = bktB.SetText(expectedMsgB2)
	assert.NoError(t, err)
	err = svc1.Save(bktB, expectedPartition1)
	assert.NoError(t, err)

	// Export bucket B
	expB, err := svc1.Export(bktB.Header.Uid, false, false, false)
	require.NotNil(t, expB)
	assert.NoError(t, err)

	// Override bucket B Export uid
	for _, h := range expB.headers {
		h.Uid = bktA.Header.Uid
	}

	// Import bucket A in svc2
	err = svc2.Import(expA, expectedPartition1)
	assert.NoError(t, err)

	// Import bucket B in svc2
	err = svc2.Import(expB, expectedPartition1)
	assert.Error(t, err)
	assert.Equal(t, ErrInconsistentLayers, err)
}
