package bucket

import (
	"cmp"
	"crypto/rand"
	"errors"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/files"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/zip"
	"github.com/mxbossard/utilz/collectionz"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/iterz"
)

const (
	bucketIdxDir = "bucketIndexes"

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
	ErrNotExist         = errors.New("bucket do not exists")
	ErrVersionMissmatch = errors.New("bucket version missmatch")
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
	// FIXME: filter on which terms ? CANNOT reuse idx filters and use it on all bucket indexes.
	Filter(o Sort, f Criteria, pageSize, preloadPageCount int) (idx.Paginer[BucketUid, *Bucket], error)
	// Save a Bucket in supplied partition.
	// If a time is supplied will use it for creation or update time.
	Save(b *Bucket, partition string, t ...time.Time) error
	// Export all layers of a bucket
	Export(uid BucketUid, headerHist, dataHist, squash bool) (*BucketExport, error)
	// Import an Exported Bucket in the service
	Import(export *BucketExport, partition string) error
	// Erase a partition of the index
	ErasePartition(partition string) error

	buildLayerIt(b *Bucket, v Version) (iter.Seq2[error, *Layer], error)
}

type bucketService struct {
	*sync.Mutex

	dir  string
	salt string
}

func NewBucketService(dir, salt string) (*bucketService, error) {
	return &bucketService{
		Mutex: &sync.Mutex{},
		dir:   dir,
		salt:  salt,
	}, nil
}

func (s *bucketService) New(name string, labels Labels) *Bucket {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	return s.new(name, labels)
}

func (s *bucketService) new(name string, labels Labels) *Bucket {
	uid := generateRandUid()
	// FIXME: check if uid already exists

	b := newNamedBucket(s, uid, name)
	b.Header.changed = true
	b.Header.Labels = labels

	return b
}

func (s *bucketService) NewText(name string, labels Labels, text string) (*Bucket, error) {
	b := s.New(name, labels)
	err := b.SetText(text)
	return b, err
}

func (s *bucketService) addName(partition string, t time.Time, name string, uid BucketUid) error {
	bucketNameIdx, err := getBucketNameIndex(s.dir, s.salt, partition)
	if err != nil {
		return err
	}
	_, err = bucketNameIdx.Add(dummyState, t, name, uid)
	if err != nil {
		return fmt.Errorf("unable to add name: %w", err)
	}
	return nil
}

func (s *bucketService) addHeader(partition string, state idx.State, h *Header) (idx.Entry[BucketUid, *HeaderRef], error) {
	// Store data
	headerRef, err := storeHeader(s.dir, partition, h)
	if err != nil {
		h.Created = nil
		return nil, fmt.Errorf("unable to store header: %w", err)
	}

	// Store ref
	headerRefIdx, err := getHeaderRefIndex(s.dir, s.salt, partition)
	if err != nil {
		return nil, err
	}
	hrEntry, err := headerRefIdx.Add(state, *h.Created, h.Uid, headerRef)
	if err != nil {
		return nil, fmt.Errorf("unable to add header ref: %w", err)
	}
	return hrEntry, nil
}

func (s *bucketService) addLayer(partition string, state idx.State, uid BucketUid, m *Metadata, d []byte) (idx.Entry[BucketUid, *BucketRef], error) {
	// Store metadata & data
	layerRef, err := storeLayerData(s.dir, partition, d)
	if err != nil {
		return nil, fmt.Errorf("unable to store layer: %w", err)
	}
	metadataRef, err := storeMetadata(s.dir, partition, m)
	if err != nil {
		return nil, fmt.Errorf("unable to store metadata: %w", err)
	}

	// Store refs
	bucketRef := BucketRef{
		MetadataRef: *metadataRef,
		LayerRef:    *layerRef,
	}
	bucketRefIdx, err := getBucketRefIndex(s.dir, s.salt, partition)
	if err != nil {
		return nil, err
	}
	brEntry, err := bucketRefIdx.Add(state, *m.Updated, uid, &bucketRef)
	if err != nil {
		return nil, fmt.Errorf("unable to add bucket ref: %w", err)
	}

	return brEntry, nil
}

// Create a bucket in supplied partition at suplied time
func (s *bucketService) create(b *Bucket, partition string, t time.Time) error {
	b.Header.Created = &t
	// if b.Header.Created == nil {
	// 	now := time.Now()
	// 	b.Header.Created = &time
	// }

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

	// TODO: 1- Add (name, uid) in bucketName Idx
	err := s.addName(partition, *b.Header.Created, b.Header.Name, b.Header.Uid)
	if err != nil {
		return fmt.Errorf("create: unable to add Name: %w", err)
	}

	// TODO: 2- store Header
	_, err = s.addHeader(partition, dummyState, &b.Header)
	if err != nil {
		b.Header.Created = nil
		return fmt.Errorf("create: unable to add Header: %w", err)
	}

	// TODO: 3- store Layer & Metadata
	metadata := Metadata{
		Uid:     b.Header.Uid,
		Version: FirstLayerVersion,
		Size:    dataLen,
		Updated: b.Header.Created,
	}
	brEntry, err := s.addLayer(partition, RootLayerState, b.Header.Uid, &metadata, data)
	if err != nil {
		return fmt.Errorf("create: unable to add Layer: %w", err)
	}
	// fmt.Printf("Added bucketRef for uid: %v / hash: %v\n", b.header.Uid, hashedUid)

	l := Layer{
		Metadata: &metadata,
		Content:  data,
		State:    brEntry.State(),
	}

	b.Header.changed = false
	b.Metadata = &metadata
	b.maxLoadedVersion = metadata.Version
	b.loadedBucketEntries[PartedVersion{metadata.Version, partition}] = &brEntry
	b.loadedMetadatas[PartedVersion{metadata.Version, partition}] = &metadata
	b.loadedLayers[PartedVersion{metadata.Version, partition}] = &l

	return nil
}

// Update a bucket in supplied partition at suplied time
func (s *bucketService) update(b *Bucket, partition string, t time.Time) error {
	// now := time.Now()

	// 1- Attempt to make a patch of the update.
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
		} else {
			// Store a diff layer
			diffLayer = true
			data = zipedPatch
		}
		dataLen = len(b.stringData)
	}

	newMetadata := &Metadata{
		Uid:     b.Header.Uid,
		Version: b.Metadata.Version + 1,
		Size:    dataLen,
		Updated: &t,
	}

	// 2- Add (RhUid, headerRef) in headerRef Idx if Header changed
	if b.Header.changed {
		b.Header.Modified = &t

		entry, err := s.addHeader(partition, dummyState, &b.Header)
		if err != nil {
			return fmt.Errorf("update: unable store header: %w", err)
		}

		_ = entry
		panic("not implemented yet")
	}

	var layerState idx.State
	if rootLayer {
		layerState = RootLayerState
	} else if diffLayer {
		layerState = DiffLayerState
	}

	brEntry, err := s.addLayer(partition, layerState, b.Header.Uid, newMetadata, data)
	if err != nil {
		return fmt.Errorf("update: unable to add Layer: %w", err)
	}

	l := Layer{
		Metadata: newMetadata,
		Content:  data,
		State:    brEntry.State(),
	}

	b.maxLoadedVersion = newMetadata.Version
	b.loadedBucketEntries[PartedVersion{newMetadata.Version, partition}] = &brEntry
	b.loadedMetadatas[PartedVersion{newMetadata.Version, partition}] = newMetadata
	b.loadedLayers[PartedVersion{newMetadata.Version, partition}] = &l
	b.Header.changed = false
	b.Metadata = newMetadata
	return nil
}

func (s *bucketService) Save(b *Bucket, partition string, t ...time.Time) error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	if b.saved {
		return nil
	}

	var operationTime time.Time
	if len(t) > 1 {
		panic("must supply only one time param")
	} else if len(t) == 1 {
		operationTime = t[0]
	} else {
		operationTime = time.Now()
	}

	// 1- Check if supplied bucket already exists
	// if b.Metadata == nil {
	// 	return s.create(b, partition, operationTime)
	// }
	// return s.update(b, partition, operationTime)

	var stored *Bucket
	var err error
	if b.Metadata != nil {
		stored, err = s.Get(b.Metadata.Uid)
	}
	if b.Metadata == nil || err == ErrNotExist {
		// swallow ErrNotExist and create the bucket
		return s.create(b, partition, operationTime)
	} else if err != nil {
		return err
	}

	// 2- Check if supplied bucket version is equal to stored bucket version
	// Need to check the last version in the partition for consistency ?
	stored, err = s.Get(b.Metadata.Uid)
	if err == ErrNotExist {
		// swallow not exist error it means the bucket is first save in this partition
		err = nil
	} else if err != nil {
		return err
	}
	if stored != nil && b.Metadata.Version != stored.Metadata.Version {
		return ErrVersionMissmatch
	}

	return s.update(b, partition, operationTime)
}

func (s *bucketService) buildLazyBucket(lastHeader *Header) (*Bucket, error) {
	existingParts, err := scanServicePartitions(s.dir)
	if err != nil {
		return nil, err
	}

	var lastMetadata *Metadata
	for _, partition := range existingParts {
		bucketRefIdx, err := getBucketRefIndex(s.dir, s.salt, partition)
		if err != nil {
			return nil, err
		}

		// 2- List all bucketRefs
		// FIXME: which uid key ?
		// FIXME: use state filter to stop on first root state ?
		// fmt.Printf("buildLazyBucket, bUid: %v\n", lastHeader.Uid)
		bucketPgnr, err := bucketRefIdx.Paginate(lastHeader.Uid, idx.BottomToTop)
		if err != nil {
			return nil, err
		}
		// fmt.Printf("Loaded bucketRef for uid: %v / hash: %v\n", uid, lastBucketHashedUid)

		// Take last metadata

		for entry := range bucketPgnr.All() {
			if entry.Error() != nil {
				return nil, entry.Error()
			}
			metadata, err := loadMetadata(&entry.Val().MetadataRef)
			if err != nil {
				return nil, err
			}
			if lastMetadata == nil || lastMetadata.Updated.Before(*metadata.Updated) {
				lastMetadata = metadata
				break
			}
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
	existingParts, err := scanServicePartitions(s.dir)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("partitions to scan: %s\n", existingParts)
	var lastHeaderAllParts *Header
	for _, partition := range existingParts {
		lastHeader, err := s.getLastHeader(uid, partition)
		if err != nil && err != ErrNotExist {
			return nil, err
		}
		// fmt.Printf("found lastHeader for part: %s => %v\n", partition, lastHeader)
		if lastHeaderAllParts == nil || lastHeaderAllParts.Modified != nil && lastHeader.Modified != nil &&
			lastHeaderAllParts.Modified.Before(*lastHeader.Modified) {
			lastHeaderAllParts = lastHeader
		}
	}
	if lastHeaderAllParts == nil {
		return nil, ErrNotExist
		// return nil, fmt.Errorf("no header found for bucket: %x", uid)
	}

	return s.buildLazyBucket(lastHeaderAllParts)
}

func (s *bucketService) getLastHeader(uid BucketUid, partition string) (*Header, error) {
	// 0- get last header for all parts
	headerRefIdx, err := getHeaderRefIndex(s.dir, s.salt, partition)
	if err != nil {
		return nil, err
	}

	// 1- Get last bucket header
	headerPgnr, err := headerRefIdx.Paginate(uid, idx.BottomToTop)
	if err != nil {
		return nil, err
	}

	lastHeader, err := getLastBucketHeader(headerPgnr)
	return lastHeader, err
}

// SHOULD not be used because it cannot be consistent (buckets are distributed on partitions)
func (s *bucketService) get(uid BucketUid, partition string) (*Bucket, error) {
	lastHeader, err := s.getLastHeader(uid, partition)
	if err != nil {
		return nil, err
	}
	return s.buildLazyBucket(lastHeader)
}

func (s *bucketService) Filter(sort Sort, c Criteria, pageSize, preloadPageCount int) (idx.Paginer[BucketUid, *Bucket], error) {
	// FIXME: SHOULD order partitons using supplied Order
	existingParts, err := scanServicePartitions(s.dir)
	if err != nil {
		return nil, err
	}

	// FIXME: ordering SHOULD be implemented correctly
	var order idx.Order
	switch sort {
	case OlderFirst:
		order = idx.TopToBottom
	case YoungerFirst:
		order = idx.BottomToTop
	}

	var paginers []iter.Seq[idx.Entry[BucketUid, *Bucket]]
	for _, partition := range existingParts {
		bucketNameIdx, headerRefIdx, bucketRefIdx, err := getServiceIndexes(s.dir, s.salt, partition)
		if err != nil {
			return nil, err
		}

		// Build a paginer of matching Buckets
		// The paginer lazy load buckets
		// panic("not implemented yet")

		// TODO: Filtering
		// - BucketUID filter => Need to scan headerRefIdx key where BucketUID is maintained.
		// - Bucket state filter => Need to scan headerRefIdx state where bucket state is maintained.
		// - Bucket name filter => Need to scan bucketNameIdx to get matching BucketUID THEN filter by BucketUID.
		// - Bucket creationTime filter => Need to scan headerRefIdx time where creationTime is indexed
		// 		THEN thin filter on Header.Created time
		// - Bucket updateTime filter => Need to scan bucketRefIdx time where bucket updateTime is indexed
		// 		THEN thin filter on Metadata.Updated time THEN filter by BucketUID stored in Metadata

		// 1- If Name criteria resolve corresponding BucketUids
		uidCriteriaMandatory := false
		var uidsCriteria []BucketUid
		var uidsFromNamesCriteria []BucketUid
		if c.BucketNameMatcher() != nil {
			uidCriteriaMandatory = true
			kf, err := bucketNameIdx.KeysFilter(false, c.MatchingNames()...)
			if err != nil {
				return nil, err
			}
			paginer, err := bucketNameIdx.FilterAll(idx.BottomToTop, kf)
			if err != nil {
				return nil, err
			}
			for e := range paginer.All() {
				if err = e.Error(); err != nil {
					return nil, err
				}
				uidsFromNamesCriteria = append(uidsFromNamesCriteria, e.Val())
			}
			// fmt.Printf("Found uids from names: %v\n", uidsFromNamesCriteria)
		}

		// 2- If UpdateTime criteria resolve corresponding BucketUids
		var uidsFromUpdateTimeCriteria []BucketUid
		if c.UpdateTimeMatcher() != nil {
			uidCriteriaMandatory = true
			tf := idx.TimeFilter(func(t time.Time) (bool, bool) {
				matcher := c.UpdateTimeMatcher()
				ok := matcher(t)
				return ok, true
			})
			paginer, err := bucketRefIdx.FilterAll(idx.BottomToTop, tf)
			if err != nil {
				return nil, err
			}
			for e := range paginer.All() {
				if err = e.Error(); err != nil {
					return nil, err
				}
				// FIXME: add Metadata caching
				metadata, err := loadMetadata(&e.Val().MetadataRef)
				if err != nil {
					return nil, err
				}
				uidsFromUpdateTimeCriteria = append(uidsFromUpdateTimeCriteria, metadata.Uid)
			}
			// fmt.Printf("Found uids from updateTimes: %v\n", uidsFromUpdateTimeCriteria)
		}

		// 2- If BucketUid criteria Merge with uidsFromNamesCriteria
		headerRefIdxFilter := idx.NewFilter()
		if c.MatchingUids() != nil {
			uidsCriteria = append(uidsCriteria, c.MatchingUids()...)
		}
		if len(uidsFromNamesCriteria) > 0 {
			uidsCriteria = append(uidsCriteria, uidsFromNamesCriteria...)
		}
		if len(uidsFromUpdateTimeCriteria) > 0 {
			uidsCriteria = append(uidsCriteria, uidsFromUpdateTimeCriteria...)
		}
		if len(uidsCriteria) > 0 {
			kf, err := headerRefIdx.KeysFilter(false, uidsCriteria...)
			if err != nil {
				return nil, err
			}
			// fmt.Printf("Filtering on uids: %v [part: %s]\n", uidsCriteria, partition)
			headerRefIdxFilter.Add(kf)
		} else if uidCriteriaMandatory {
			// No matching uid but mandatory => no bucket matching
			continue
		}

		// 3- If State criteria add it to headerRefIdxFilter
		if c.BucketStateMatcher() != nil {
			f := idx.StateFilter(func(s idx.State) (bool, bool) {
				matcher := c.BucketStateMatcher()
				ok := matcher(State(s))
				return ok, true
			})
			headerRefIdxFilter.Add(f)
		}

		// 4- If CreationTime criteria add it to headerRefIdxFilter
		if c.CreationTimeMatcher() != nil {
			tf := idx.TimeFilter(func(t time.Time) (bool, bool) {
				matcher := c.CreationTimeMatcher()
				ok := matcher(t)
				return ok, true
			})
			headerRefIdxFilter.Add(tf)
		}

		// 5- Load Header matching criteria
		headerPgnr, err := headerRefIdx.FilterAll(order, headerRefIdxFilter)
		if err != nil {
			return nil, err
		}

		builtBuckets := make(map[BucketUid]*Bucket)
		// 2- Return an iterator building a bucket for each header
		paginer := idx.NewPaginer(pageSize, preloadPageCount, func(push func(e idx.Entry[BucketUid, *Bucket]) bool) {
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
				e := idx.NewEntry(headerEntry.Key(), b, headerEntry.Seq(), headerEntry.Time(), headerEntry.State(), err, headerEntry.KeyBytes())
				// fmt.Printf("pushing entry #%d of part: %s\n", headerEntry.Seq(), partition)
				if !push(e) {
					break
				}
			}
		})
		paginers = append(paginers, paginer.All())
	}

	compare := func(a, b idx.Entry[BucketUid, *Bucket]) int {
		// Ordered by time & partition
		i := cmp.Compare(a.Val().Metadata.Updated.UnixMilli(), b.Val().Metadata.Updated.UnixMilli())
		if sort == YoungerFirst {
			i = -i
		}
		return i
	}
	iterator := iterz.Merge(compare, paginers...)
	paginer := idx.NewPaginer(pageSize, preloadPageCount, func(push func(e idx.Entry[BucketUid, *Bucket]) bool) {
		for i := range iterator {
			// fmt.Printf("merging entry for buid: %v\n", i)
			if !push(i) {
				return
			}
		}
	})

	return paginer, nil
}

// FIXME: for now return last BucketUid by name found (last partition override previous ones)
func (s *bucketService) Names() (map[string]BucketUid, error) {
	existingParts, err := scanServicePartitions(s.dir)
	if err != nil {
		return nil, err
	}

	lastUidByNameMap := make(map[string]BucketUid)
	for _, partition := range existingParts {
		bucketNameIdx, err := getBucketNameIndex(s.dir, s.salt, partition)
		if err != nil {
			return nil, err
		}

		nameIt, err := bucketNameIdx.All(idx.TopToBottom)
		if err != nil {
			return nil, err
		}

		nameByUidMap := make(map[BucketUid]string)
		for entry := range nameIt {
			if entry.Error() != nil {
				return nil, entry.Error()
			}
			nameByUidMap[BucketUid(entry.Val())] = entry.Key()
			lastUidByNameMap[entry.Key()] = entry.Val()
		}
	}

	return lastUidByNameMap, nil
}

func (s *bucketService) ErasePartition(partition string) error {
	bucketNameIdxDir, headerRefIdxDir, bucketRefIdxDir, err := forgeIndexesDir(s.dir, partition)
	if err != nil {
		return err
	}
	err = os.RemoveAll(bucketNameIdxDir)
	if err != nil {
		return err
	}
	err = os.RemoveAll(headerRefIdxDir)
	if err != nil {
		return err
	}
	err = os.RemoveAll(bucketRefIdxDir)
	if err != nil {
		return err
	}
	return nil
}

func (s *bucketService) buildLayerIt(b *Bucket, version Version) (iter.Seq2[error, *Layer], error) {
	existingParts, err := scanServicePartitions(s.dir)
	if err != nil {
		return nil, err
	}

	// var iterators []iter.Seq2[error, *Layer]
	for _, partition := range existingParts {
		bucketRefIdx, err := getBucketRefIndex(s.dir, s.salt, partition)
		if err != nil {
			return nil, err
		}

		// We want to iterate over last layers only, starting with the last root layer by default.
		// We may have some layers already loaded in b.loadedLayers.
		// Need to iterate over all layers metadata
		// Load content only if necessary

		var f idx.Filter
		if lastSeq, ok := b.lastBucketRefSeq[partition]; ok {
			// Search for new bucketRef
			f = idx.AfterSeqFilter(lastSeq)
			// fmt.Printf("will scan all bucketRef after seq: %d\n", lastSeq)
		} else {
			// Search for all bucketRef
			f = idx.AfterSeqFilter(-1)
			// fmt.Printf("will scan all bucketRefs\n")
		}

		bucketPgnr, err := bucketRefIdx.Filter(b.Header.Uid, idx.BottomToTop, f)
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
				return nil, err
			}

			pv := PartedVersion{metadata.Version, partition}
			b.loadedBucketEntries[pv] = &entry
			b.loadedMetadatas[pv] = metadata
			b.maxLoadedVersion = max(b.maxLoadedVersion, metadata.Version)
			// fmt.Printf("loaded layer #%v metadata (seq: %d) maxLoadedVersion:%d\n", pv, entry.Seq(), b.maxLoadedVersion)
			b.lastBucketRefSeq[partition] = entry.Seq()
		}
	}

	if version == LatestVersion {
		version = b.maxLoadedVersion
	} else if version < 0 {
		version += b.maxLoadedVersion
	}

	if version <= 0 {
		// Asked for a version which do not exists.
		return nil, ErrNotExist
	} else if version > b.maxLoadedVersion {
		// Asked for a version which do not exists.
		return nil, ErrNotExist
	}

	iterator := func(yield func(error, *Layer) bool) {
		layerVersions := collectionz.Keys(b.loadedBucketEntries)
		slices.SortFunc(layerVersions, func(a, b PartedVersion) int {
			if a.Version < b.Version {
				return -1
			} else if a.Version == b.Version {
				return 0
			}
			return 1
		})
		for _, pv := range layerVersions {
			if pv.Version > version {
				// Stop iterating when catched up supplied version
				break
			}
			var ok bool
			var l *Layer
			if l, ok = b.loadedLayers[pv]; !ok {
				// Layer not already loaded

				bucketEntry, ok := b.loadedBucketEntries[pv]
				if !ok {
					// If bucket entry not loaded it does not exists for this partition
					// => skip it
					// FIXME: for now scan all version for each partition in b.loadedLayers map
					continue
					// panic(fmt.Sprintf("bucketEntry v%d not loaded", v))
				}

				data, err := loadLayerData(&(*bucketEntry).Val().LayerRef)
				if err != nil {
					if !yield(err, nil) {
						break
					}
				}
				metadata := b.loadedMetadatas[pv]
				// fmt.Printf("loaded metadata v%d: %v\n", v, metadata)
				l = &Layer{
					Metadata: metadata,
					Content:  data,
					State:    (*bucketEntry).State(),
				}
				// PartedVersion{metadata.Version, partition}
				b.loadedLayers[pv] = l
				// fmt.Printf("loaded layer #%v data\n", pv)
			}

			// fmt.Printf("yield %s layer v%d: %v\n", pv.Part, pv.Version, *l)
			if !yield(nil, l) {
				break
			}
		}
	}
	return iterator, nil
}

func storeHeader(dir, partition string, h *Header) (*HeaderRef, error) {
	serializer := serialize.StructSerializer[Header]{}
	data := make([]byte, 1000) // FIXME: what is the good size ?
	n, err := serializer.Serialize(h, data)
	if err != nil {
		return nil, err
	}

	virtualBloc, err := files.StoreBlocData(dir, partition, "bucketHeaderStore", data[:n])
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

	virtualBloc, err := files.StoreBlocData(dir, partition, "layerMetadataStore", data[:n])
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
	virtualBloc, err := files.StoreBlocData(dir, partition, "layerDataStore", data)
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

// return bucketNameIdxDir, headerRefIdxDir, bucketRefIdxDir
func forgeIndexesDir(dir, partition string) (string, string, string, error) {
	bucketNameIdxDir := filepath.Join(dir, bucketIdxDir, partition, "bucketNameIdx")
	headerRefIdxDir := filepath.Join(dir, bucketIdxDir, partition, "headerRefIdx")
	bucketRefIdxDir := filepath.Join(dir, bucketIdxDir, partition, "bucketRefIdx")

	err := os.MkdirAll(bucketNameIdxDir, 0700)
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

	return bucketNameIdxDir, headerRefIdxDir, bucketRefIdxDir, nil
}

func getBucketNameIndex(dir, salt, partition string) (BucketNameIndex, error) {
	bucketIdxDir, _, _, err := forgeIndexesDir(dir, partition)
	if err != nil {
		return nil, err
	}
	return newBucketNameIndex(bucketIdxDir, "")
}

func getHeaderRefIndex(dir, salt, partition string) (HeaderRefIndex, error) {
	_, headerRefIdxDir, _, err := forgeIndexesDir(dir, partition)
	if err != nil {
		return nil, err
	}
	return newHeaderRefIndex(headerRefIdxDir, "", salt)
}

func getBucketRefIndex(dir, salt, partition string) (BucketRefIndex, error) {
	_, _, bucketRefIdxDir, err := forgeIndexesDir(dir, partition)
	if err != nil {
		return nil, err
	}
	return newBucketRefIndex(bucketRefIdxDir, "", salt)
}

func getServiceIndexes(dir, salt, partition string) (BucketNameIndex,
	HeaderRefIndex, BucketRefIndex, error) {
	bucketNameIdx, err := getBucketNameIndex(dir, salt, partition)
	if err != nil {
		return nil, nil, nil, err
	}
	headerRefIdx, err := getHeaderRefIndex(dir, salt, partition)
	if err != nil {
		return nil, nil, nil, err
	}
	bucketRefIdx, err := getBucketRefIndex(dir, salt, partition)
	if err != nil {
		return nil, nil, nil, err
	}

	return bucketNameIdx, headerRefIdx, bucketRefIdx, nil
}

func scanServicePartitions(dir string) ([]string, error) {
	// FIXME put a preferenced partition first
	dirs, err := filepath.Glob(filepath.Join(dir, bucketIdxDir, "*"))
	var names []string
	for _, dir := range dirs {
		names = append(names, filepath.Base(dir))
	}
	sort.Strings(names)
	return names, err
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
