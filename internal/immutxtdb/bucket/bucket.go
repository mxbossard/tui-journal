package bucket

import (
	"crypto/rand"
	"errors"
	"fmt"
	"hash/fnv"
	"iter"
	"sort"
	"sync"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/zip"
	"github.com/mxbossard/utilz/collectionz"
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
	dummyState         = idx.BuildStringState(BucketNameIdxStateSize, "dummy")
	RootLayerState     = idx.BuildState(RootLayerFlag)
	DiffLayerState     = idx.BuildState(DiffLayerFlag)
	ErrVersionConflict = errors.New("bucket version conflict")
)

type BucketUid [16]byte
type Version int

type Labels map[string]string

func NewLabels(labels ...string) Labels {
	l := make(Labels)
	for k := 0; k < len(labels); k += 2 {
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
	Uid     BucketUid
	Version Version
	Updated *time.Time
	Size    int
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

type PartedVersion struct {
	Version Version
	Part    string
}

type Bucket struct {
	*sync.Mutex
	service Service

	// lastHashedUid HashedBucketUid
	Header Header
	// Last layer Metadata
	Metadata *Metadata

	lastBucketRefSeq    map[string]int
	maxLoadedVersion    Version
	layerIt             iter.Seq2[error, *Layer]
	loadedBucketEntries map[PartedVersion]*idx.Entry[BucketUid, *BucketRef]
	loadedMetadatas     map[PartedVersion]*Metadata
	loadedLayers        map[PartedVersion]*Layer

	data       []byte
	stringData string
	saved      bool
}

func newBucket(s Service) *Bucket {
	b := Bucket{
		Mutex:               &sync.Mutex{},
		service:             s,
		lastBucketRefSeq:    make(map[string]int),
		loadedBucketEntries: make(map[PartedVersion]*idx.Entry[BucketUid, *BucketRef]),
		loadedMetadatas:     make(map[PartedVersion]*Metadata),
		loadedLayers:        make(map[PartedVersion]*Layer),
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

func (b *Bucket) String() string {
	return fmt.Sprintf("Bucket[%s,#%v]", b.Header.Name, b.Header.Uid)
}

func (b *Bucket) LayerIt(versions ...Version) (iter.Seq2[error, *Layer], error) {
	var version Version
	if len(versions) == 0 {
		version = LatestVersion
	} else if len(versions) == 1 {
		version = versions[0]
	} else {
		panic("must supply 0 ore 1 version not more")
	}
	return b.service.buildLayerIt(b, version)
}

func (b *Bucket) ProjectBinary(versions ...Version) (data []byte, err error) {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()

	return b.projectBinary(versions...)
}

func (b *Bucket) projectBinary(versions ...Version) (data []byte, err error) {
	panic("not implemented yet")
}

func (b *Bucket) ProjectText(versions ...Version) (txt string, err error) {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()

	return projectText(b, versions...)
}

func (b *Bucket) SetBytes(data []byte) (int, error) {
	if b.Header.Mode == TextMode {
		return -1, fmt.Errorf("use UpdateText for text mode bucket")
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

func (b *Bucket) SetText(text string) error {
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

func projectText(b *Bucket, version ...Version) (string, error) {
	txt := ""
	layerIt, err := b.LayerIt(version...)
	if err != nil {
		return "", err
	}

	// Collect all patchesMap
	patchesMap := make(map[int][]string)
	k := 1
	for err, l := range layerIt {
		// fmt.Printf("projecting text layer #%d v%d\n", k, l.Metadata.Version)
		if err != nil {
			return "", fmt.Errorf("error iterating layer #%d: %w", k, err)
		}
		data, err := zip.UnzipString(l.Content)
		if err != nil {
			return "", fmt.Errorf("error decompressing layer #%d v%d: %w", k, l.Metadata.Version, err)
		}
		if idx.MatchFlag(l.State[0], RootLayerFlag) {
			// fmt.Printf("found root layer #%d v%d\n", k, l.Metadata.Version)
			txt = data
		} else if idx.MatchFlag(l.State[0], DiffLayerFlag) {
			// fmt.Printf("found diff layer #%d v%d\n", k, l.Metadata.Version)
			// Checking layer versions
			patchesMap[int(l.Metadata.Version)] = append(patchesMap[int(l.Metadata.Version)], data)
		} else {
			panic("state not supported yet")
		}
		k++
	}

	// Apply all patches
	versions := collectionz.Keys(patchesMap)
	// fmt.Printf("patchesMap: %v\n", patchesMap)
	sort.Ints(versions)
	// fmt.Printf("all patch versions: %v\n", versions)
	for _, v := range versions {
		patches := patchesMap[v]
		// fmt.Printf("layer v%d patches: %v\n", v, patches)
		if len(patches) > 1 {
			// 2 layers with same version => Confilct problem
			return txt, ErrVersionConflict
		}
		txt, err = PatchText(txt, patches[0])
		if err != nil {
			return "", fmt.Errorf("error patching layer #%d v%d: %w", k, v, err)
		}
	}
	return txt, nil
}

func RandUid() BucketUid {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		panic(err)
	}
	return BucketUid(randBytes)
}

func DeterministicUid(s string) BucketUid {
	h := fnv.New128a()
	_, err := h.Write([]byte(s))
	if err != nil {
		panic(err)
	}
	hBytes := h.Sum(nil)
	return BucketUid(hBytes)
}
