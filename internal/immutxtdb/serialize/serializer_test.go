package serialize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBinarySerializer(t *testing.T) {
	bs := BinarySerializer{}

	k1 := int8(42)
	out1 := make([]byte, 100)
	n1, err1 := bs.Serialize(k1, &out1)
	assert.NoError(t, err1)
	assert.Equal(t, 1, n1)
	assert.Equal(t, []byte{42}, out1[:n1])

	k2 := int32(42)
	out2 := make([]byte, 100)
	n2, err2 := bs.Serialize(k2, &out2)
	assert.NoError(t, err2)
	assert.Equal(t, 4, n2)
	assert.Equal(t, []byte{0, 0, 0, 42}, out2[:n2])

	var k3 [8]byte
	k3 = [8]byte{1, 0, 8}
	out3 := make([]byte, 100)
	n3, err3 := bs.Serialize(k3, &out3)
	assert.NoError(t, err3)
	assert.Equal(t, 8, n3)
	assert.Equal(t, []byte{1, 0, 8, 0, 0, 0, 0, 0}, out3[:n3])
}
