package model

import (
	"iter"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
)

type BucketIdxEntry idx.BasicEntry[string, bool]
type BayerIdxEntry idx.BasicEntry[string, Layer]

/*
- BucketRef -> []LayerRef -> []BlocRef
- Bucket -> []Layers => Document => [](MetaData + Data)
- List Buckets (= list documents) by name, time, topics
-
*/

type BucketUid string
type HashedBucketUid [128]byte

type Labels map[string]string

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

type BucketRef struct {
	Uid BucketUid
}

type TextRef struct {
	BucketUid BucketUid
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

type LayerMetadata struct {
	Version int
	Created time.Time
}

type Layer struct {
	Metadata   LayerMetadata
	Content    []byte
	Commited   bool
	Snapshoted bool
}

type BucketMetadata struct {
	Version int
	Created time.Time
	Updated time.Time
	Labels  Labels
}

type Bucket struct {
	Uid        BucketUid
	Metadata   BucketMetadata
	LayerRefIt iter.Seq[*LayerRef]
	//layers     []*Layer
}

func (b Bucket) Project() (txt string, err error) {
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
