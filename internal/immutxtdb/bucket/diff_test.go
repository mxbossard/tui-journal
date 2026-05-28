package bucket

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/zip"
	"github.com/mxbossard/utilz/ztring"
	"github.com/stretchr/testify/assert"
)

func TestDiff_TextDiffPatch_Add(t *testing.T) {
	allWords := ztring.MdLoremIpsumSplitedWords()
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
	allWords := ztring.MdLoremIpsumSplitedWords()
	count := 50
	//count := len(allWords)
	firstTxt := strings.Join(allWords[:count], " ")
	appendedTxt := strings.Join(allWords[:count/2], " ")

	patch, err := TextDiffPatch(firstTxt, appendedTxt)
	assert.NoError(t, err)
	assert.Less(t, len(patch), len(appendedTxt), "patch larger than full text")
	fmt.Printf("Patch stats firstTxt: %d ; appendedTxt: %d ; patch: %d\n", len(firstTxt), len(appendedTxt), len(patch))
}

func TestDiff_TextDiffPatch_Then_PatchText(t *testing.T) {
	allWords := ztring.MdLoremIpsumSplitedWords()
	count := 50
	//count := len(allWords)
	firstTxt := strings.Join(allWords[:count/2], " ")
	appendedTxt := strings.Join(allWords[:count], " ")

	patch, err := TextDiffPatch(firstTxt, appendedTxt)
	assert.NoError(t, err)
	assert.Less(t, len(patch), len(appendedTxt), "patch larger than full text")

	mergedText, err := PatchText(firstTxt, patch)
	assert.NoError(t, err)
	assert.Equal(t, appendedTxt, mergedText)
}

func TestDiff_TextZipedPatch(t *testing.T) {
	allWords := ztring.MdLoremIpsumSplitedWords()
	count := 50
	//count := len(allWords)
	firstTxt := strings.Join(allWords[:count/2], " ")
	appendedTxt := strings.Join(allWords[:count], " ")

	patch, err := TextDiffPatch(firstTxt, appendedTxt)
	assert.NoError(t, err)

	zipedFullText, err := zip.ZlibCompressText(appendedTxt)
	assert.NoError(t, err)
	zipedPatch, err := zip.ZlibCompressText(patch)
	assert.NoError(t, err)
	fmt.Printf("Ziped patch stats fullTxt: %d ; zipedFullTxt: %d ; zipedPatch: %d\n", len(appendedTxt), len(zipedFullText), len(zipedPatch))
}
