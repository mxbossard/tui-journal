package bucket

import (
	"os"
	"testing"

	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/ztring"
	"github.com/stretchr/testify/assert"
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
