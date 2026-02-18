package index

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/encoder"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/mxbossard/utilz/filez"
)

// (BUCKET_UID, STATE_PRIVATE_DATA)
type BucketIndex struct {
	model.Index[string, model.State]
	*sync.Mutex
	// FIXME: add a filelock

	encoder        encoder.Encoder[string]
	filepathes     []string
	deviceIdxFiles []*filez.BlocsFile
	otherIdxFiles  []*filez.BlocsFile
	seqs           map[string]int
}

func NewBucketIndex(bucketDir, device string) (*BucketIndex, error) {
	// Init bucketIndex
	// FIXME: manage multiple idx files (rotation)
	// FIXME: add a filelock
	// FIXME: add BlocsEncryption
	firstDeviceFilepath := filepath.Join(bucketDir, fmt.Sprintf("bucket-%s-001.idx", device))
	dbf1, err := filez.NewBlocsFile(firstDeviceFilepath, 256, 100)
	if err != nil {
		return nil, err
	}
	e := encoder.NewAsciiEncoder(asciiEncoderDefaultVersion, asciiEncoderStateSize, asciiEncoderDataSize)
	idx := &BucketIndex{
		Mutex:          &sync.Mutex{},
		encoder:        e,
		deviceIdxFiles: []*filez.BlocsFile{dbf1},
		otherIdxFiles:  nil,
		seqs:           make(map[string]int),
	}

	// FIXME: need to setup the encoder!
	//e.Setup()

	return idx, nil
}

func (i *BucketIndex) preload() error {
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

		seq, _, _, err := i.encoder.DecodeLastWord(buf.Bytes()[0:n])
		// _, seq, _, err := decodeLastIndexLine[bool](buf.Bytes()[0:n])
		if err != nil {
			return err
		}
		i.seqs[bf.Name()] = seq

	}
	return nil
}

func (i *BucketIndex) selectDeviceBlocFile(uid string) *filez.BlocsFile {
	return i.deviceIdxFiles[0]
}

func (i *BucketIndex) Add(uid string, s model.State) error {
	i.Lock()
	defer i.Unlock()
	// TODO: write to block file
	// TODO: how to connect with block file ?
	// TODO: how to manage bloc file appending ?
	bf := i.selectDeviceBlocFile(uid)

	bfName := bf.Name()
	seq := i.seqs[bfName]

	// FIXME: use good byte encoding !
	entry, err := i.encoder.Encode(seq, s, uid)
	// entry, err := encodeIndexLine(seq, []byte(uid), s)
	if err != nil {
		return err
	}
	fmt.Printf("writing encoded content (#%d, uid: %s): %v\n", seq, uid, entry)
	_, err = bf.Write(entry)
	if err != nil {
		return err
	}
	i.seqs[bfName] = seq + 1
	return nil
}

func (i *BucketIndex) Count() (int, error) {
	// Should be performent and not read all the index to count lines.
	i.Lock()
	defer i.Unlock()
	count := 0
	for _, k := range i.seqs {
		count += k
	}
	return count, nil
}

func (i *BucketIndex) Paginate(key string, order model.Order, limit int) (model.Paginer[string, model.State], chan error) {
	panic("not implemented yet")
}

func (i *BucketIndex) PaginateAll(order model.Order, limit int) (model.Paginer[string, model.State], chan error) {
	// TODO: cache all the bloc file content ?
	// TODO: call all the index content ?
	errChan := make(chan error)
	idxFiles := append(i.deviceIdxFiles, i.otherIdxFiles...)
	p := model.NewPaginer(defaultPageSize, 0, func(push func(k string, v model.State, err error) bool) {
		//panic("not implemented yet")
		for _, bf := range idxFiles {
			for b := range bf.All(filez.BlocOrdering(order), errChan) {
				i.encoder.DecodeAll(order, b.Bytes(), func(seq int, s model.State, uid string, err error) {
					if !push(uid, s, err) {
						return
					}
				})
			}
		}
	})
	return p, errChan
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
