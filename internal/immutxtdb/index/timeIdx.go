package index

import (
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
)

const (
	timeIdxStateSize         = 8
	timeIdxKeySize           = 32
	timeIdxDataSize          = 200
	timeIdxPageSize          = 10
	creationTimeIdxQualifier = "creationTime"
)

var (
// timeEncoderEuid = idx.Euid(binary.BigEndian.Uint64([]byte("time00000")))
)

type RotatingHash []byte

type DocByTimeIndex idx.Index[time.Time, []byte]

// PLAIN(KEY: TIME, STATE, VAL: RH(BUCKET_UID))
func NewCreationTimeIndex(indexDir, device string) (DocByTimeIndex, error) {
	ser := serialize.TimeSerializer{}
	// enc := NewDocumentRefEncoder(0, timeIdxStateSize, timeIdxKeySize, timeIdxDataSize)
	enc := idx.NewByteSliceEncoder(0, timeIdxStateSize, timeIdxKeySize, timeIdxDataSize)
	return idx.NewBasicIndex(indexDir, creationTimeIdxQualifier, device, ser, enc, timeIdxPageSize)
}
