package index

import (
	"bytes"
	"io"
	"sync"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
	"github.com/mxbossard/utilz/filez"
)

const (
	BucketIdxStateSize = 8
	bucketIdxKeySize   = 0
	bucketIdxDataSize  = 100
	bucketIdxPageSize  = 10
	bucketIdxQualifier = "bucket"
)

type BucketIndex idx.Index[idx.Void, string]

// (KEY: Void, STATE, VAL: BUCKET_UID)
func NewBucketIndex(indexDir, device string) (BucketIndex, error) {
	enc := idx.NewAsciiEncoder(0, BucketIdxStateSize, bucketIdxKeySize, bucketIdxDataSize)
	return idx.NewBasicIndex[idx.Void](indexDir, bucketIdxQualifier, device, nil, enc, bucketIdxPageSize)
}

type BucketIndex0 struct {
	idx.Index[idx.Void, string]
	*sync.Mutex
	// FIXME: add a filelock

	encoder        idx.IdxEncoder[idx.Void]
	keySerializer  serialize.AsciiSerializer
	filepathes     []string
	deviceIdxFiles []*filez.BlocsFile
	otherIdxFiles  []*filez.BlocsFile
	seqs           map[string]int
}

func (i *BucketIndex0) preload() error {
	i.Lock()
	defer i.Unlock()
	// TODO: read current seq in all files
	// TODO: load last blocs in cache ?

	idxFiles := append(i.deviceIdxFiles, i.otherIdxFiles...)
	for _, bf := range idxFiles {
		b, err := bf.GetLastNonEmptyBloc()
		if err == filez.ErrNotExist {
			continue
		} else if err != nil {
			return err
		}

		buf := &bytes.Buffer{}
		n, err := io.Copy(buf, b)
		if err != nil {
			return err
		}

		seq, _, _, _, err := i.encoder.DecodeLastWord(buf.Bytes()[0:n])
		if err != nil {
			return err
		}
		i.seqs[bf.Name()] = seq

	}
	return nil
}

/*
type documentIndex struct {
	filepathes []string
}

func (i documentIndex) add(topic string, bucketUid string) error {
	panic("not implemented yet")
}

func (i documentIndex) count() (int, error) {
	panic("not implemented yet")
}

// Manage sorting
func (i documentIndex) listAll(uid string, limit int) (cursor[Bucket], error) {
	panic("not implemented yet")
}

type textIndex struct {
	filepathes []string
}

func (i textIndex) add(topic string, bucketUid string, pos, len int) error {
	panic("not implemented yet")
}

func (i textIndex) count() (int, error) {
	panic("not implemented yet")
}

// Manage sorting
func (i textIndex) listAll(uid string, limit int) (cursor[Bucket], error) {
	panic("not implemented yet")
}
*/
