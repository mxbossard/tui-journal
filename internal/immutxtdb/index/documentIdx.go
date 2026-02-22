package index

import (
	"encoding/binary"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
)

const (
	docIdxStateSize = 8
	docIdxKeySize   = 32
	docIdxDataSize  = 200
	docIdxPageSize  = 10
	docIdxQualifier = "doc"
)

var (
	docEncoderEuid = idx.Euid(binary.BigEndian.Uint64([]byte("doc00000")))
)

type DocIndex idx.Index[[]byte, *model.DocumentRef]

// (KEY: BUCKET_UID, STATE, VAL: DocumentRef)
func NewDocumentIndex(indexDir, device string) (DocIndex, error) {
	ser := serialize.ByteSliceSerializer{}
	enc := NewDocumentRefEncoder(0, docIdxStateSize, docIdxKeySize, docIdxDataSize)
	return idx.NewBasicIndex(indexDir, docIdxQualifier, device, ser, enc, docIdxPageSize)
}

func NewDocumentRefEncoder(version int32, stateSize, keySize, valSize int) idx.IdxEncoder[*model.DocumentRef] {
	return idx.NewAbstractEncoder(docEncoderEuid, version, stateSize, keySize, valSize, gobSerializer[model.DocumentRef]{})
}
