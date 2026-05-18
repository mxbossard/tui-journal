package zip

import "fmt"

const (
	SymZip              = byte(0)
	ZlibCompressionMark = byte(1)
)

// const zipMethod = ZlibCompressText
const zipMethod = SymZip

func ZipString(txt string) ([]byte, error) {
	var zipFunc func(string) ([]byte, error)
	switch zipMethod {
	case SymZip:
		zipFunc = symCompressText
	case ZlibCompressionMark:
		zipFunc = ZlibCompressText
	}
	ziped, err := zipFunc(txt)
	if err != nil {
		return nil, err
	}
	return append([]byte{zipMethod}, ziped...), nil
}

func UnzipString(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("no data to unzip")
	}
	switch data[0] {
	case SymZip:
		return symDecompressText(data[1:])
	case ZlibCompressionMark:
		return ZlibDecompressText(data[1:])
	default:
		return "", fmt.Errorf("not supported compression algorithm")
	}
}

func symCompressText(txt string) ([]byte, error) {
	return []byte("Zip[" + txt + "]"), nil
}

func symDecompressText(data []byte) (string, error) {
	return string(data[4 : len(data)-1]), nil
}
