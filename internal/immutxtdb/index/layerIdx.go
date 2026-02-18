package index

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/encoder"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/mxbossard/utilz/filez"
)

// (RH(BUCKET_UID), RH(LAYER_FILE), BLOC_ID, STATE_PUBLIC_DATA)
type LayerIndex struct {
	model.Index[[]byte, *model.LayerRef]
	*sync.Mutex
	// FIXME: add a filelock

	encoder        encoder.Encoder[[]byte]
	filepathes     []string
	deviceIdxFiles []*filez.BlocsFile
	otherIdxFiles  []*filez.BlocsFile
	seqs           map[string]int
}

func NewLayerIndex(layerDir, device string) (*LayerIndex, error) {
	// Init bucketIndex
	// FIXME: manage multiple idx files (rotation)
	// FIXME: add a filelock
	// FIXME: addRotatingHash ?
	firstDeviceFilepath := filepath.Join(layerDir, fmt.Sprintf("layer-%s-001.idx", device))
	dbf1, err := filez.NewBlocsFile(firstDeviceFilepath, 256, 100)
	if err != nil {
		return nil, err
	}
	e := encoder.NewBytesEncoder(asciiEncoderDefaultVersion, asciiEncoderStateSize, asciiEncoderDataSize)
	idx := &LayerIndex{
		Mutex:          &sync.Mutex{},
		encoder:        e,
		deviceIdxFiles: []*filez.BlocsFile{dbf1},
		otherIdxFiles:  nil,
		seqs:           make(map[string]int),
	}

	// FIXME: need to setup the encoder!
	//e.Setup()

	return idx, nil
}

func (i *LayerIndex) Add(uidHash []byte, l *model.LayerRef) error {
	// Write to plain text file but private data is hashed
	panic("not implemented yet")
}

func (i *LayerIndex) Count() (int, error) {
	// FIXME: Is it needed ?
	panic("not implemented yet")
}

func (i *LayerIndex) Paginate(key string, order model.Order, limit int) (model.Paginer[[]byte, *model.LayerRef], chan error) {
	// TODO: build a cursor to iterate of the layers of a bucket
	panic("not implemented yet")
}

func (i *LayerIndex) PaginateAll(order model.Order, limit int) (model.Paginer[[]byte, *model.LayerRef], chan error) {
	// FIXME: Is it needed ?
	panic("not implemented yet")
}
