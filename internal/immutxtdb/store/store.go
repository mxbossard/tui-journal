package store

import (
	"path/filepath"

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

	namesCache map[string]bucket.BucketUid
}

func NewTwoPhasesStore(dir, salt string) (*TwoPhasesStore, error) {
	// Multiple partitions SHOULD be managed by bucket service ?
	// Or do I manage one service by partition ?
	// For ephemeral bucket I may need one partition by bucket so one service by bucket ?
	// Supply partition for writes operation (New(), Save(), Commit())
	// Add a bucket.NewPartedBucketService ?
	// Update Service to manage multiple partitions ?
	// FIXME ephemeral store should be stored in home local cache dir ?
	partition := "foo"
	eStore, err := bucket.NewBucketService(filepath.Join(dir, "ephemeral"), partition, salt)
	if err != nil {
		return nil, err
	}
	rStore, err := bucket.NewBucketService(filepath.Join(dir, "rested"), partition, salt)
	if err != nil {
		return nil, err
	}
	s := &TwoPhasesStore{
		ephemeral:  eStore,
		rested:     rStore,
		namesCache: make(map[string]bucket.BucketUid),
	}
	return s, nil
}

func (s TwoPhasesStore) NewBucket(name, partition string, labels bucket.Labels) *bucket.Bucket {
	b := s.ephemeral.New(name, partition, labels)
	s.namesCache[name] = b.Header.Uid
	return b
}

// Save in ephemeral store
func (s TwoPhasesStore) Save(b *bucket.Bucket) error {
	err := s.ephemeral.Save(b)
	return err
}

// Commit in rested store
func (s TwoPhasesStore) Commit(b *bucket.Bucket, squash bool) error {
	if squash {
		panic("not implemented yet")
	}

	// TODO: promote bucket from ephemeral to rested store.

	// How to import bucket from one service to another ?
	// Service COULD implements an Export() / Import() pair of methods to :
	// - duplicate a bucket with exactly same data
	// - duplicate a bucket with some layers squashed
	// - compact buckets data in minimum count of blocs

	// How to erase data from ephemeral store ?
	// Idea 1: each ephemeral bucket could be created in a separate partition based on bucket name. Drawbacks: could not have 2 ephemeral buckets with same name.
	// Idea 2: Use idx.Hide considering all ephemeral hidden entries not existing. Drawbacks: how to clean ephemeral store ?
	// Idea 3: Use a bucket.Service not based on blocs files but on rewritable files ?

	// If squashing HOW TO rewrite layers in ephemeral store ?
	// Do we want to erase Bucket from ephemeral store ?
	// We could use a dedicated bloc file by Bucket in ephemeral store ?

	panic("not implemented yet")
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

func (s TwoPhasesStore) Filter(o idx.Order, f idx.Filter, pageSize, preloadPageCount int) (idx.Paginer[bucket.BucketUid, *bucket.Bucket], error) {
	pe, err := s.ephemeral.Filter(o, f, pageSize, preloadPageCount)
	if err != nil {
		return nil, err
	}
	pr, err := s.rested.Filter(o, f, pageSize, preloadPageCount)
	if err != nil {
		return nil, err
	}

	// TODO: Need to concat 2 paginers ordered by what ?
	// need to apply a distinct filter to not list same bucket from ephemeral & stored.
	_ = pe
	_ = pr

	panic("not implemented yet")
}
