package index

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	idxhash "github.com/mxbossard/tui-journal/internal/immutxtdb/idx/idxhash"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx/idxrepo"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
	"github.com/mxbossard/utilz/inoutz"
)

const (
	LayerIdxStateSize = 16
	LayerIdxKeySize   = 128
	LayerIdxDataSize  = 200
	LayerIdxPageSize  = 10
	LayerIdxQualifier = "layer"
)

var (
	layerEncoderEuid = idxrepo.Euid(binary.BigEndian.Uint64([]byte("layer000")))
)

func ByteSliceToBucketUid(uid []byte) model.HashedBucketUid {
	var a model.HashedBucketUid
	copy(a[:], uid)
	return a
}

func StringToBucketUid(uid string) model.HashedBucketUid {
	return ByteSliceToBucketUid([]byte(uid))
}

type BucketUidSerializer struct {
	serialize.Serializer[model.HashedBucketUid]
}

func (s BucketUidSerializer) Serialize(i *model.HashedBucketUid, o []byte) (int, error) {
	n := len(i)
	for k := range n {
		o[k] = (*i)[k]
	}
	return n, nil
}

func (s BucketUidSerializer) Deserialize(i []byte) (*model.HashedBucketUid, error) {
	var o model.HashedBucketUid
	for k := 0; k < len(o) && k < len(i); k++ {
		(o)[k] = i[k]
	}
	return &o, nil
}

type LayerIndex idx.Index[*model.HashedBucketUid, *model.LayerRef]

// (KEY: BUCKET_UID, STATE, VAL: LayerRef)
func NewLayerIndex(indexDir, device, salt string) (LayerIndex, error) {
	keySer := BucketUidSerializer{}
	// keySer := serialize.AsciiSerializer{}
	valSer := gobSerializer[model.LayerRef]{}
	enc := NewLayerRefEncoder(0, LayerIdxStateSize, LayerIdxKeySize, LayerIdxDataSize)
	return idx.NewBasicIndex0(indexDir, LayerIdxQualifier, device, keySer, valSer, []byte(salt), idxhash.NewRotatingHasher([]byte(salt), LayerIdxKeySize), nil, enc, LayerIdxPageSize, 0)
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

func NewLayerRefEncoder(version int32, stateSize, keySize, valSize int) idxrepo.IdxEncoder {
	// return idx.NewAbstractEncoder(layerEncoderEuid, version, stateSize, keySize, valSize, layerRefSerializer{})
	return idxrepo.NewBasicEncoder(layerEncoderEuid, version, stateSize, keySize, valSize)
}
