package index

import (
	"encoding/binary"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
)

const (
	textIdxStateSize = 8
	textIdxKeySize   = 32
	textIdxDataSize  = 200
	textIdxPageSize  = 10
)

var (
	textEncoderEuid = idx.Euid(binary.BigEndian.Uint64([]byte("txt00000")))
)

type TextIndex idx.Index[*[128]byte, *model.TextRef]

// (KEY: BUCKET_UID, STATE, VAL: TextRef)
func NewTextIndex(indexDir, device string) (TextIndex, error) {
	ser := serialize.ByteArray128Serializer{}
	enc := NewTextRefRefEncoder(0, textIdxStateSize, textIdxKeySize, textIdxDataSize)
	return idx.NewBasicIndex(indexDir, "text", device, ser, enc, textIdxPageSize)
}

func NewTextRefRefEncoder(version int32, stateSize, keySize, valSize int) idx.IdxEncoder[*model.TextRef] {
	return idx.NewAbstractEncoder(textEncoderEuid, version, stateSize, keySize, valSize, gobSerializer[model.TextRef]{})
}
