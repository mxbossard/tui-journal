package index

import (
	"bytes"
	"encoding/gob"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
	"github.com/mxbossard/utilz/inoutz"
)

const (
	asciiEncoderStateSize = 8
)

var (
	Document = idx.BuildState(asciiEncoderStateSize, "doc")
	Dump     = idx.BuildState(asciiEncoderStateSize, "dmp")
)

type gobSerializer[T any] struct {
	serialize.Serializer[T]
}

func (s gobSerializer[T]) Serialize(i *T, o []byte) error {
	// FIXME: use rotating hash ?
	bw := inoutz.NewByteSliceWriter(o)
	var err error
	if i != nil {
		enc := gob.NewEncoder(bw)
		err = enc.Encode(i)
	}
	return err

}

func (s gobSerializer[T]) Deserialize(b []byte) (*T, error) {
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	var l T
	err := dec.Decode(&l)
	return &l, err
}
