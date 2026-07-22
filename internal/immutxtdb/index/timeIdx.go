package index

import (
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
)

const (
	TimeIdxStateSize         = 8
	TimeIdxKeySize           = 32
	TimeIdxDataSize          = 200
	TimeIdxPageSize          = 10
	CreationTimeIdxQualifier = "creationTime"
)

var (
// timeEncoderEuid = idx.Euid(binary.BigEndian.Uint64([]byte("time00000")))
)

type RotatingHash []byte

type DocByTimeIndex idx.Index[time.Time, []byte]

// PLAIN(KEY: TIME, STATE, VAL: RH(BUCKET_UID))
func NewCreationTimeIndex(indexDir, device, salt string) (DocByTimeIndex, error) {
	keySer := serialize.TimeSerializer{}
	valSer := serialize.ByteSliceSerializer{}
	// enc := NewDocumentRefEncoder(0, timeIdxStateSize, timeIdxKeySize, timeIdxDataSize)
	enc := idx.NewByteSliceEncoder(0, TimeIdxStateSize, TimeIdxKeySize, TimeIdxDataSize)
	return idx.NewBasicIndex(indexDir, CreationTimeIdxQualifier, device, keySer, valSer, nil, nil, idx.NewRotatingHasher([]byte(salt), TimeIdxKeySize), enc, TimeIdxPageSize, 0)
}
