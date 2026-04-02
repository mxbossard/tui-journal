package bucket

import (
	"fmt"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/zip"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func TextDiffPatch(txt1, txt2 string) (string, error) {
	dmp := diffmatchpatch.New()
	diff := dmp.DiffMain(txt1, txt2, false)
	patch := dmp.PatchMake(diff)
	textPatch := dmp.PatchToText(patch)
	return textPatch, nil
}

func PatchText(txt string, patches ...string) (string, error) {
	dmp := diffmatchpatch.New()
	var ok []bool
	for _, patch := range patches {
		diffPatch, err := dmp.PatchFromText(patch)
		if err != nil {
			return "", err
		}

		txt, ok = dmp.PatchApply(diffPatch, txt)
		if !ok[0] {
			return "", fmt.Errorf("unable to apply patch: [%s]", patch)
		}
	}
	return txt, nil
}

func textPatchData(b *Bucket) ([]byte, error) {
	storedText, err := projectText(b, LatestVersion)
	if err != nil {
		return nil, err
	}
	nexText := b.stringData
	dmp := diffmatchpatch.New()
	diff := dmp.DiffMain(storedText, nexText, false)
	patch := dmp.PatchMake(diff)
	textPatch := dmp.PatchToText(patch)
	if len(textPatch) == 0 {
		return nil, err
	}
	zipedPatch, err := zip.ZlibCompressText(textPatch)
	return zipedPatch, err
}
