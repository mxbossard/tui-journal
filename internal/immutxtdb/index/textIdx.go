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
	keySer := serialize.ByteArray128Serializer{}
	valSer := gobSerializer[model.TextRef]{}
	enc := NewTextRefRefEncoder(0, textIdxStateSize, textIdxKeySize, textIdxDataSize)
	return idx.NewBasicIndex0[*[128]byte, *model.TextRef](indexDir, "text", device, keySer, valSer, nil, nil, nil, enc, textIdxPageSize, 0)
}

func NewTextRefRefEncoder(version int32, stateSize, keySize, valSize int) idx.IdxEncoder {
	return idx.NewBasicEncoder(textEncoderEuid, version, stateSize, keySize, valSize)
}
