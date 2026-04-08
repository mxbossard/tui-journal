package bucket

import (
	"fmt"
	"iter"
	"sync"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/zip"
	"github.com/mxbossard/utilz/filez"
)

const (
	BinaryMode = int8(1)
	TextMode   = int8(2)

	RootLayerFlag = 0x1
	DiffLayerFlag = 0x2

	FirstLayerVersion = Version(1)
	LatestVersion     = Version(0)
)

var (
	dummyState     = idx.BuildStringState(BucketNameIdxStateSize, "dummy")
	RootLayerState = idx.BuildState(RootLayerFlag)
	DiffLayerState = idx.BuildState(DiffLayerFlag)
)

type BucketUid [16]byte
type Version int

type Labels map[string]string

func NewLabels(labels ...string) Labels {
	l := make(Labels)
	for k := 0; k <= len(labels); k += 2 {
		key := labels[k]
		if len(labels) < k+2 {
			panic(fmt.Sprintf("missing label value for key: %s", key))
		}
		val := labels[k+1]
		l[key] = val
	}
	return l
}

// Store Layer Metadata
type Metadata struct {
	Version Version
	Updated *time.Time
	Size    int
}

type Layer struct {
	Metadata   *Metadata
	Content    []byte
	State      idx.State
	Commited   bool
	Snapshoted bool
}

// Store Bucket metadata which rarely changes
type Header struct {
	Uid      BucketUid
	Name     string
	Created  *time.Time
	Modified *time.Time // Header modification time
	Labels   Labels
	Mode     int8

	changed bool
}

type HashedBucketUid [32]byte // Match SHA256

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

	// lastHashedUid HashedBucketUid
	Header Header
	// Last layer Metadata
	Metadata *Metadata

	maxLoadedVersion    Version
	layerIt             iter.Seq2[error, *Layer]
	loadedBucketEntries map[Version]*idx.Entry[BucketUid, *BucketRef]
	loadedMetadatas     map[Version]*Metadata
	loadedLayers        map[Version]*Layer

	data       []byte
	stringData string
	saved      bool
}

func newBucket(s Service) *Bucket {
	b := Bucket{
		Mutex:               &sync.Mutex{},
		service:             s,
		loadedBucketEntries: make(map[Version]*idx.Entry[BucketUid, *BucketRef]),
		loadedMetadatas:     make(map[Version]*Metadata),
		loadedLayers:        make(map[Version]*Layer),
	}

	return &b
}

func newNamedBucket(s Service, uid BucketUid, name string) *Bucket {
	b := newBucket(s)
	b.Header = Header{
		Uid:  uid,
		Name: name,
	}
	return b
}

func (b *Bucket) LayerIt(version Version) (iter.Seq2[error, *Layer], error) {
	return b.service.buildLayerIt(b, version)
}

func (b *Bucket) ProjectBinary(version Version) (data []byte, err error) {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()

	return b.projectBinary(version)
}

func (b *Bucket) projectBinary(version Version) (data []byte, err error) {
	panic("not implemented yet")
}

func (b *Bucket) ProjectText(version Version) (txt string, err error) {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()

	return projectText(b, version)
}

func (b *Bucket) Write(data []byte) (int, error) {
	if b.Header.Mode == TextMode {
		return -1, fmt.Errorf("use WriteText for text mode bucket")
	}

	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	if b.Header.Mode == 0 {
		b.Header.Mode = BinaryMode
	}
	b.data = data
	b.saved = false
	return len(data), nil
}

func (b *Bucket) WriteText(text string) error {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	if b.Header.Mode != 0 && b.Header.Mode != TextMode {
		return fmt.Errorf("not a text mode bucket")
	}
	if b.Header.Mode == 0 {
		b.Header.Mode = TextMode
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

func projectText(b *Bucket, version Version) (string, error) {
	txt := ""
	k := 1
	layerIt, err := b.LayerIt(version)
	if err != nil {
		return "", err
	}
	for err, l := range layerIt {
		if err != nil {
			return "", fmt.Errorf("error iterating layer #%d: %w", k, err)
		}
		// First layer is root layer ?
		// TODO: is layer a root layer ?
		data, err := zip.UnzipString(l.Content)
		if err != nil {
			return "", fmt.Errorf("error decompressing layer #%d v%d: %w", k, l.Metadata.Version, err)
		}
		if idx.MatchFlag(l.State[0], RootLayerFlag) {
			txt = data
		} else if idx.MatchFlag(l.State[0], DiffLayerFlag) {
			txt, err = PatchText(txt, data)
			if err != nil {
				return "", fmt.Errorf("error patching layer #%d v%d: %w", k, l.Metadata.Version, err)
			}
		} else {
			panic("state not supported yet")
		}
		k++
	}
	return txt, nil
}
