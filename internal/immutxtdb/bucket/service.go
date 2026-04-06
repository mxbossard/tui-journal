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
	"github.com/mxbossard/tui-journal/internal/immutxtdb/zip"
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

	BucketRefIdxStateSize = 2
	BucketRefIdxKeySize   = 32
	BucketRefIdxDataSize  = 1000 //FIXME
	BucketRefIdxPageSize  = 100
	BucketRefIdxQualifier = "bucketRef"
)

var (
	ErrNotExist = errors.New("bucket do not exists")
)

var (
	headerSerializer   = serialize.StructSerializer[Header]{}
	metadataSerializer = serialize.StructSerializer[Metadata]{}
)

type BucketNameIndex idx.Index[string, BucketUid]
type HeaderRefIndex idx.Index[BucketUid, *HeaderRef]
type BucketRefIndex idx.Index[BucketUid, *BucketRef]

type Service interface {
	// Create a new Bucket
	New(name string, labels Labels) *Bucket
	// Return all Names associated with it's last Bucket Uid
	Names() (map[string]BucketUid, error)
	// Get a Bucket by it's Uid
	Get(uid BucketUid) (*Bucket, error)
	// Get a slice of Buckets
	Filter(o idx.Order, f idx.Filter, pageSize, preloadPageCount int) (idx.Paginer[BucketUid, *Bucket], error)
	// Save a Bucket
	Save(b *Bucket, partition string) error

	buildLayerIt(b *Bucket, v Version) (iter.Seq2[error, *Layer], error)
}

type bucketService struct {
	dir string
	// partition string
	salt string

	// bucketNameIdx BucketNameIndex
	// headerRefIdx  HeaderRefIndex
	// bucketRefIdx  BucketRefIndex
}

func NewBucketService(dir, salt string) (*bucketService, error) {
	return &bucketService{
		dir: dir,
		// partition: partition,
		salt: salt,

		// bucketNameIdx: bucketNameIdx,
		// headerRefIdx:  headerRefIdx,
		// bucketRefIdx:  bucketRefIdx,
	}, nil
}

func (s *bucketService) New(name string, labels Labels) *Bucket {
	uid := generateRandUid()
	// FIXME: check if uid already exists

	b := newNamedBucket(s, uid, name)
	b.Header.changed = true
	b.Header.Labels = labels

	return b
}

func (s *bucketService) NewText(name string, labels Labels, text string) (*Bucket, error) {
	b := s.New(name, labels)
	err := b.WriteText(text)
	return b, err
}

func (s *bucketService) create(b *Bucket, partition string) error {
	if b.Header.Created == nil {
		now := time.Now()
		b.Header.Created = &now
	}

	var data []byte
	var dataLen int
	switch b.Header.Mode {
	case BinaryMode:
		data = b.data
		dataLen = len(data)
	case TextMode:
		zipedFullText, err := zip.ZipString(b.stringData)
		if err != nil {
			return fmt.Errorf("update: unable to zip full text: %w", err)
		}
		data = zipedFullText
		dataLen = len(b.stringData)
	}
	if len(data) == 0 {
		return fmt.Errorf("create: no data to save")
	}

	bucketNameIdx, headerRefIdx, bucketRefIdx, err := getServiceIndexes(s.dir, s.salt, partition)
	if err != nil {
		return err
	}

	// TODO: 1- Add (name, uid) in bucketName Idx
	_, err = bucketNameIdx.Add(dummyState, *b.Header.Created, b.Header.Name, b.Header.Uid)
	if err != nil {
		return fmt.Errorf("create: unable to add bucket name: %w", err)
	}

	// TODO: 3- store Header, Layer & Metadata
	headerRef, err := storeHeader(s.dir, partition, &b.Header)
	if err != nil {
		b.Header.Created = nil
		return fmt.Errorf("create: unable to store header: %w", err)
	}

	layerRef, err := storeLayerData(s.dir, partition, data)
	if err != nil {
		return fmt.Errorf("create: unable to store layer: %w", err)
	}

	metadata := Metadata{
		Version: FirstLayerVersion,
		Size:    dataLen,
		Updated: b.Header.Created,
	}
	metadataRef, err := storeMetadata(s.dir, partition, &metadata)
	if err != nil {
		return fmt.Errorf("create: unable to store metadata: %w", err)
	}

	// TODO; 4- build Refs
	bucketRef := BucketRef{
		MetadataRef: *metadataRef,
		LayerRef:    *layerRef,
	}

	// TODO: 5- Add (RhUid, headerRef) in headerRef Idx
	hrEntry, err := headerRefIdx.Add(dummyState, *b.Header.Created, b.Header.Uid, headerRef)
	if err != nil {
		return fmt.Errorf("create: unable to add header ref: %w", err)
	}
	_ = hrEntry
	// hashedUid := HashedBucketUid(hrEntry.KeyBytes())

	// TODO: 6- Add (HUid, bucketRef) in bucketRef Idx
	brEntry, err := bucketRefIdx.Add(RootLayerState, *b.Header.Created, b.Header.Uid, &bucketRef)
	if err != nil {
		return fmt.Errorf("create: unable to add bucket ref: %w", err)
	}
	// fmt.Printf("Added bucketRef for uid: %v / hash: %v\n", b.header.Uid, hashedUid)

	l := Layer{
		Metadata: &metadata,
		Content:  data,
		State:    brEntry.State(),
	}

	b.Header.changed = false
	b.Metadata = &metadata
	// b.lastHashedUid = hashedUid
	b.maxLoadedVersion = metadata.Version
	b.loadedBucketEntries[metadata.Version] = &brEntry
	b.loadedMetadatas[metadata.Version] = &metadata
	b.loadedLayers[metadata.Version] = &l

	return nil
}

func (s *bucketService) update(b *Bucket, partition string) error {
	now := time.Now()

	// TODO: 1- Attempt to make a patch of the update.
	var rootLayer, diffLayer bool
	var data []byte
	var dataLen int
	switch b.Header.Mode {
	case BinaryMode:
		data = b.data
		dataLen = len(data)
		// TODO: check if data changed
		// TODO: store a bynary layer ?
		// FIXME: add a binary append mode ?
		// FIXME: autodetect appending ?
		panic("not implemented yet")
	case TextMode:
		var err error
		stored, err := projectText(b, LatestVersion)
		if err != nil {
			return fmt.Errorf("update: unable to project text: %w", err)
		}
		patch, err := TextDiffPatch(stored, b.stringData)
		if err != nil {
			return fmt.Errorf("update: unable to make patch: %w", err)
		}
		zipedPatch, err := zip.ZipString(patch)
		if err != nil {
			return fmt.Errorf("update: unable to zip patch: %w", err)
		}
		if len(zipedPatch) == 0 {
			// No diff => nothing to save
			return fmt.Errorf("update: no data to save")
		}
		zipedFullText, err := zip.ZipString(b.stringData)
		if err != nil {
			return fmt.Errorf("update: unable to zip full text: %w", err)
		}

		if len(zipedPatch) > len(zipedFullText) {
			// Diff bigger than raw data => store a root layer
			rootLayer = true
			data = zipedFullText
			// TODO
			// panic("not implemented yet")
		} else {
			// Store a diff layer
			diffLayer = true
			data = zipedPatch
			// TODO
			// panic("not implemented yet")
		}
		dataLen = len(b.stringData)
	}

	newMetadata := &Metadata{
		Version: b.Metadata.Version + 1,
		Size:    dataLen,
		Updated: &now,
	}

	// hashedUid := b.lastHashedUid
	_, headerRefIdx, bucketRefIdx, err := getServiceIndexes(s.dir, s.salt, partition)
	if err != nil {
		return err
	}

	// TODO: 2- Add (RhUid, headerRef) in headerRef Idx if Header changed
	if b.Header.changed {
		// TODO store header + headerRef
		headerRef, err := storeHeader(s.dir, partition, &b.Header)
		if err != nil {
			return fmt.Errorf("update: unable store header: %w", err)
		}
		entry, err := headerRefIdx.Add(dummyState, now, b.Header.Uid, headerRef)
		if err != nil {
			return fmt.Errorf("update: unable to add header ref: %w", err)
		}
		_ = entry
		// hashedUid = HashedBucketUid(entry.KeyBytes())
		panic("not implemented yet")
	}

	// TODO: Store Layer + Metadata
	var layerState idx.State
	if rootLayer {
		// TODO Index a root layer
		layerState = RootLayerState
	} else if diffLayer {
		// TODO Index a diff layer
		layerState = DiffLayerState
	}

	layerRef, err := storeLayerData(s.dir, partition, data)
	if err != nil {
		return fmt.Errorf("update: unable store layer: %w", err)
	}
	metadataRef, err := storeMetadata(s.dir, partition, newMetadata)
	if err != nil {
		return fmt.Errorf("update: unable store metadata: %w", err)
	}

	// TODO: 3- Create bucketRef (layerRef + metadataRef)
	bucketRef := BucketRef{
		MetadataRef: *metadataRef,
		LayerRef:    *layerRef,
	}

	// TODO: 4- Add (HUid, bucketRef) in bucketRef Idx
	brEntry, err := bucketRefIdx.Add(layerState, now, b.Header.Uid, &bucketRef)
	if err != nil {
		return fmt.Errorf("update: unable to add bucket ref: %w", err)
	}
	// fmt.Printf("Added bucketRef for uid: %v / hash: %v\n", b.header.Uid, hashedUid)

	l := Layer{
		Metadata: newMetadata,
		Content:  data,
		State:    brEntry.State(),
	}

	// panic("not implemented yet")
	b.maxLoadedVersion = newMetadata.Version
	b.loadedBucketEntries[newMetadata.Version] = &brEntry
	b.loadedMetadatas[newMetadata.Version] = newMetadata
	b.loadedLayers[newMetadata.Version] = &l
	b.Header.changed = false
	b.Metadata = newMetadata
	return nil
}

func (s *bucketService) Save(b *Bucket, partition string) error {
	if b.saved {
		return nil
	}

	// 1- Check if supplied bucket already exists
	if b.Metadata == nil {
		return s.create(b, partition)
	}
	return s.update(b, partition)
}

func getLastBucketHeader(paginer idx.Paginer[BucketUid, *HeaderRef]) (*Header, error) {
	var lastHeader *Header
	// FIXME: MUST use all HashedBucketUid !
	for entry := range paginer.All() {
		if entry.Error() != nil {
			return nil, entry.Error()
		}
		var err error
		lastHeader, err = loadHeader(entry.Val())
		if err != nil {
			return nil, err
		}
		// Keep only last bucket header found
		break
	}

	if lastHeader == nil {
		return nil, ErrNotExist
	}

	return lastHeader, nil
}

func (s *bucketService) buildLazyBucket(lastHeader *Header) (*Bucket, error) {
	// 2- List all bucketRefs
	// FIXME: which uid key ?
	// FIXME: use state filter to stop on first root state ?
	fmt.Printf("buildLazyBucket, bUid: %v\n", lastHeader.Uid)
	bucketPgnr, err := s.bucketRefIdx.HashedPaginate(lastHeader.Uid, idx.BottomToTop)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("Loaded bucketRef for uid: %v / hash: %v\n", uid, lastBucketHashedUid)

	// Use all layers until first root layer
	var lastMetadata *Metadata

	for entry := range bucketPgnr.All() {
		if entry.Error() != nil {
			return nil, entry.Error()
		}
		metadata, err := loadMetadata(&entry.Val().MetadataRef)
		if err != nil {
			return nil, err
		}
		if lastMetadata == nil {
			lastMetadata = metadata
			break
		}
	}

	b := newBucket(s)
	// b.lastHashedUid = *lastBucketHashedUid
	b.Header = *lastHeader
	b.Metadata = lastMetadata
	b.saved = true
	return b, nil
}

func (s *bucketService) Get(uid BucketUid) (*Bucket, error) {
	// 1- Get last bucket header
	headerPgnr, err := s.headerRefIdx.HashedPaginate(uid, idx.BottomToTop)
	if err != nil {
		return nil, err
	}

	lastHeader, err := getLastBucketHeader(headerPgnr)
	if err != nil {
		return nil, err
	}

	return s.buildLazyBucket(lastHeader)
}

func (s *bucketService) Filter(o idx.Order, f idx.Filter, pageSize, preloadPageCount int) (idx.Paginer[BucketUid, *Bucket], error) {
	// Build a paginer of matching Buckets
	// The paginer lazy load buckets
	// panic("not implemented yet")

	// 1- Get a paginer of matching Bucket headers
	headerPgnr, err := s.headerRefIdx.FilterAll(o, f)
	if err != nil {
		return nil, err
	}

	builtBuckets := make(map[BucketUid]*Bucket)
	// 2- Return an iterator building a bucket for each header
	return idx.NewPaginer(pageSize, preloadPageCount, func(push func(e idx.Entry[BucketUid, *Bucket]) bool) {
		for headerEntry := range headerPgnr.All() {
			if headerEntry.Error() != nil {
				if !push(idx.NewErrEntry[BucketUid, *Bucket](err)) {
					break
				}
			}
			if _, ok := builtBuckets[headerEntry.Key()]; ok {
				// Bucket already built and pushed
				continue
			}
			if err != nil {
				if !push(idx.NewErrEntry[BucketUid, *Bucket](err)) {
					break
				}
			}

			lastHeader, err := loadHeader(headerEntry.Val())
			if err != nil {
				if !push(idx.NewErrEntry[BucketUid, *Bucket](err)) {
					break
				}
			}
			// lastBucketHashedUid := HashedBucketUid(headerEntry.KeyBytes())
			b, err := s.buildLazyBucket(lastHeader)
			builtBuckets[headerEntry.Key()] = b
			if !push(idx.NewEntry(headerEntry.Key(), b, headerEntry.Seq(), headerEntry.Time(), headerEntry.State(), err, headerEntry.KeyBytes())) {
				break
			}
		}
	}), nil
}

func (s *bucketService) Names() (map[string]BucketUid, error) {
	nameIt, err := s.bucketNameIdx.All(idx.TopToBottom)
	if err != nil {
		return nil, err
	}

	nameByUidMap := make(map[BucketUid]string)
	lastUidByNameMap := make(map[string]BucketUid)
	for entry := range nameIt {
		if entry.Error() != nil {
			return nil, entry.Error()
		}
		nameByUidMap[BucketUid(entry.Val())] = entry.Key()
		lastUidByNameMap[entry.Key()] = entry.Val()
	}

	return lastUidByNameMap, nil
}

func (s *bucketService) buildLayerIt(b *Bucket, version Version) (iter.Seq2[error, *Layer], error) {
	// We want to iterate over last layers only, starting with the last root layer by default.
	// We may have some layers already loaded in b.loadedLayers.
	// Need to iterate over all layers metadata
	// Load content only if necessary

	bucketPgnr, err := s.bucketRefIdx.HashedPaginate(b.Header.Uid, idx.BottomToTop)
	if err != nil {
		return nil, err
	}

	// Eagerly load all Layer metadatas
	for entry := range bucketPgnr.All() {
		if entry.Error() != nil {
			return nil, fmt.Errorf("error iterating bucketRefIdx: %w", entry.Error())
		}
		metadata, err := loadMetadata(&entry.Val().MetadataRef)
		if err != nil {

		}
		b.loadedBucketEntries[metadata.Version] = &entry
		b.loadedMetadatas[metadata.Version] = metadata
		b.maxLoadedVersion = max(b.maxLoadedVersion, metadata.Version)
	}

	if version == LatestVersion {
		version = b.maxLoadedVersion
	}

	// First implem : for now take all layers
	firstVersion := FirstLayerVersion
	return func(yield func(error, *Layer) bool) {
		for v := firstVersion; v <= version; v++ {
			var ok bool
			var l *Layer
			if l, ok = b.loadedLayers[v]; !ok {
				// Layer not already loaded
				bucketEntry, ok := b.loadedBucketEntries[v]
				if !ok {
					panic(fmt.Sprintf("bucketEntry v%d not loaded", v))
				}
				// data, err := files.LoadBlocPart((*filez.BlocPart)(&(*bucketEntry).Val().LayerRef))
				data, err := loadLayerData(&(*bucketEntry).Val().LayerRef)
				if err != nil {
					if !yield(err, nil) {
						break
					}
				}
				metadata := b.loadedMetadatas[v]
				l = &Layer{
					Metadata: metadata,
					Content:  data,
					State:    (*bucketEntry).State(),
				}
				b.loadedLayers[metadata.Version] = l
			}

			if !yield(nil, l) {
				break
			}
		}
	}, nil
}

func storeHeader(dir, partition string, h *Header) (*HeaderRef, error) {
	serializer := serialize.StructSerializer[Header]{}
	data := make([]byte, 1000) // FIXME: what is the good size ?
	n, err := serializer.Serialize(h, data)
	if err != nil {
		return nil, err
	}

	virtualBloc, err := files.StoreBlocData(dir, partition, "bucketHeader", data[:n])
	if err != nil {
		return nil, err
	}

	if len(virtualBloc.Parts()) > 1 {
		panic("does not support multi part writes")
	}

	ref := HeaderRef(virtualBloc.Parts()[0])
	return &ref, nil
}

func storeMetadata(dir, partition string, m *Metadata) (*MetadataRef, error) {
	serializer := serialize.StructSerializer[Metadata]{}
	data := make([]byte, 1000) // FIXME: what is the good size ?
	n, err := serializer.Serialize(m, data)
	if err != nil {
		return nil, err
	}

	virtualBloc, err := files.StoreBlocData(dir, partition, "layerMetadata", data[:n])
	if err != nil {
		return nil, err
	}

	if len(virtualBloc.Parts()) > 1 {
		panic("does not support multi part writes")
	}

	ref := MetadataRef(virtualBloc.Parts()[0])
	return &ref, nil
}

func storeLayerData(dir, partition string, data []byte) (*LayerRef, error) {
	virtualBloc, err := files.StoreBlocData(dir, partition, "layerData", data)
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

func loadLayerData(ref *LayerRef) ([]byte, error) {
	b, err := files.LoadBlocPart((*filez.BlocPart)(ref))
	return b, err
}

// (KEY: string, VAL: BucketUid)
func newBucketNameIndex(indexDir, partition string) (BucketNameIndex, error) {
	enc := idx.NewAsciiEncoder(0, BucketNameIdxStateSize, BucketNameIdxKeySize, BucketNameIdxDataSize)
	keySer := serialize.AsciiSerializer{}
	valSer := BucketUidSerializer{}
	return idx.NewBasicIndex(indexDir, BucketNameIdxQualifier, partition, keySer, valSer,
		nil, nil, enc, BucketNameIdxPageSize, 0)
}

// (KEY: RH(BucketUid), VAL: HeaderRef)
func newHeaderRefIndex(indexDir, partition, salt string) (HeaderRefIndex, error) {
	enc := idx.NewAsciiEncoder(0, HeaderRefIdxStateSize, HeaderRefIdxKeySize, HeaderRefIdxDataSize)
	keySer := BucketUidSerializer{}
	valSer := serialize.StructSerializer[HeaderRef]{}
	return idx.NewBasicIndex(indexDir, HeaderRefIdxQualifier, partition, keySer, valSer,
		idx.NewRotatingHasher([]byte(salt), HeaderRefIdxKeySize), nil, enc, HeaderRefIdxPageSize, 0)
}

// (KEY: H(BucketUid), VAL: BucketRef)
func newBucketRefIndex(indexDir, partition, salt string) (BucketRefIndex, error) {
	enc := idx.NewAsciiEncoder(0, BucketRefIdxStateSize, BucketRefIdxKeySize, BucketRefIdxDataSize)
	keySer := BucketUidSerializer{}
	valSer := serialize.StructSerializer[BucketRef]{}
	return idx.NewBasicIndex(indexDir, BucketRefIdxQualifier, partition, keySer, valSer,
		idx.NewRotatingHasher([]byte(salt), BucketRefIdxKeySize), nil, enc, BucketRefIdxPageSize, 0)
}

func generateRandUid() BucketUid {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		panic(err)
	}
	return BucketUid(randBytes)
}

func forgeIndexesDir(dir, partition string) (string, string, string, error) {
	bucketIdxDir := filepath.Join(dir, partition, "bucketIdx")
	headerRefIdxDir := filepath.Join(dir, partition, "headerRefIdx")
	bucketRefIdxDir := filepath.Join(dir, partition, "bucketRefIdx")

	err := os.MkdirAll(bucketIdxDir, 0700)
	if err != nil {
		return "", "", "", err
	}

	err = os.MkdirAll(headerRefIdxDir, 0700)
	if err != nil {
		return "", "", "", err
	}

	err = os.MkdirAll(bucketRefIdxDir, 0700)
	if err != nil {
		return "", "", "", err
	}

	return bucketIdxDir, headerRefIdxDir, bucketRefIdxDir, nil
}

func getServiceIndexes(dir, salt, partition string) (BucketNameIndex,
	HeaderRefIndex, BucketRefIndex, error) {
	bucketIdxDir, headerRefIdxDir, bucketRefIdxDir, err := forgeIndexesDir(dir, partition)
	if err != nil {
		return nil, nil, nil, err
	}

	bucketNameIdx, err := newBucketNameIndex(bucketIdxDir, "")
	if err != nil {
		return nil, nil, nil, err
	}
	headerRefIdx, err := newHeaderRefIndex(headerRefIdxDir, "", salt)
	if err != nil {
		return nil, nil, nil, err
	}
	bucketRefIdx, err := newBucketRefIndex(bucketRefIdxDir, "", salt)
	if err != nil {
		return nil, nil, nil, err
	}

	return bucketNameIdx, headerRefIdx, bucketRefIdx, nil
}

func scanServicePartitions(dir string) ([]string, error) {
	dirs, err := filepath.Glob(filepath.Join(dir, "*"))
	return dirs, err
}
