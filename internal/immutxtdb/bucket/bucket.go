package bucket

import (
	"iter"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/files"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
)

type BucketUid [16]byte
type Labels map[string]string

// Store Layer Metadata
type Metadata struct {
	Version int
	Updated time.Time
	Size    int
}

type Layer struct {
	Metadata   Metadata
	Content    []byte
	Commited   bool
	Snapshoted bool
}

// Store Bucket metadata which rarely changes
type Header struct {
	Uid     BucketUid
	Name    string
	Created time.Time
	Labels  Labels

	changed bool
}

// Match SHA256 ?
type HashedBucketUid [32]byte

type MetadataRef files.BlocRefPart
type LayerRef files.BlocRefPart
type HeaderRef files.BlocRefPart

type BucketRef struct {
	metadataRef MetadataRef
	layerRef    LayerRef
}

type Bucket struct {
	service *bucketService

	header Header
	// Last layer Metadata
	metadata   *Metadata
	layerRefIt iter.Seq2[error, idx.Entry[HashedBucketUid, *LayerRef]]
	//layers     []*Layer

	data  []byte
	saved bool
}

func (b Bucket) Project() (txt string, err error) {
	panic("not implemented yet")
}

func (b Bucket) Write(content []byte, labels Labels) error {
	b.saved = false
	panic("not implemented yet")
}

func (b Bucket) WriteString(content string, labels Labels) error {
	b.saved = false
	panic("not implemented yet")
}

func (b Bucket) Save() error {
	return b.service.save(&b)
}

func (b Bucket) Commit() error {
	return b.service.commit(&b)
}

func (b Bucket) Squash() error {
	return b.service.squash(&b)
}
