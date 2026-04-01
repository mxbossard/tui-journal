package zip

import "fmt"

const (
	ZlibCompressionMark = byte(1)
)

func ZipString(txt string) ([]byte, error) {
	ziped, err := ZlibCompressText(txt)
	if err != nil {
		return nil, err
	}
	return append([]byte{ZlibCompressionMark}, ziped...), nil
}

func UnzipString(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("no data to unzip")
	}
	switch data[0] {
	case ZlibCompressionMark:
		return ZlibDecompressText(data[1:])
	default:
		return "", fmt.Errorf("not supported compression algorithm")
	}
}
