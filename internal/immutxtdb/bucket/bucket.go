package bucket

import (
	"iter"
	"sync"
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
	*sync.Mutex
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
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	panic("not implemented yet")
}

func (b *Bucket) Write(content []byte) error {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	b.data = content
	b.saved = false
	return nil
}

func (b *Bucket) WriteString(content string) error {
	return b.Write([]byte(content))
}

func (b *Bucket) Labels(labels Labels) error {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	panic("not implemented yet")
}

func (b *Bucket) Save() error {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	return b.service.save(b)
}

func (b *Bucket) Commit() error {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	return b.service.commit(b)
}

func (b *Bucket) Squash() error {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	return b.service.squash(b)
}
