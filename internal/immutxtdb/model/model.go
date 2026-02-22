package model

import (
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
)

type BucketIdxEntry idx.BasicEntry[string, bool]
type BayerIdxEntry idx.BasicEntry[string, Layer]

type Labels map[string]string

type Metadata struct {
	Version    int
	Created    time.Time
	Updated    time.Time
	Labels     Labels
	Commited   bool
	Snapshoted bool
}

type Document struct {
	Content  string
	Metadata *Metadata
}

type BlocRef struct {
	BlocsFilepath string
	BlocId        int
}

type LayerRef struct {
	*BlocRef
	Pos int
	Len int
	// State         idx.State
}

type DocumentRef struct {
	BucketUid string
}

type TextRef struct {
	BucketUid string
	Pos       int
	Len       int
}

// func NewBlocRef(blocsFilepath string, blocId int, state idx.State) *LayerRef {
// 	return &LayerRef{
// 		BlocsFilepath: blocsFilepath,
// 		BlocId:        blocId,
// 		// State: state,
// 	}
// }

func NewLayerRef(blocsFilepath string, blocId int, state idx.State) *LayerRef {
	return &LayerRef{
		BlocRef: &BlocRef{
			BlocsFilepath: blocsFilepath,
			BlocId:        blocId,
		},
		// State: state,
	}
}

type Layer struct {
	content  string
	metadata *Metadata
}

type Bucket struct {
	//db     *DB
	uid    string
	layers []*LayerRef
	//layers *paginer[string, *layer]
}

func NewBucket(uid string, layers []*LayerRef) *Bucket {
	return &Bucket{
		uid:    uid,
		layers: layers,
	}
}

func (b Bucket) Project() (Document, error) {
	panic("not implemented yet")
}

func (b Bucket) Save(content string, labels Labels) error {
	panic("not implemented yet")
}

func (b Bucket) Commit() error {
	panic("not implemented yet")
}

func (b Bucket) Squash() error {
	panic("not implemented yet")
}

// type cursor[K comparable, V any] struct {
// 	pageSize int
// }

// // func (c cursor[K, V]) HasNext() bool {
// // 	panic("not implemented yet")
// // }

// // func (c *cursor[K, V]) Next() (page[K, V], error) {
// // 	panic("not implemented yet")
// // }

// func (c *cursor[K, V]) preLoad() error {
// 	panic("not implemented yet")
// }

// func (c *cursor[K, V]) All(errChan chan error) iter.Seq2[K, V] {
// 	panic("not implemented yet")
// }
