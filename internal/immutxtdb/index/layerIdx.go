package index

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
	"github.com/mxbossard/utilz/inoutz"
)

const (
	LayerIdxStateSize = 8
	LayerIdxKeySize   = 128
	LayerIdxDataSize  = 200
	LayerIdxPageSize  = 10
	LayerIdxQualifier = "layer"
)

var (
	layerEncoderEuid = idx.Euid(binary.BigEndian.Uint64([]byte("layer000")))
)

type HashedBucketUid [LayerIdxKeySize]byte

func stringToBucketUid(uid string) *HashedBucketUid {
	var a HashedBucketUid
	copy(a[:], []byte(uid))
	return &a
}

type BucketUidSerializer struct {
	serialize.Serializer[HashedBucketUid]
}

func (s BucketUidSerializer) Serialize(i *HashedBucketUid, o []byte) error {
	for k := range len(i) {
		o[k] = (*i)[k]
	}
	return nil
}

func (s BucketUidSerializer) Deserialize(i []byte) (*HashedBucketUid, error) {
	var o HashedBucketUid
	for k := 0; k < len(o) && k < len(i); k++ {
		(o)[k] = i[k]
	}
	return &o, nil
}

type LayerIndex idx.Index[*HashedBucketUid, *model.LayerRef]

// (KEY: BUCKET_UID, STATE, VAL: LayerRef)
func NewLayerIndex(indexDir, device string) (LayerIndex, error) {
	ser := BucketUidSerializer{}
	enc := NewLayerRefEncoder(0, LayerIdxStateSize, LayerIdxKeySize, LayerIdxDataSize)
	return idx.NewBasicIndex(indexDir, LayerIdxQualifier, device, ser, enc, LayerIdxPageSize)
}

type layerRefSerializer struct {
	serialize.Serializer[*model.LayerRef]
}

func (s layerRefSerializer) Serialize(i *model.LayerRef, o []byte) error {
	// FIXME: use rotating hash ?
	bw := inoutz.NewByteSliceWriter(o)
	var err error
	if i != nil {
		enc := gob.NewEncoder(bw)
		err = enc.Encode(*i)
	}
	return err

}

func (s layerRefSerializer) Deserialize(b []byte) (*model.LayerRef, error) {
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	var l model.LayerRef
	err := dec.Decode(&l)
	return &l, err
}

func NewLayerRefEncoder(version int32, stateSize, keySize, valSize int) idx.IdxEncoder[*model.LayerRef] {
	// return idx.NewAbstractEncoder(layerEncoderEuid, version, stateSize, keySize, valSize, layerRefSerializer{})
	return idx.NewAbstractEncoder(layerEncoderEuid, version, stateSize, keySize, valSize, gobSerializer[model.LayerRef]{})
}
