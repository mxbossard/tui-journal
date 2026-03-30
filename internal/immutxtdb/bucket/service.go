package bucket

import (
	"crypto/rand"
	"errors"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/files"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
	"github.com/mxbossard/utilz/filez"
)

const (
	BucketNameIdxStateSize = 8
	BucketNameIdxKeySize   = 64
	BucketNameIdxDataSize  = 16
	BucketNameIdxPageSize  = 100
	BucketNameIdxQualifier = "bucketName"

	HeaderRefIdxStateSize = 8
	HeaderRefIdxKeySize   = 32
	HeaderRefIdxDataSize  = 1000 //FIXME
	HeaderRefIdxPageSize  = 100
	HeaderRefIdxQualifier = "headerRef"

	BucketRefIdxStateSize = 8
	BucketRefIdxKeySize   = 32
	BucketRefIdxDataSize  = 1000 //FIXME
	BucketRefIdxPageSize  = 100
	BucketRefIdxQualifier = "bucketRef"

	zlibCompressionMark = byte(1)
)

var (
	ErrNotExist = errors.New("bucket do not exists")
)

var (
	dummyState = idx.BuildState(BucketNameIdxStateSize, "dummy")
)

var (
	headerSerializer   = serialize.StructSerializer[Header]{}
	metadataSerializer = serialize.StructSerializer[Metadata]{}
)

type BucketNameIndex idx.Index[string, BucketUid]
type HeaderRefIndex idx.Index[BucketUid, *HeaderRef]
type BucketRefIndex idx.Index[HashedBucketUid, *BucketRef]

type Service interface {
	New(name string, labels Labels) *Bucket
	Names() (map[string]BucketUid, error)
	Get(uid BucketUid) (*Bucket, error)
	Filter(f idx.Filter, o idx.Order, pageSize, preloadCount int) idx.Paginer[BucketUid, *Bucket]

	save(b *Bucket) error
	squash(b *Bucket) error
	commit(b *Bucket) error
}

type bucketService struct {
	dir    string
	device string
	salt   string

	bucketNameIdx BucketNameIndex
	headerRefIdx  HeaderRefIndex
	bucketRefIdx  BucketRefIndex
}

func NewBucketService(dir, device, salt string) (*bucketService, error) {
	bucketIdxDir := filepath.Join(dir, "bucketIdx")
	bucketByTimeIdxDir := filepath.Join(dir, "bucketByTimeIdx")
	layerIdxDir := filepath.Join(dir, "layerIdx")

	err := os.MkdirAll(bucketIdxDir, 0700)
	if err != nil {
		return nil, err
	}
	bucketNameIdx, err := newBucketNameIndex(bucketIdxDir, device)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(bucketByTimeIdxDir, 0700)
	if err != nil {
		return nil, err
	}
	headerRefIdx, err := newHeaderRefIndex(bucketByTimeIdxDir, device, salt)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(layerIdxDir, 0700)
	if err != nil {
		return nil, err
	}
	bucketRefIdx, err := newBucketRefIndex(layerIdxDir, device)
	if err != nil {
		return nil, err
	}

	return &bucketService{
		dir:    dir,
		device: device,
		salt:   salt,

		bucketNameIdx: bucketNameIdx,
		headerRefIdx:  headerRefIdx,
		bucketRefIdx:  bucketRefIdx,
	}, nil
}

func (s *bucketService) New(name string, labels Labels) *Bucket {
	uid := generateRandUid()
	// FIXME: check if uid already exists

	b := newBucket(s, uid, name)
	b.header.changed = true
	b.header.Labels = labels

	return b
}

func (s *bucketService) create(b *Bucket) error {
	now := time.Now()

	var data []byte
	switch b.header.Mode {
	case BinaryMode:
		data = b.data
	case TextMode:
		data = []byte(b.stringData)
	}
	if len(data) == 0 {
		return fmt.Errorf("no data to save")
	}

	// TODO: 1- Add (name, uid) in bucketName Idx
	_, err := s.bucketNameIdx.Add(dummyState, now, b.header.Name, b.header.Uid)
	if err != nil {
		return fmt.Errorf("unable to add bucket name: %w", err)
	}

	// TODO: 3- store Header, Layer & Metadata
	b.header.Created = &now
	headerRef, err := storeHeader(s.dir, s.device, &b.header)
	if err != nil {
		b.header.Created = nil
		return err
	}
	layerRef, err := storeLayer(s.dir, s.device, data)
	if err != nil {
		return err
	}

	metadata := Metadata{
		Version: 0,
		Size:    len(data),
		Updated: &now,
	}
	metadataRef, err := storeMetadata(s.dir, s.device, &metadata)
	if err != nil {
		return err
	}

	// TODO; 4- build Refs
	bucketRef := BucketRef{
		MetadataRef: *metadataRef,
		LayerRef:    *layerRef,
	}

	// TODO: 5- Add (RhUid, headerRef) in headerRef Idx
	entry, err := s.headerRefIdx.Add(dummyState, now, b.header.Uid, headerRef)
	if err != nil {
		return fmt.Errorf("unable to add header ref: %w", err)
	}
	hashedUid := HashedBucketUid(entry.BytesKey())

	// TODO: 6- Add (HUid, bucketRef) in bucketRef Idx
	_, err = s.bucketRefIdx.Add(dummyState, now, hashedUid, &bucketRef)
	if err != nil {
		return fmt.Errorf("unable to add bucket ref: %w", err)
	}

	b.header.Created = &now
	b.header.changed = false
	b.metadata = &metadata
	// var layerRefIt iter.Seq2[error, idx.Entry[HashedBucketUid, *LayerRef]] = func(yield func(error, idx.Entry[HashedBucketUid, *LayerRef]) bool) {
	// 	entry := idx.NewEntry(hashedUid, layerRef, -1, now, dummyState, nil, entry.BytesKey())
	// 	yield(nil, entry)
	// }
	var layerIt iter.Seq2[error, *Layer] = func(yield func(error, *Layer) bool) {
		l := Layer{
			Metadata: &metadata,
			Content:  data,
		}
		yield(nil, &l)
	}
	b.layerIt = layerIt
	return nil
}

func (s *bucketService) update(b *Bucket) error {
	now := time.Now()

	// TODO: 1- Attempt to make a patch of the update.
	var rootLayer, diffLayer bool
	var data []byte
	switch b.header.Mode {
	case BinaryMode:
		data = b.data
		// TODO: check if data changed
		// TODO: store a bynary layer ?
		// FIXME: add a binary append mode ?
		// FIXME: autodetect appending ?
		panic("not implemented yet")
	case TextMode:
		var err error
		stored, err := b.projectText()
		if err != nil {
			return err
		}
		zipedPatch, err := zipedPatch(stored, b.stringData)
		if err != nil {
			return err
		}
		if len(zipedPatch) == 0 {
			// No diff => nothing to save
			return fmt.Errorf("no data to save")
		}
		zipedFullText, err := ZlibCompressText(b.stringData)
		if err != nil {
			return err
		}

		if len(zipedPatch) > len(zipedFullText) {
			// Diff bigger than raw data => store a root layer
			rootLayer = true
			data = zipedFullText
			// TODO
			panic("not implemented yet")
		} else {
			// Store a diff layer
			diffLayer = true
			data = zipedPatch
			// TODO
			panic("not implemented yet")
		}
	}

	newMetadata := &Metadata{
		Version: b.metadata.Version + 1,
		Size:    len(data),
		Updated: &now,
	}

	newLayer := &Layer{
		Metadata: newMetadata,
		Content:  data,
	}

	// TODO: 2- Add (RhUid, headerRef) in headerRef Idx if Header changed
	if b.header.changed {
		// TODO
		panic("not implemented yet")
	}

	// TODO: 3- Create bucket layer + layerRef + metadataRef
	if rootLayer {
		// TODO Index a root layer
		_ = newLayer
	} else if diffLayer {
		// TODO Index a diff layer
	}

	// TODO: 4- Add (HUid, bucketRef) in bucketRef Idx
	panic("not implemented yet")

	b.header.changed = false
	b.metadata = newMetadata
	return nil
}

func (s *bucketService) save(b *Bucket) error {
	if b.saved {
		return nil
	}

	// 1- Check if supplied bucket already exists
	if b.metadata == nil {
		return s.create(b)
	}
	return s.update(b)
}

func (s *bucketService) Get(uid BucketUid) (*Bucket, error) {
	// 1- Get last bucket header
	headerPgnr, err := s.headerRefIdx.HashedPaginate(uid, idx.BottomToTop)
	if err != nil {
		return nil, err
	}

	var lastHeader *Header
	var lastBucketHashedUid HashedBucketUid
	// FIXME: MUST use all HashedBucketUid !
	for _, entry := range headerPgnr.All() {
		lastHeader, err = loadHeader(entry.Val())
		if err != nil {
			return nil, err
		}
		lastBucketHashedUid = HashedBucketUid(entry.BytesKey())
		// Keep only last bucket header found
		break
	}

	if lastHeader == nil {
		return nil, ErrNotExist
	}

	// 2- List all bucketRefs
	// FIXME: which uid key ?
	// FIXME: use state filter to stop on first root state ?
	bucketPgnr, err := s.bucketRefIdx.Paginate(lastBucketHashedUid, idx.BottomToTop)
	if err != nil {
		return nil, err
	}

	// Use all layers until first root layer
	var layers []*Layer
	var lastMetadata *Metadata

	for _, entry := range bucketPgnr.All() {
		metadata, err := loadMetadata(&entry.Val().MetadataRef)
		if err != nil {
			return nil, err
		}
		if lastMetadata == nil {
			lastMetadata = metadata
			break
		}
	}
	b := &Bucket{
		service:  s,
		header:   *lastHeader,
		metadata: lastMetadata,
		saved:    true,
	}

	b.layerIt = func(yield func(error, *Layer) bool) {
		bucketPgnr.Reset()
		for err, entry := range bucketPgnr.All() {
			if err != nil {
				if !yield(fmt.Errorf("error iterating bucketRefIdx: %w", err), nil) {
					break
				}
			}
			metadata, err := loadMetadata(&entry.Val().MetadataRef)
			if err != nil {
				if !yield(err, nil) {
					break
				}
			}

			data, err := files.LoadBlocPart((*filez.BlocPart)(&entry.Val().LayerRef))
			if err != nil {
				if !yield(err, nil) {
					break
				}
			}
			l := Layer{
				Metadata: metadata,
				Content:  data,
			}
			layers = append(layers, &l)

			if !yield(nil, &l) {
				break
			}
		}

	}
	return b, nil
}

func (s *bucketService) Filter(f idx.Filter, o idx.Order, pageSize, preloadCount int) idx.Paginer[BucketUid, *Bucket] {
	panic("not implemented yet")
}

func (s *bucketService) commit(b *Bucket) error {
	panic("not implemented yet")
}

func (s *bucketService) squash(b *Bucket) error {
	panic("not implemented yet")
}

func (s *bucketService) Names() (map[string]BucketUid, error) {
	nameIt, err := s.bucketNameIdx.All(idx.TopToBottom)
	if err != nil {
		return nil, err
	}

	nameByUidMap := make(map[BucketUid]string)
	lastUidByNameMap := make(map[string]BucketUid)
	for err, entry := range nameIt {
		if err != nil {
			return nil, err
		}
		nameByUidMap[BucketUid(entry.Val())] = entry.Key()
		lastUidByNameMap[entry.Key()] = entry.Val()
	}

	return lastUidByNameMap, nil
}

func storeHeader(dir, device string, h *Header) (*HeaderRef, error) {
	serializer := serialize.StructSerializer[Header]{}
	data := make([]byte, 1000) // FIXME: what is the good size ?
	n, err := serializer.Serialize(h, data)
	if err != nil {
		return nil, err
	}

	virtualBloc, err := files.StoreBlocData(dir, device, "bucketHeader", data[:n])
	if err != nil {
		return nil, err
	}

	if len(virtualBloc.Parts()) > 1 {
		panic("does not support multi part writes")
	}

	ref := HeaderRef(virtualBloc.Parts()[0])
	return &ref, nil
}

func storeMetadata(dir, device string, m *Metadata) (*MetadataRef, error) {
	serializer := serialize.StructSerializer[Metadata]{}
	data := make([]byte, 1000) // FIXME: what is the good size ?
	n, err := serializer.Serialize(m, data)
	if err != nil {
		return nil, err
	}

	virtualBloc, err := files.StoreBlocData(dir, device, "layerMetadata", data[:n])
	if err != nil {
		return nil, err
	}

	if len(virtualBloc.Parts()) > 1 {
		panic("does not support multi part writes")
	}

	ref := MetadataRef(virtualBloc.Parts()[0])
	return &ref, nil
}

func storeLayer(dir, device string, data []byte) (*LayerRef, error) {
	virtualBloc, err := files.StoreBlocData(dir, device, "layerData", data)
	if err != nil {
		return nil, err
	}

	if len(virtualBloc.Parts()) > 1 {
		panic("does not support multi part writes")
	}

	ref := LayerRef(virtualBloc.Parts()[0])
	return &ref, nil

}

func loadHeader(ref *HeaderRef) (*Header, error) {
	b, err := files.LoadBlocPart((*filez.BlocPart)(ref))
	if err != nil {
		return nil, err
	}
	return headerSerializer.Deserialize(b)
}

func loadMetadata(ref *MetadataRef) (*Metadata, error) {
	b, err := files.LoadBlocPart((*filez.BlocPart)(ref))
	if err != nil {
		return nil, err
	}
	return metadataSerializer.Deserialize(b)
}

func loadLayer(ref *LayerRef) ([]byte, error) {
	b, err := files.LoadBlocPart((*filez.BlocPart)(ref))
	return b, err
}

// (KEY: string, VAL: BucketUid)
func newBucketNameIndex(indexDir, device string) (BucketNameIndex, error) {
	enc := idx.NewAsciiEncoder(0, BucketNameIdxStateSize, BucketNameIdxKeySize, BucketNameIdxDataSize)
	keySer := serialize.AsciiSerializer{}
	valSer := BucketUidSerializer{}
	return idx.NewBasicIndex(indexDir, BucketNameIdxQualifier, device, keySer, valSer,
		nil, nil, enc, BucketNameIdxPageSize)
}

// (KEY: RH(BucketUid), VAL: HeaderRef)
func newHeaderRefIndex(indexDir, device, salt string) (HeaderRefIndex, error) {
	enc := idx.NewAsciiEncoder(0, HeaderRefIdxStateSize, HeaderRefIdxKeySize, HeaderRefIdxDataSize)
	keySer := BucketUidSerializer{}
	valSer := serialize.StructSerializer[HeaderRef]{}
	return idx.NewBasicIndex(indexDir, HeaderRefIdxQualifier, device, keySer, valSer,
		idx.NewRotatingHasher([]byte(salt), HeaderRefIdxKeySize), nil, enc, HeaderRefIdxPageSize)
}

// (KEY: H(BucketUid), VAL: BucketRef)
func newBucketRefIndex(indexDir, device string) (BucketRefIndex, error) {
	enc := idx.NewAsciiEncoder(0, BucketRefIdxStateSize, BucketRefIdxKeySize, BucketRefIdxDataSize)
	keySer := HashedBucketUidSerializer{}
	valSer := serialize.StructSerializer[BucketRef]{}
	return idx.NewBasicIndex(indexDir, BucketRefIdxQualifier, device, keySer, valSer,
		nil, nil, enc, BucketRefIdxPageSize)
}

func generateRandUid() BucketUid {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		panic(err)
	}
	return BucketUid(randBytes)
}

func zipedPatch(txt1, txt2 string) ([]byte, error) {
	patch, err := TextDiffPatch(txt1, txt2)
	if err != nil {
		return nil, err
	}
	ziped, err := ZlibCompressText(patch)
	if err != nil {
		return nil, err
	}
	return append([]byte{zlibCompressionMark}, ziped...), nil
}
