package index

import (
	"encoding/binary"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
)

const (
	docIdxStateSize = 8
	docIdxKeySize   = 128
	docIdxDataSize  = 200
	docIdxPageSize  = 10
	docIdxQualifier = "doc"
)

var (
	docEncoderEuid = idx.Euid(binary.BigEndian.Uint64([]byte("doc00000")))
)

type DocIndex idx.Index[*[docIdxKeySize]byte, *model.BucketRef]

// (KEY: BUCKET_UID, STATE, VAL: DocumentRef)
func NewDocumentIndex(indexDir, device string) (DocIndex, error) {
	keySer := serialize.ByteArray128Serializer{}
	valSer := gobSerializer[model.BucketRef]{}
	enc := NewDocumentRefEncoder(0, docIdxStateSize, docIdxKeySize, docIdxDataSize)
	return idx.NewBasicIndex0(indexDir, docIdxQualifier, device, keySer, valSer, nil, nil, nil, enc, docIdxPageSize, 0)
}

func NewDocumentRefEncoder(version int32, stateSize, keySize, valSize int) idx.IdxEncoder {
	return idx.NewBasicEncoder(docEncoderEuid, version, stateSize, keySize, valSize)
}
