package bucket

import (
	"fmt"
	"iter"
	"sync"
	"time"

	"github.com/mxbossard/utilz/filez"
)

const (
	BinaryMode = int8(1)
	TextMode   = int8(2)
)

type BucketUid [16]byte
type Labels map[string]string

// Store Layer Metadata
type Metadata struct {
	Version int
	Updated *time.Time
	Size    int
}

type Layer struct {
	Metadata   *Metadata
	Content    []byte
	Commited   bool
	Snapshoted bool
}

// Store Bucket metadata which rarely changes
type Header struct {
	Uid     BucketUid
	Name    string
	Created *time.Time
	Labels  Labels
	Mode    int8

	changed bool
}

// Match SHA256 ?
type HashedBucketUid [32]byte

type MetadataRef filez.BlocPart
type LayerRef filez.BlocPart
type HeaderRef filez.BlocPart

type BucketRef struct {
	MetadataRef MetadataRef
	LayerRef    LayerRef
}

type Bucket struct {
	*sync.Mutex
	service Service

	header Header
	// Last layer Metadata
	metadata *Metadata
	layerIt  iter.Seq2[error, *Layer]
	//layers     []*Layer

	data       []byte
	stringData string
	saved      bool
}

func newBucket(s Service, uid BucketUid, name string) *Bucket {
	b := Bucket{
		Mutex:   &sync.Mutex{},
		service: s,
		header: Header{
			Uid:  uid,
			Name: name,
		},
	}

	return &b
}

func (b Bucket) Project() (txt string, err error) {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()

	return b.project()
}

func (b Bucket) project() (txt string, err error) {
	panic("not implemented yet")
}

func (b *Bucket) Write(data []byte) error {
	if b.header.Mode == TextMode {
		return fmt.Errorf("use WriteText for text mode bucket")
	}

	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	if b.header.Mode == 0 {
		b.header.Mode = BinaryMode
	}
	b.data = data
	b.saved = false
	return nil
}

func (b *Bucket) WriteText(text string) error {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	if b.header.Mode != 0 && b.header.Mode != TextMode {
		return fmt.Errorf("not a text mode bucket")
	}
	if b.header.Mode == 0 {
		b.header.Mode = TextMode
	}
	b.stringData = text
	b.saved = false
	return nil
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
