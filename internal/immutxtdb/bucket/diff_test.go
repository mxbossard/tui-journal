package bucket

import (
	"fmt"
	"strings"
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

func TestDiff_TextDiffPatch_Add(t *testing.T) {
	allWords := ztring.MdSplitedLoremIpsumWords()
	count := 50
	//count := len(allWords)
	firstTxt := strings.Join(allWords[:count/2], " ")
	appendedTxt := strings.Join(allWords[:count], " ")

	patch, err := TextDiffPatch(firstTxt, appendedTxt)
	assert.NoError(t, err)
	assert.Less(t, len(patch), len(appendedTxt), "patch larger than full text")
	fmt.Printf("Patch stats firstTxt: %d ; appendedTxt: %d ; patch: %d\n", len(firstTxt), len(appendedTxt), len(patch))
}

func TestDiff_TextDiffPatch_Remove(t *testing.T) {
	allWords := ztring.MdSplitedLoremIpsumWords()
	count := 50
	//count := len(allWords)
	firstTxt := strings.Join(allWords[:count], " ")
	appendedTxt := strings.Join(allWords[:count/2], " ")

	patch, err := TextDiffPatch(firstTxt, appendedTxt)
	assert.NoError(t, err)
	assert.Less(t, len(patch), len(appendedTxt), "patch larger than full text")
	fmt.Printf("Patch stats firstTxt: %d ; appendedTxt: %d ; patch: %d\n", len(firstTxt), len(appendedTxt), len(patch))
}

func TestDiff_TextZipedPatch(t *testing.T) {
	allWords := ztring.MdSplitedLoremIpsumWords()
	count := 50
	//count := len(allWords)
	firstTxt := strings.Join(allWords[:count/2], " ")
	appendedTxt := strings.Join(allWords[:count], " ")

	patch, err := TextDiffPatch(firstTxt, appendedTxt)
	assert.NoError(t, err)

	zipedFullText, err := ZlibCompressText(appendedTxt)
	assert.NoError(t, err)
	zipedPatch, err := ZlibCompressText(patch)
	assert.NoError(t, err)
	fmt.Printf("Ziped patch stats fullTxt: %d ; zipedFullTxt: %d ; zipedPatch: %d\n", len(appendedTxt), len(zipedFullText), len(zipedPatch))
}
