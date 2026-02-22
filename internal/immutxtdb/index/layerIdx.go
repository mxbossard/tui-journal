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
	layerIdxStateSize = 8
	layerIdxKeySize   = 32
	layerIdxDataSize  = 200
	layerIdxPageSize  = 10
	layerIdxQualifier = "layer"
)

var (
	layerEncoderEuid = idx.Euid(binary.BigEndian.Uint64([]byte("layer000")))
)

type LayerIndex idx.Index[[]byte, *model.LayerRef]

// (KEY: BUCKET_UID, STATE, VAL: LayerRef)
func NewLayerIndex(indexDir, device string) (LayerIndex, error) {
	ser := serialize.ByteSliceSerializer{}
	enc := NewLayerRefEncoder(0, layerIdxStateSize, layerIdxKeySize, layerIdxDataSize)
	return idx.NewBasicIndex(indexDir, layerIdxQualifier, device, ser, enc, layerIdxPageSize)
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
