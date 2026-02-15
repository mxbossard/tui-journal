package immutxtdb

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"path/filepath"
	"sync"

	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/inoutz"
)

const (
	indexLineBufferSize = 1000
	blocBufferSize      = 1000
	delimiterChar       = ','
	newLineChar         = '\n'
	defaultPageSize     = 10
)

type index[K comparable, V any] interface {
	Add(key K, val V) error
	Paginate(key K, order IdxOrder, pageSize int) (*paginer[K, V], chan error)
	PaginateAll(order IdxOrder, pageSize int) (*paginer[K, V], chan error)
	All(order IdxOrder, errChan chan error) iter.Seq2[K, V]
	Count() (int, error)
}

type bucketIndex struct {
	*sync.Mutex
	// FIXME: add a filelock
	index[string, bool]

	filepathes     []string
	deviceIdxFiles []*filez.BlocsFile
	otherIdxFiles  []*filez.BlocsFile
	seqs           map[string]int
}

func newBucketIndex(bucketDir, device string) (*bucketIndex, error) {
	// Init bucketIndex
	// FIXME: manage multiple idx files (rotation)
	// FIXME: add a filelock
	firstDeviceFilepath := filepath.Join(bucketDir, fmt.Sprintf("bucket-%s-001.idx"))
	dbf1, err := filez.NewBlocsFile(firstDeviceFilepath, 256, 100)
	if err != nil {
		return nil, err
	}
	return &bucketIndex{
		Mutex:          &sync.Mutex{},
		deviceIdxFiles: []*filez.BlocsFile{dbf1},
		otherIdxFiles:  nil,
		seqs:           make(map[string]int),
	}, nil
}

func (i *bucketIndex) preload() error {
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

		_, seq, _, err := decodeLastIndexLine[any](buf.Bytes()[0:n])
		if err != nil {
			return err
		}
		i.seqs[bf.Name()] = seq

	}
	return nil
}

func (i *bucketIndex) selectDeviceBlocFile(uid string) *filez.BlocsFile {
	return i.deviceIdxFiles[0]
}

func (i *bucketIndex) Add(uid string, ok bool) error {
	i.Lock()
	defer i.Unlock()
	// TODO: write to block file
	// TODO: how to connect with block file ?
	// TODO: how to manage bloc file appending ?
	bf := i.selectDeviceBlocFile(uid)

	bfName := bf.Name()
	seq := i.seqs[bfName]

	// FIXME: use good byte encoding !
	entry, err := encodeIndexLine(seq, uid)
	if err != nil {
		return err
	}
	_, err = bf.Write(entry)
	if err != nil {
		return err
	}
	i.seqs[bfName] = seq + 1
	return nil
}

func (i *bucketIndex) Count() (int, error) {
	// Should be performent and not read all the index to count lines.
	i.Lock()
	defer i.Unlock()
	count := 0
	for _, k := range i.seqs {
		count += k
	}
	return count, nil
}

func (i *bucketIndex) Paginate(key string, order IdxOrder, limit int) (*paginer[string, bool], chan error) {
	panic("not implemented yet")
}

func (i *bucketIndex) PaginateAll(order IdxOrder, limit int) (*paginer[string, bool], chan error) {
	// TODO: cache all the bloc file content ?
	// TODO: call all the index content ?
	errChan := make(chan error)
	idxFiles := append(i.deviceIdxFiles, i.otherIdxFiles...)
	p := newPaginer[string, bool](defaultPageSize, 0, errChan, func(push func(k string, v bool) bool) {
		//panic("not implemented yet")
		for _, bf := range idxFiles {
			for b := range bf.All(filez.TopToBottom, errChan) {
				sit := inoutz.NewSplitIterator(b, newLineChar, blocBufferSize)
				for line := range sit.All(errChan) {
					_ = line
					if !push(line, true) {
						return
					}
				}
			}
		}
	})
	return p, errChan

	panic("not implemented yet")
}

type layerIndex struct {
	index[string, *layer]
	filepathes []string
}

func (i *layerIndex) Add(uid string, l *layer) error {
	// Write to plain text file but private data is hashed
	panic("not implemented yet")
}

func (i *layerIndex) Count() (int, error) {
	// FIXME: Is it needed ?
	panic("not implemented yet")
}

func (i *layerIndex) Paginate(key string, order IdxOrder, limit int) (*paginer[string, *layer], chan error) {
	// TODO: build a cursor to iterate of the layers of a bucket
	panic("not implemented yet")
}

func (i *layerIndex) PaginateAll(order IdxOrder, limit int) (*paginer[string, *layer], chan error) {
	// FIXME: Is it needed ?
	panic("not implemented yet")
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
