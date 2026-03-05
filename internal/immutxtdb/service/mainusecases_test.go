package service

import (
	"testing"
	"time"

	"github.com/mxbossard/utilz/ztring"
	"github.com/stretchr/testify/assert"
)

func TestUseCaseDump0_Create(t *testing.T) {
	now := time.Now()
	expextedDevice := "foo"
	expectedTxt := ztring.LoremIpsum()
	d, err := UseCaseDump0_Create(expextedDevice, expectedTxt)
	assert.NoError(t, err)
	assert.NotNil(t, d)

	assert.NotEmpty(t, d.Uid)
	assert.True(t, d.Metadata.Created.After(now))
	assert.Equal(t, d.Metadata.Created, d.Metadata.Updated)
	assert.NotNil(t, d.LayerRefIt)
	k := 0
	for l := range d.LayerRefIt {
		assert.NotNil(t, l)
	}
	assert.Equal(t, 1, k)

	txt, err := d.Project()
	assert.NoError(t, err)
	assert.Equal(t, expectedTxt, txt)

}
