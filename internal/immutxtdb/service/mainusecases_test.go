package service

import (
	"os"
	"testing"
	"time"

	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/ztring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseCaseDump0_Create(t *testing.T) {
	tmpDir := filez.MkTempOrPanic("TestUseCaseDump0_Create")
	defer os.Remove(tmpDir)

	now := time.Now()
	expextedDevice := "foo"
	expectedTxt := ztring.LoremIpsumWords(10)
	d, err := UseCaseDump0_Create(tmpDir, "salt", expextedDevice, expectedTxt)
	assert.NoError(t, err)
	assert.NotNil(t, d)

	assert.NotEmpty(t, d.Uid)
	assert.True(t, d.Metadata.Created.After(now))
	assert.Equal(t, d.Metadata.Created, d.Metadata.Updated)
	assert.NotNil(t, d.LayerRefIt)
	k := 0
	for _, l := range d.LayerRefIt {
		assert.NotNil(t, l)
		k++
	}
	assert.Equal(t, 1, k)
	k = 0
	for _, l := range d.LayerRefIt {
		assert.NotNil(t, l)
		k++
	}
	assert.Equal(t, 1, k)

	txt, err := project(&d.Bucket)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxt, txt)
}

func TestUseCaseDump1_ListLast(t *testing.T) {
	tmpDir := filez.MkTempOrPanic("TestUseCaseDump1_ListLast")
	defer os.Remove(tmpDir)

	expectedSalt := "salt"
	expextedDevice := "foo"
	expectedTxt1 := ztring.LoremIpsumWords(5)
	expectedTxt2 := ztring.LoremIpsumWords(10)
	expectedTxt3 := ztring.LoremIpsumWords(15)

	// Add first dump
	d1, err := UseCaseDump0_Create(tmpDir, expectedSalt, expextedDevice, expectedTxt1)
	assert.NoError(t, err)
	assert.NotNil(t, d1)

	dumps, err := UseCaseDump1_ListLast(tmpDir, expextedDevice, expectedSalt, 5)
	assert.NoError(t, err)
	assert.NotNil(t, dumps)
	require.Len(t, dumps, 1)
	txt1, err := project(&dumps[0].Bucket)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxt1, txt1)

	// Add second dump
	d2, err := UseCaseDump0_Create(tmpDir, expectedSalt, expextedDevice, expectedTxt2)
	assert.NoError(t, err)
	assert.NotNil(t, d2)

	dumps, err = UseCaseDump1_ListLast(tmpDir, expextedDevice, expectedSalt, 5)
	assert.NoError(t, err)
	assert.NotNil(t, dumps)
	require.Len(t, dumps, 2)
	txt2, err := project(&dumps[0].Bucket)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxt2, txt2)
	txt1, err = project(&dumps[1].Bucket)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxt1, txt1)

	// Add third dump
	d3, err := UseCaseDump0_Create(tmpDir, expectedSalt, expextedDevice, expectedTxt3)
	assert.NoError(t, err)
	assert.NotNil(t, d3)

	dumps, err = UseCaseDump1_ListLast(tmpDir, expextedDevice, expectedSalt, 5)
	assert.NoError(t, err)
	assert.NotNil(t, dumps)
	require.Len(t, dumps, 3)
	txt3, err := project(&dumps[0].Bucket)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxt3, txt3)
	txt2, err = project(&dumps[1].Bucket)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxt2, txt2)
	txt1, err = project(&dumps[2].Bucket)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxt1, txt1)

	dumps, err = UseCaseDump1_ListLast(tmpDir, expextedDevice, expectedSalt, 2)
	assert.NoError(t, err)
	assert.NotNil(t, dumps)
	require.Len(t, dumps, 2)
	txt3, err = project(&dumps[0].Bucket)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxt3, txt3)
	txt2, err = project(&dumps[1].Bucket)
	assert.NoError(t, err)
	assert.Equal(t, expectedTxt2, txt2)
}
