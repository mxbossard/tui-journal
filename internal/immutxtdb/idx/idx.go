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

type Order string

const (
	TopToBottom Order = "TopToBottom"
	BottomToTop Order = "BottomToTop"
)

type State []byte

var dummyState = BuildState(8, "dummy")

// Return ok=true to select entry, return loop=false to stop iterating.
type StateFilter func(s State) (ok bool, loop bool)

// Return ok=true to select entry, return loop=false to stop iterating.
type KeyFilter func(k []byte, s State) (ok bool, loop bool)
type RotatingHasher func(int, []byte) ([]byte, error)
type KeyRotatingHasher[K comparable] func(int, K) (K, error)

func BuildState(size int, s string) State {
	data := make([]byte, size)
	_, err := binary.Encode(data, binary.BigEndian, []byte(s))
	if err != nil {
		panic(err)
	}
	return State(data)
}

type Index[K comparable, V any] interface {
	Add(s State, key K, val V) error
	Count() (int, error)
	Filter(key K, order Order, sf StateFilter) (Paginer[K, V], chan error)
	FilterAll(order Order, sf StateFilter, kf KeyFilter) (Paginer[K, V], chan error)
	Paginate(key K, order Order) (Paginer[K, V], chan error)
	PaginateAll(order Order) (Paginer[K, V], chan error)
	All(order Order, errChan chan error) iter.Seq2[K, V]
}

type void struct{}
type Void *void

type Entry[K comparable, V any] interface {
	Key() K
	Val() V
	State() State
	Error() error
}

type BasicEntry[K comparable, V any] struct {
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

type basicIndex[K comparable, V any] struct {
	Index[K, V]
	*sync.Mutex
	// FIXME: add a filelock

	keySerializer  serialize.Serializer[K]
	valSerializer  serialize.Serializer[V]
	keyHasher      RotatingHasher
	valHasher      RotatingHasher
	encoder        IdxEncoder // FIXME: encoder must be attached to each BlocsFile or to each Bloc !
	pageSize       int
	filepathes     []string
	deviceIdxFiles []*filez.BlocsFile
	otherIdxFiles  []*filez.BlocsFile
	seqs           map[string]int
}

func NewBasicIndex[K comparable, V any](indexDir, qualifier, device string, keySer serialize.Serializer[K], valSer serialize.Serializer[V], keyH, valH RotatingHasher, enc IdxEncoder, pageSize int) (*basicIndex[K, V], error) {
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
		keySerializer:  keySer,
		valSerializer:  valSer,
		keyHasher:      keyH,
		valHasher:      valH,
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

func (i *basicIndex[K, V]) selectDeviceBlocFile(s State, k K) *filez.BlocsFile {
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
			return fmt.Errorf("error serializing key: %w", err)
		}
	}

	bf := i.selectDeviceBlocFile(s, k)
	bfName := bf.Name()
	seq := i.seqs[bfName]

	// TODO: use 2 optionals RotatingHasher to hash key and value
	val := make([]byte, i.encoder.ValSize())
	err = i.valSerializer.Serialize(v, val)
	if err != nil {
		return fmt.Errorf("error serializing val: %w", err)
	}

	// Rotating Hash
	if i.keyHasher != nil {
		key, err = i.keyHasher(seq, key)
		if err != nil {
			return fmt.Errorf("error hashing key: %w", err)
		}
	}
	if i.valHasher != nil {
		val, err = i.valHasher(seq, val)
		if err != nil {
			return fmt.Errorf("error hashing val: %w", err)
		}
	}

	entry, err := i.encoder.Encode(seq, s, key, val)
	if err != nil {
		return fmt.Errorf("error encoding entry: %w", err)
	}

	//fmt.Printf("writing encoded content (#%d, uid: %s): %v\n", seq, uid, entry)
	_, err = bf.Write(entry)
	if err != nil {
		return fmt.Errorf("error writing entry: %w", err)
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

func (i *basicIndex[K, V]) Filter(suppliedKey K, order Order, sf StateFilter) (Paginer[K, V], chan error) {
	errChan := make(chan error)

	var hashedK []byte
	filteringK := make([]byte, i.encoder.KeySize())
	var err error
	if i.keySerializer != nil {
		err = i.keySerializer.Serialize(suppliedKey, filteringK)
		if err != nil {
			errChan <- err
			return nil, errChan
		}
	}

	idxFiles := append(i.deviceIdxFiles, i.otherIdxFiles...)
	p := NewPaginer(i.pageSize, 0, func(push func(State, K, V, error) bool) {
		//panic("not implemented yet")
	End:
		for _, bf := range idxFiles {
			for b := range bf.All(filez.BlocOrdering(order), errChan) {
				loop := true
				i.encoder.DecodeAll(order, b.Bytes(), func(seq int, s State, key []byte, val []byte, err error) bool {
					if err != nil {
						errChan <- err
						return true
					}

					if sf != nil {
						// If StateFilter does not match ignore the entry
						ok := false
						ok, loop = sf(s)
						if !ok {
							return true
						}
					}
					if i.keyHasher != nil {
						// Rotating Hash
						hashedK, err = i.keyHasher(seq, filteringK)
						if err != nil {
							errChan <- err
							return true
						}
					} else {
						hashedK = filteringK
					}
					// FIXME: do not use serializer if K or V is of []byte type.
					if bytes.Equal(hashedK, key) {
						// if key == filteringKey {
						// FIXME: if key was hashed => cannot be deserialized ! => return nil ?
						var k K
						if i.keySerializer != nil {
							k, err = i.keySerializer.Deserialize(key)
						}
						var v V
						if i.valSerializer != nil {
							v, err = i.valSerializer.Deserialize(val)
						}
						if !push(s, k, v, err) {
							return false
						}
					}
					return loop

				})
				if !loop {
					// Stop iterating
					goto End
				}
			}
		}
	})
	return p, errChan
}

func (i *basicIndex[K, V]) FilterAll(order Order, sf StateFilter, kf KeyFilter) (Paginer[K, V], chan error) {
	// TODO: cache all the bloc file content ?
	// TODO: call all the index content ?
	// FIXME : which order of idx files to iterate ?
	errChan := make(chan error)
	idxFiles := append(i.deviceIdxFiles, i.otherIdxFiles...)
	p := NewPaginer(i.pageSize, 0, func(push func(State, K, V, error) bool) {
		//panic("not implemented yet")
	End:
		for _, bf := range idxFiles {
			for b := range bf.All(filez.BlocOrdering(order), errChan) {
				loop := true
				i.encoder.DecodeAll(order, b.Bytes(), func(seq int, s State, key []byte, val []byte, err error) bool {
					if err != nil {
						errChan <- err
						return true
					}

					if sf != nil {
						// If StateFilter does not match ignore the entry
						ok := false
						ok, loop = sf(s)
						if !ok {
							return true
						}
					}
					if kf != nil {
						ok := false
						ok, loop = kf(key, s)
						if !ok {
							return true
						}
					}

					// FIXME: do not use serializer if K or V is of []byte type.
					var k K
					if i.keySerializer != nil {
						k, err = i.keySerializer.Deserialize(key)
					}
					var v V
					if i.valSerializer != nil {
						v, err = i.valSerializer.Deserialize(val)
					}
					if !push(s, k, v, err) {
						return false
					}
					return loop
				})

				if !loop {
					// Stop iterating
					break End
				}
			}
		}
	})
	return p, errChan
}

func (i *basicIndex[K, V]) Paginate(key K, order Order) (Paginer[K, V], chan error) {
	return i.Filter(key, order, nil)
}

func (i *basicIndex[K, V]) PaginateAll(order Order) (Paginer[K, V], chan error) {
	return i.FilterAll(order, nil, nil)
}

func FixedSizeString(s int, k string) []byte {
	b := make([]byte, s)
	n := copy(b, []byte(k))
	if n > s {
		panic(fmt.Sprintf("string too long for fixed size: %d", s))
	}
	return b
}

func FixedSizeByteSlice(s int, k []byte) []byte {
	if len(k) == s {
		return k
	}
	b := make([]byte, s)
	n := copy(b, k)
	if n > s {
		panic(fmt.Sprintf("byte slice too long for fixed size: %d", s))
	}
	return b
}
