package zip

import (
	"fmt"
	"testing"

	"github.com/mxbossard/utilz/ztring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiff_ZlibCompressText_Small(t *testing.T) {
	expectedTxt := ztring.LoremIpsumWords(20)
	uncompressedTxt := []byte(expectedTxt)
	compressedTxt, err := ZlibCompressText(expectedTxt)
	assert.NoError(t, err)
	require.NotEmpty(t, compressedTxt)
	assert.Less(t, len(compressedTxt), len(uncompressedTxt), "text not compressed")
	fmt.Printf("ZlibCompressText small stats txt: %d ; txtBytes: %d ; ziped: %d\n", len(expectedTxt), len(uncompressedTxt), len(compressedTxt))
}

func TestDiff_ZlibCompressText_Big(t *testing.T) {
	expectedTxt := ztring.LoremIpsumWords(1000)
	uncompressedTxt := []byte(expectedTxt)
	compressedTxt, err := ZlibCompressText(expectedTxt)
	assert.NoError(t, err)
	require.NotEmpty(t, compressedTxt)
	assert.Less(t, len(compressedTxt), len(uncompressedTxt)/3, "text not compressed")
	fmt.Printf("ZlibCompressText big stats txt: %d ; txtBytes: %d ; ziped: %d\n", len(expectedTxt), len(uncompressedTxt), len(compressedTxt))
}

func TestDiff_ZlibCompressText_ZlibDecompressText(t *testing.T) {
	expectedTxt := ztring.LoremIpsumWords(100)
	compressedTxt, err := ZlibCompressText(expectedTxt)
	assert.NoError(t, err)
	require.NotEmpty(t, compressedTxt)
	decompressedTxt, err := ZlibDecompressText(compressedTxt)
	assert.NoError(t, err)
	require.NotEmpty(t, decompressedTxt)
	assert.Equal(t, expectedTxt, decompressedTxt, "decompressed text not identical")
}
