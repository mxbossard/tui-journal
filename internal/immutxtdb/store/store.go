package store

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/bucket"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
)

// Store managing "2 phases indexing"
// 1 ephemeral store
// 1 rested store
// Commit() to promote ephemeral storage into rested storage
type TwoPhasesStore struct {
	ephemeral bucket.Service
	rested    bucket.Service

	writePartition string

	namesCache map[string]bucket.BucketUid
}

func NewTwoPhasesStore(ephemeralDir, restedDir, partition, salt string) (*TwoPhasesStore, error) {
	// Multiple partitions are supported by bucket service ?
	// For ephemeral bucket I may need one partition by bucket ?
	// Supply partition for writes operation (Save(), Commit())
	eStore, err := bucket.NewBucketService(filepath.Join(ephemeralDir, "ephemeral"), salt)
	if err != nil {
		return nil, err
	}
	rStore, err := bucket.NewBucketService(filepath.Join(restedDir, "rested"), salt)
	if err != nil {
		return nil, err
	}
	s := &TwoPhasesStore{
		ephemeral:      eStore,
		rested:         rStore,
		writePartition: partition,
		namesCache:     make(map[string]bucket.BucketUid),
	}
	return s, nil
}

func (s TwoPhasesStore) NewBucket(name string, labels bucket.Labels) *bucket.Bucket {
	b := s.ephemeral.New(name, labels)
	s.namesCache[name] = b.Header.Uid
	return b
}

func (s TwoPhasesStore) ephemeralBucketPartition(b *bucket.Bucket) string {
	return fmt.Sprintf("%s-%x", s.writePartition, b.Header.Uid)
}

// Save in ephemeral store at supplied time
func (s TwoPhasesStore) Save(b *bucket.Bucket, t ...time.Time) error {
	partition := s.ephemeralBucketPartition(b)
	if b.Metadata != nil && b.Metadata.Version > 0 {
		// For bucket save with version greater than 0
		_, err := s.ephemeral.Get(b.Header.Uid)
		if err == bucket.ErrNotExist {
			// If bucket not in ephemeral service, need to import if back from rested service.
			export, err := s.rested.Export(b.Header.Uid, false, false, false)
			if err != nil {
				return err
			}
			err = s.ephemeral.Import(export, partition)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	err := s.ephemeral.Save(b, partition, t...)
	return err
}

// Commit in rested store
func (s TwoPhasesStore) Commit(b *bucket.Bucket, squash bool) error {
	// How to import bucket from one service to another ?
	// Service COULD implements an Export() / Import() pair of methods to :
	// - duplicate a bucket with exactly same data
	// - duplicate a bucket with some layers squashed
	// - duplicate a bucket removing hidden data
	// - compact buckets data in minimum count of blocs
	// => Export(BucketUid, includeHidden, squash, squashAll) ExportedBucket
	// - Do we need to implements 2 squash ?
	// - includeHidden & squashAll seems mutually exclusive
	// - squash would merge as many a possible layers
	// - squashAll would produce only one layer (keep only last snapshot)

	// Service Export could leverage an Idx Export supporting :
	// - Export of hidden or not entries (if hide feature is implemented in Idx)
	// - Entry merging ? Based on what ? Is this not of idx user responsibility ?
	// - Could supply a merging func(e1, e2 Entry) (Entry, error)
	// - Compaction would be implemented by design, all entries would be inserted contiguously.
	// => No interest implementing export in Idx. Export can easily be done in another pkg.

	// How to erase data from ephemeral store ?
	// Idea 1: each ephemeral bucket could be created in a separate partition based on bucket name. Drawbacks: could not have 2 ephemeral buckets with same name.
	// Idea 1b: each ephemeral bucket could be created in a separate partition based on bucket UID. Drawbacks: none.
	// Idea 2: Use idx.Hide considering all ephemeral hidden entries not existing. Drawbacks: how to clean ephemeral store ?
	// Idea 3: Use a bucket.Service not based on blocs files but on rewritable files ?

	// 1- Import ephemeral bucket in rested service
	export, err := s.ephemeral.Export(b.Header.Uid, false, false, squash)
	if err != nil {
		return fmt.Errorf("unable to export bucket: %w", err)
	}
	err = s.rested.Import(export, s.writePartition)
	if err != nil {
		return fmt.Errorf("unable to import bucket: %w", err)
	}

	// 2- After ephemeral import => erase ephemeral partition
	partition := s.ephemeralBucketPartition(b)
	err = s.ephemeral.ErasePartition(partition)
	if err != nil {
		return fmt.Errorf("unable to erase partition: %w", err)
	}

	return nil
}

// Return all Names associated with it's last Bucket Uid
func (s TwoPhasesStore) Names() (map[string]bucket.BucketUid, error) {
	if len(s.namesCache) == 0 {
		s.namesCache = make(map[string]bucket.BucketUid)

		names, err := s.rested.Names()
		if err != nil {
			return nil, err
		}
		for name, uid := range names {
			s.namesCache[name] = uid
		}

		names, err = s.ephemeral.Names()
		if err != nil {
			return nil, err
		}
		for name, uid := range names {
			s.namesCache[name] = uid
		}
	}

	return s.namesCache, nil
}

func (s TwoPhasesStore) Get(uid bucket.BucketUid) (*bucket.Bucket, error) {
	// Need to get from ephemeral store if it exists or from rested store
	b, err := s.ephemeral.Get(uid)
	if err == bucket.ErrNotExist {
		b, err = s.rested.Get(uid)
	}
	return b, err
}

func (s TwoPhasesStore) Filter(sort bucket.Sort, c bucket.Criteria, pageSize, preloadPageCount int) (idx.Paginer[bucket.BucketUid, *bucket.Bucket], error) {
	pe, err := s.ephemeral.Filter(sort, c, pageSize, preloadPageCount)
	if err != nil {
		return nil, err
	}
	pr, err := s.rested.Filter(sort, c, pageSize, preloadPageCount)
	if err != nil {
		return nil, err
	}

	concat := idx.CatPaginers(idx.EntryTimeCompare[bucket.BucketUid, *bucket.Bucket](idx.TopToBottom), pageSize, preloadPageCount, pe, pr)
	distinct := idx.DistinctPaginer(func(entries ...idx.Entry[bucket.BucketUid, *bucket.Bucket]) idx.Entry[bucket.BucketUid, *bucket.Bucket] {
		// keep bucket with higher version
		var maxVersion bucket.Version
		var selected idx.Entry[bucket.BucketUid, *bucket.Bucket]
		for _, e := range entries {
			if e.Val().Metadata.Version > maxVersion {
				selected = e
				maxVersion = e.Val().Metadata.Version
			}
		}
		return selected
	}, concat)

	return distinct, nil
}
