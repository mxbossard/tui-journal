package zip

import (
	"bytes"
	"compress/zlib"
	"io"
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
