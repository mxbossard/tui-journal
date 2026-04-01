package files

import (
	"os"
	"testing"

	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/ztring"
	"github.com/stretchr/testify/assert"
)

func TestBloc_Store_And_LoadPart(t *testing.T) {
	tmpDir := filez.MkdirTempOrPanic("TestBloc_Store_And_LoadPart")
	defer os.RemoveAll(tmpDir)

	expectedDevice := "device"
	expectedQualifier := "foo"
	excptectedData1 := []byte(ztring.LoremIpsumWords(10))
	excptectedData2 := []byte(ztring.LoremIpsumWords(5))
	excptectedData3 := []byte(ztring.LoremIpsumWords(15))

	vb1, err := StoreBlocData(tmpDir, expectedDevice, expectedQualifier, excptectedData1)
	assert.NoError(t, err)
	assert.NotNil(t, vb1)

	vb2, err := StoreBlocData(tmpDir, expectedDevice, expectedQualifier, excptectedData2)
	assert.NoError(t, err)
	assert.NotNil(t, vb2)

	vb3, err := StoreBlocData(tmpDir, expectedDevice, expectedQualifier, excptectedData3)
	assert.NoError(t, err)
	assert.NotNil(t, vb3)

	data1, err := LoadBlocPart(&vb1.Parts()[0])
	assert.NoError(t, err)
	assert.Equal(t, excptectedData1, data1)

	data2, err := LoadBlocPart(&vb2.Parts()[0])
	assert.NoError(t, err)
	assert.Equal(t, excptectedData2, data2)

	data3, err := LoadBlocPart(&vb3.Parts()[0])
	assert.NoError(t, err)
	assert.Equal(t, excptectedData3, data3)
}
