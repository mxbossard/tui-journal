package bucket

import (
	"os"
	"testing"
	"time"

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

	svc, err := NewBucketService(tmpDir, expectedDevice, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	bkt := svc.New(expectedName, nil)
	assert.NotNil(t, bkt)

	err = bkt.WriteString(expectedMsg)
	assert.NoError(t, err)

	err = bkt.Save()
	assert.NoError(t, err)
}

func TestBucketService_Save_Get(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBucketService_Save_Get")
	defer os.RemoveAll(tmpDir)

	expectedDevice := "device"
	expectedSalt := "salt"
	expectedName := "foo"
	expectedMsg := ztring.LoremIpsumWords(10)

	svc, err := NewBucketService(tmpDir, expectedDevice, expectedSalt)
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	bkt := svc.New(expectedName, nil)
	assert.NotNil(t, bkt)

	err = bkt.WriteString(expectedMsg)
	assert.NoError(t, err)

	before := time.Now()
	err = bkt.Save()
	assert.NoError(t, err)
	after := time.Now()

	bkt2, err := svc.Get(bkt.header.Uid)
	assert.NoError(t, err)
	require.NotNil(t, bkt2)

	assert.Equal(t, bkt.header, bkt2.header)
	assert.Equal(t, expectedName, bkt2.header.Name)
	require.NotNil(t, bkt2.metadata)
	assert.Equal(t, len(expectedMsg), bkt2.metadata.Size)
	assert.Greater(t, before, bkt2.metadata.Updated)
	assert.Less(t, after, bkt2.metadata.Updated)
	assert.Equal(t, 0, bkt2.metadata.Version)

	require.NotNil(t, bkt2.layerIt)
	k := 0
	for err, l := range bkt2.layerIt {
		k++
		assert.NoError(t, err)
		assert.NotNil(t, l)
		assert.Equal(t, expectedMsg, l.Content)
	}
	assert.Equal(t, 1, k)
}
