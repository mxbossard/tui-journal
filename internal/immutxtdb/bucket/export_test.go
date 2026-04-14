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
	err = bktA.WriteText(expectedMsgA1)
	assert.NoError(t, err)
	err = svc1.Save(bktA, expectedPartition1)
	assert.NoError(t, err)

	// Update bucket A
	err = bktA.WriteText(expectedMsgA2)
	assert.NoError(t, err)
	err = svc1.Save(bktA, expectedPartition1)
	assert.NoError(t, err)

	// Make bucket B
	bktB := svc1.New(expectedNameB, expectedLabels)
	assert.NotNil(t, bktB)
	err = bktB.WriteText(expectedMsgB1)
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
