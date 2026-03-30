package bucket

import (
	"bytes"
	"compress/zlib"
	"io"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func ZlibCompressText(text string) ([]byte, error) {
	var b bytes.Buffer
	w, err := zlib.NewWriterLevel(&b, zlib.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = w.Write([]byte(text))
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func ZlibDecompressText(data []byte) (string, error) {
	dataReader := bytes.NewReader(data)
	r, err := zlib.NewReader(dataReader)
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	_, err = io.Copy(&b, r)
	if err != nil {
		return "", err
	}
	err = r.Close()
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func TextDiffPatch(txt1, txt2 string) (string, error) {
	dmp := diffmatchpatch.New()
	diff := dmp.DiffMain(txt1, txt2, false)
	patch := dmp.PatchMake(diff)
	textPatch := dmp.PatchToText(patch)
	return textPatch, nil
}

func textPatchData(b *Bucket) ([]byte, error) {
	storedText, err := b.projectText()
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
	zipedPatch, err := ZlibCompressText(textPatch)
	return zipedPatch, err
}
