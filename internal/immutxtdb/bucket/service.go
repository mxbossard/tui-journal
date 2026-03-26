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
)

var (
	ErrNotExist = errors.New("bucket do not exists")
)

var (
	dummyState = idx.BuildState(BucketNameIdxStateSize, "dummy")
)

type BucketNameIndex idx.Index[string, BucketUid]
type HeaderRefIndex idx.Index[BucketUid, *HeaderRef]
type BucketRefIndex idx.Index[HashedBucketUid, *BucketRef]

type BucketUidSerializer struct {
	serialize.Serializer[BucketUid]
}

func (s BucketUidSerializer) Serialize(i BucketUid, o []byte) (int, error) {
	for k := range len(i) {
		o[k] = (i)[k]
	}
	return len(i), nil
}

func (s BucketUidSerializer) Deserialize(i []byte) (BucketUid, error) {
	var o BucketUid
	for k := 0; k < len(o) && k < len(i); k++ {
		o[k] = i[k]
	}
	return o, nil
}

type HashedBucketUidSerializer struct {
	serialize.Serializer[HashedBucketUid]
}

func (s HashedBucketUidSerializer) Serialize(i HashedBucketUid, o []byte) (int, error) {
	for k := range len(i) {
		o[k] = (i)[k]
	}
	return len(i), nil
}

func (s HashedBucketUidSerializer) Deserialize(i []byte) (HashedBucketUid, error) {
	var o HashedBucketUid
	for k := 0; k < len(o) && k < len(i); k++ {
		o[k] = i[k]
	}
	return o, nil
}

// (KEY: string, VAL: BucketUid)
func NewBucketNameIndex(indexDir, device string) (BucketNameIndex, error) {
	enc := idx.NewAsciiEncoder(0, BucketNameIdxStateSize, BucketNameIdxKeySize, BucketNameIdxDataSize)
	keySer := serialize.AsciiSerializer{}
	valSer := BucketUidSerializer{}
	return idx.NewBasicIndex(indexDir, BucketNameIdxQualifier, device, keySer, valSer,
		nil, nil, enc, BucketNameIdxPageSize)
}

// (KEY: RH(BucketUid), VAL: HeaderRef)
func NewHeaderRefIndex(indexDir, device, salt string) (HeaderRefIndex, error) {
	enc := idx.NewAsciiEncoder(0, HeaderRefIdxStateSize, HeaderRefIdxKeySize, HeaderRefIdxDataSize)
	keySer := BucketUidSerializer{}
	valSer := serialize.StructSerializer[HeaderRef]{}
	return idx.NewBasicIndex(indexDir, HeaderRefIdxQualifier, device, keySer, valSer,
		idx.NewRotatingHasher([]byte(salt), HeaderRefIdxKeySize), nil, enc, HeaderRefIdxPageSize)
}

// (KEY: H(BucketUid), VAL: BucketRef)
func NewBucketRefIndex(indexDir, device string) (BucketRefIndex, error) {
	enc := idx.NewAsciiEncoder(0, BucketRefIdxStateSize, BucketRefIdxKeySize, BucketRefIdxDataSize)
	keySer := HashedBucketUidSerializer{}
	valSer := serialize.StructSerializer[BucketRef]{}
	return idx.NewBasicIndex(indexDir, BucketRefIdxQualifier, device, keySer, valSer,
		nil, nil, enc, BucketRefIdxPageSize)
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
	bucketNameIdx, err := NewBucketNameIndex(bucketIdxDir, device)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(bucketByTimeIdxDir, 0700)
	if err != nil {
		return nil, err
	}
	headerRefIdx, err := NewHeaderRefIndex(bucketByTimeIdxDir, device, salt)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(layerIdxDir, 0700)
	if err != nil {
		return nil, err
	}
	bucketRefIdx, err := NewBucketRefIndex(layerIdxDir, device)
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

func generateRandUid() BucketUid {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		panic(err)
	}
	return BucketUid(randBytes)
}

func (s *bucketService) New(name string, labels Labels) *Bucket {
	uid := generateRandUid()
	// FIXME: check if uid already exists

	b := newBucket(s, uid, name)
	b.header.changed = true

	return b
}

func (s *bucketService) create(b *Bucket) error {
	now := time.Now()

	// TODO: 1- Add (name, uid) in bucketName Idx
	_, err := s.bucketNameIdx.Add(dummyState, now, b.header.Name, b.header.Uid)
	if err != nil {
		return fmt.Errorf("unable to add bucket name: %w", err)
	}

	// TODO: 2- Build first metadata
	metadata := Metadata{
		Version: 0,
	}

	// TODO: 3- store Header, Layer & Metadata
	headerRef, err := storeHeader(s.dir, s.device, &b.header)
	if err != nil {
		return err
	}
	metadataRef, err := storeMetadata(s.dir, s.device, &metadata)
	if err != nil {
		return err
	}
	layerRef, err := storeLayer(s.dir, s.device, b.data)
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

	b.header.Created = now
	b.header.changed = false
	b.metadata = &metadata
	// var layerRefIt iter.Seq2[error, idx.Entry[HashedBucketUid, *LayerRef]] = func(yield func(error, idx.Entry[HashedBucketUid, *LayerRef]) bool) {
	// 	entry := idx.NewEntry(hashedUid, layerRef, -1, now, dummyState, nil, entry.BytesKey())
	// 	yield(nil, entry)
	// }
	var layerIt iter.Seq2[error, *Layer] = func(yield func(error, *Layer) bool) {
		l := Layer{
			Metadata: &metadata,
			Content:  b.data,
		}
		yield(nil, &l)
	}
	b.layerIt = layerIt
	return nil
}

func (s *bucketService) update(b *Bucket) error {
	now := time.Now()

	// TODO: 2- Add (RhUid, headerRef) in headerRef Idx if Header changed
	// TODO: 3- Create bucket root layer + layerRef + metadataRef
	// TODO: 4- Add (HUid, bucketRef) in bucketRef Idx

	metadata := Metadata{
		Version: b.metadata.Version + 1,
	}
	b.header.Created = now
	b.header.changed = false
	b.metadata = &metadata
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

	b := newBucket(s, uid, lastHeader.Name)
	b.saved = true
	b.metadata = lastMetadata
	b.layerIt = func(yield func(error, *Layer) bool) {
		for _, entry := range bucketPgnr.All() {
			metadata, err := loadMetadata(&entry.Val().MetadataRef)
			if err != nil {
				if !yield(err, nil) {
					break
				}
			}
			if lastMetadata == nil {
				lastMetadata = metadata
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

func (s *bucketService) Filter(f idx.Filter, o idx.Order) idx.Paginer[BucketUid, *Bucket] {
	panic("not implemented yet")
}

func (s *bucketService) commit(b *Bucket) error {
	panic("not implemented yet")
}

func (s *bucketService) squash(b *Bucket) error {
	panic("not implemented yet")
}

func (s *bucketService) Names() (error, map[string]BucketUid) {
	panic("not implemented yet")
}

var (
	headerSerializer   = serialize.StructSerializer[Header]{}
	metadataSerializer = serialize.StructSerializer[Metadata]{}
)

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
