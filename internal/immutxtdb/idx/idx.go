package idx

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"iter"
	"path/filepath"
	"sync"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
	"github.com/mxbossard/utilz/filez"
)

/*
Ideas:
- Supply multiple basic Index impl:
  - Default (Key, Value) use a Serializer for Key and for Value ?
  - Uniq key ? => when adding an already existing key, override the previous entry which will never be returned anymore
  - Only value ? => drop the key will always use PaginateAll and never Paginate(key) ?


*/

type Order int

const (
	TopToBottom Order = iota
	BottomToTop
)

type State []byte

var dummyState = BuildState(8, "dummy")

type StateFilter func(s State, stop func()) bool
type KeyFilter[K any] func(k K, s State, stop func()) bool

func BuildState(size int, s string) State {
	data := make([]byte, size)
	_, err := binary.Encode(data, binary.BigEndian, []byte(s))
	if err != nil {
		panic(err)
	}
	return State(data)
}

type Index[K any, V any] interface {
	Add(s State, key K, val V) error
	Count() (int, error)
	Filter(key K, order Order, pageSize int, sf StateFilter) (Paginer[K, V], chan error)
	FilterAll(order Order, pageSize int, sf StateFilter, kf KeyFilter[K]) (Paginer[K, V], chan error)
	Paginate(key K, order Order, pageSize int) (Paginer[K, V], chan error)
	PaginateAll(order Order, pageSize int) (Paginer[K, V], chan error)
	All(order Order, errChan chan error) iter.Seq2[K, V]
}

type void struct{}
type Void *void

type Entry[K any, V any] interface {
	Key() K
	Val() V
	State() State
	Error() error
}

type BasicEntry[K any, V any] struct {
	key   K
	val   V
	state State
	err   error
}

func (e BasicEntry[K, V]) Key() K {
	return e.key
}

func (e BasicEntry[K, V]) Val() V {
	return e.val
}

func (e BasicEntry[K, V]) State() State {
	return e.state
}

func (e BasicEntry[K, V]) Error() error {
	return e.err
}

type basicIndex[K, V any] struct {
	Index[K, V]
	*sync.Mutex
	// FIXME: add a filelock

	keySerializer  serialize.Serializer[K]
	encoder        IdxEncoder[V] // FIXME: encoder must be attached to each BlocsFile or to each Bloc !
	pageSize       int
	filepathes     []string
	deviceIdxFiles []*filez.BlocsFile
	otherIdxFiles  []*filez.BlocsFile
	seqs           map[string]int
}

func NewBasicIndex[K, V any](indexDir, qualifier, device string, ser serialize.Serializer[K], enc IdxEncoder[V], pageSize int) (*basicIndex[K, V], error) {
	// Init bucketIndex
	// FIXME: manage multiple idx files (rotation)
	// FIXME: add a filelock
	// FIXME: addRotatingHash ?
	firstDeviceFilepath := filepath.Join(indexDir, fmt.Sprintf("%s-%s-001.idx", qualifier, device))
	dbf1, err := filez.NewBlocsFile(firstDeviceFilepath, 256, 100)
	if err != nil {
		return nil, err
	}
	idx := &basicIndex[K, V]{
		Mutex:          &sync.Mutex{},
		keySerializer:  ser,
		encoder:        enc,
		pageSize:       pageSize,
		deviceIdxFiles: []*filez.BlocsFile{dbf1},
		otherIdxFiles:  nil,
		seqs:           make(map[string]int),
	}

	// FIXME: need to setup the encoder!
	//e.Setup()

	return idx, nil
}

func (i *basicIndex[K, V]) selectDeviceBlocFile(s State, k []byte) *filez.BlocsFile {
	return i.deviceIdxFiles[0]
}

func (i *basicIndex[K, V]) Add(s State, k K, v V) error {
	i.Lock()
	defer i.Unlock()

	key := make([]byte, i.encoder.KeySize())
	var err error
	if i.keySerializer != nil {
		err = i.keySerializer.Serialize(k, key)
		if err != nil {
			return err
		}
	}

	bf := i.selectDeviceBlocFile(s, key)
	bfName := bf.Name()
	seq := i.seqs[bfName]

	entry, err := i.encoder.Encode(seq, s, key, v)
	if err != nil {
		return err
	}
	//fmt.Printf("writing encoded content (#%d, uid: %s): %v\n", seq, uid, entry)
	_, err = bf.Write(entry)
	if err != nil {
		return err
	}
	i.seqs[bfName] = seq + 1
	return nil
}

func (i *basicIndex[K, V]) Count() (int, error) {
	// Should be performent and not read all the index to count lines.
	i.Lock()
	defer i.Unlock()
	count := 0
	for _, k := range i.seqs {
		count += k
	}
	return count, nil
}

func (i *basicIndex[K, V]) Paginate(key K, order Order, limit int) (Paginer[K, V], chan error) {
	// TODO: build a cursor to entries with the supplied key
	errChan := make(chan error)

	filteringK := make([]byte, i.encoder.KeySize())
	var err error
	if i.keySerializer != nil {
		err = i.keySerializer.Serialize(key, filteringK)
		if err != nil {
			errChan <- err
			return nil, errChan
		}
	}

	idxFiles := append(i.deviceIdxFiles, i.otherIdxFiles...)
	p := NewPaginer(i.pageSize, 0, func(push func(State, K, V, error) bool) {
		//panic("not implemented yet")
		for _, bf := range idxFiles {
			for b := range bf.All(filez.BlocOrdering(order), errChan) {
				i.encoder.DecodeAll(order, b.Bytes(), func(seq int, s State, key []byte, val V, err error) {
					if err != nil {
						errChan <- err
						return
					}
					// FIXME: value missing ?
					if bytes.Equal(filteringK, key) {
						var k K
						if i.keySerializer != nil {
							k, err = i.keySerializer.Deserialize(key)
						}
						if !push(s, k, val, err) {
							return
						}
					}

				})
			}
		}
	})
	return p, errChan
}

func (i *basicIndex[K, V]) PaginateAll(order Order, limit int) (Paginer[K, V], chan error) {
	// TODO: cache all the bloc file content ?
	// TODO: call all the index content ?
	// FIXME : which order of idx files to iterate ?
	errChan := make(chan error)
	idxFiles := append(i.deviceIdxFiles, i.otherIdxFiles...)
	p := NewPaginer(i.pageSize, 0, func(push func(State, K, V, error) bool) {
		//panic("not implemented yet")
		for _, bf := range idxFiles {
			for b := range bf.All(filez.BlocOrdering(order), errChan) {
				i.encoder.DecodeAll(order, b.Bytes(), func(seq int, s State, key []byte, val V, err error) {
					if err != nil {
						errChan <- err
						return
					}
					// FIXME: value missing ?
					var k K
					if i.keySerializer != nil {
						k, err = i.keySerializer.Deserialize(key)
					}
					if !push(s, k, val, err) {
						return
					}
				})
			}
		}
	})
	return p, errChan
}

func FixedSizeStringKey(s int, k string) []byte {
	b := make([]byte, s)
	copy(b, []byte(k))
	return b
}
