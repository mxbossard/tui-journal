package idx

import (
	"bytes"
	"fmt"
	"iter"
	"path/filepath"
	"sync"
	"time"

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

const (
	TopToBottom Order = "TopToBottom"
	BottomToTop Order = "BottomToTop"
)

type Order string

type void struct{}
type Void *void

type RotatingHasher func(int, []byte) ([]byte, error)
type KeyRotatingHasher[K comparable] func(int, K) (K, error)

// FIXME: Paginer SHOULD return KV Entries also embedding seq, time and state ?
type Index[K comparable, V any] interface {
	// Add a KV entry
	Add(s State, t time.Time, key K, val V) (Entry[K, V], error)
	// Return last entry seq
	LastSeq() (int, error)

	// ----- Browsing Methods -----
	// Paginate all KV entries matching supplied key & Filter
	Filter(key K, order Order, f Filter) (Paginer[K, V], error)
	// Paginate all KV entries matching supplied key & Filter
	HashedFilter(key K, order Order, f Filter) (Paginer[K, V], error)
	// Paginate all KV entries matching supplied Filter
	FilterAll(order Order, f Filter) (Paginer[K, V], error)
	// Paginate all KV entries exactly matching supplied key
	Paginate(key K, order Order) (Paginer[K, V], error)
	// Paginate all KV entries matching supplied key which will be rotating hashed
	HashedPaginate(key K, order Order) (Paginer[K, V], error)
	// Paginate all KV entries
	PaginateAll(order Order) (Paginer[K, V], error)
	// Return an iterator of all KV entries
	All(order Order) (iter.Seq[Entry[K, V]], error)
}

type basicIndex[K comparable, V any] struct {
	Index[K, V]
	*sync.Mutex
	// FIXME: add a filelock

	keySerializer    serialize.Serializer[K]
	valSerializer    serialize.Serializer[V]
	keyHasher        RotatingHasher
	valHasher        RotatingHasher
	encoder          IdxEncoder // FIXME: encoder must be attached to each BlocsFile or to each Bloc !
	pageSize         int
	preloadPageCount int
	filepathes       []string
	deviceIdxFiles   []*filez.BlocsFile
	otherIdxFiles    []*filez.BlocsFile
	seqs             map[string]int
}

func NewBasicIndex[K comparable, V any](indexDir, qualifier, device string, keySer serialize.Serializer[K], valSer serialize.Serializer[V], keyH, valH RotatingHasher, enc IdxEncoder, pageSize int) (*basicIndex[K, V], error) {
	// Init bucketIndex
	// FIXME: manage multiple idx files (rotation)
	// FIXME: add a filelock
	// FIXME: addRotatingHash ?
	firstDeviceFilepath := filepath.Join(indexDir, fmt.Sprintf("%s-%s-001.idx", qualifier, device))
	dbf1, err := filez.NewBlocsFile(firstDeviceFilepath, 256, 100)
	if err != nil {
		return nil, fmt.Errorf("unable to build blocs file: %w", err)
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

	// TODO: need to load idx.seqs !
	for _, bf := range idx.deviceIdxFiles {
		// Decode last line of last bloc to get current seq
		bloc, err := bf.GetLastNonEmptyBloc()
		if err == filez.ErrNotExist {
			// No bloc to read
			err = nil
			continue
		} else if err != nil {
			return nil, fmt.Errorf("unable to get last bloc: %w", err)
		}
		if bloc.Len() > 0 {
			lastSeq, _, _, _, _, err := enc.DecodeLastWord(bloc.Bytes())
			if err != nil {
				return nil, fmt.Errorf("unable to decode last word: %w", err)
			}
			idx.seqs[bf.Name()] = lastSeq + 1
		}
	}

	return idx, nil
}

func (i *basicIndex[K, V]) selectDeviceBlocFile(s State, k K) *filez.BlocsFile {
	return i.deviceIdxFiles[0]
}

func (i *basicIndex[K, V]) Add(s State, t time.Time, k K, v V) (Entry[K, V], error) {
	i.Lock()
	defer i.Unlock()

	truncatedTime := t.Truncate(24 * time.Hour)
	normalizedState := CatState(i.encoder.StateSize(), s)

	var err error
	var ok bool
	var key []byte
	if i.keySerializer != nil {
		key = make([]byte, i.encoder.KeySize())
		_, err = i.keySerializer.Serialize(k, key)
		if err != nil {
			return nil, fmt.Errorf("error serializing key: %w", err)
		}
	} else if key, ok = any(k).([]byte); !ok {
		panic("cannot convert key to []byte, need a keySerializer")
	}

	bf := i.selectDeviceBlocFile(normalizedState, k)
	bfName := bf.Name()
	seq := i.seqs[bfName]

	var val []byte
	if i.valSerializer != nil {
		val = make([]byte, i.encoder.ValSize())
		_, err = i.valSerializer.Serialize(v, val)
		if err != nil {
			return nil, fmt.Errorf("error serializing val: %w", err)
		}
	} else if val, ok = any(v).([]byte); !ok {
		panic("cannot convert val to []byte, need a valSerializer")
	}

	// Rotating Hash
	if i.keyHasher != nil {
		key, err = i.keyHasher(seq, key)
		if err != nil {
			return nil, fmt.Errorf("error hashing key: %w", err)
		}
	}
	if i.valHasher != nil {
		val, err = i.valHasher(seq, val)
		if err != nil {
			return nil, fmt.Errorf("error hashing val: %w", err)
		}
	}

	data, err := i.encoder.Encode(seq, truncatedTime, normalizedState, key, val)
	if err != nil {
		return nil, fmt.Errorf("error encoding data: %w", err)
	}

	//fmt.Printf("writing encoded content (#%d, uid: %s): %v\n", seq, uid, entry)
	bw := bf.Writer()
	_, err = bw.Write(data)
	if err != nil {
		return nil, fmt.Errorf("error writing data: %w", err)
	}
	i.seqs[bfName] = seq + 1
	return NewEntry(k, v, seq, truncatedTime, normalizedState, nil, key), nil
}

func (i *basicIndex[K, V]) LastSeq() (int, error) {
	// Should be performent and not read all the index to count lines.
	i.Lock()
	defer i.Unlock()
	count := 0
	for _, k := range i.seqs {
		count += k
	}
	return count, nil
}

func (i *basicIndex[K, V]) filter(suppliedKey K, keyFiltering, hashedKey bool, order Order, f Filter) (Paginer[K, V], error) {
	// TODO: cache all the bloc file content ?
	// TODO: call all the index content ?
	// FIXME : which order of idx files to iterate ?

	var hashedK []byte
	var filteringK []byte
	if keyFiltering {
		filteringK = make([]byte, i.encoder.KeySize())
		var err error
		if i.keySerializer != nil {
			_, err = i.keySerializer.Serialize(suppliedKey, filteringK)
			if err != nil {
				return nil, err
			}
		}
	}

	idxFiles := append(i.deviceIdxFiles, i.otherIdxFiles...)
	p := NewPaginer(i.pageSize, i.preloadPageCount, func(push func(Entry[K, V]) bool) {
		// pusher func impl

	End:
		for _, bf := range idxFiles {
			for err, b := range bf.All(filez.BlocOrdering(order)) {
				if err != nil {
					e := NewErrEntry[K, V](err)
					if !push(e) {
						return
					}
				}
				loop := true
				i.encoder.DecodeAll(order, b.Bytes(), func(seq int, t time.Time, s State, key []byte, val []byte, err error) bool {
					// callback func impl
					if err != nil {
						// decoding err => we want to push it and keep iterating
						e := NewErrEntry[K, V](err)
						if !push(e) {
							// we want to stop iterating and then stop decoding
							return false
						}
						return true
					}

					if f != nil {
						sf := f.StateFilter()
						if sf != nil {
							// If StateFilter does not match ignore the entry
							ok, iloop := sf(s)
							if !ok {
								return true
							}
							loop = loop && iloop
						}
						tf := f.TimeFilter()
						if tf != nil {
							// If TimeFilter does not match ignore the entry
							ok, iloop := tf(t)
							if !ok {
								return true
							}
							loop = loop && iloop
						}
						kf := f.KeyFilter()
						if kf != nil {
							// If KeyFilter does not match ignore the entry
							ok, iloop := kf(key, s)
							if !ok {
								return true
							}
							loop = loop && iloop
						}
					}

					if keyFiltering {
						if hashedKey && i.keyHasher != nil {
							// Rotating Hash
							hashedK, err = i.keyHasher(seq, filteringK)
							if err != nil {
								// hashing err => we want to push it and keep iterating
								e := NewErrEntry[K, V](err)
								if !push(e) {
									// we want to stop iterating and then stop decoding
									return false
								}
								return true
							}
						} else {
							hashedK = filteringK
						}
					}

					// FIXME: do not use serializer if K or V is of []byte type.
					if !keyFiltering || bytes.Equal(hashedK, key) {
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
						e := NewEntry(k, v, seq, t, s, err, hashedK)
						if !push(e) {
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
	return p, nil
}

func (i *basicIndex[K, V]) Filter(suppliedKey K, order Order, f Filter) (Paginer[K, V], error) {
	return i.filter(suppliedKey, true, false, order, f)
}

func (i *basicIndex[K, V]) HashedFilter(suppliedKey K, order Order, f Filter) (Paginer[K, V], error) {
	return i.filter(suppliedKey, true, true, order, f)
}

func (i *basicIndex[K, V]) FilterAll(order Order, f Filter) (Paginer[K, V], error) {
	var noKey K
	return i.filter(noKey, false, false, order, f)
}

func (i *basicIndex[K, V]) Paginate(key K, order Order) (Paginer[K, V], error) {
	return i.filter(key, true, false, order, nil)
}

func (i *basicIndex[K, V]) HashedPaginate(key K, order Order) (Paginer[K, V], error) {
	return i.filter(key, true, true, order, nil)
}

func (i *basicIndex[K, V]) PaginateAll(order Order) (Paginer[K, V], error) {
	var noKey K
	return i.filter(noKey, false, false, order, nil)
}

func (i *basicIndex[K, V]) All(order Order) (iter.Seq[Entry[K, V]], error) {
	paginer, err := i.PaginateAll(order)
	if err != nil {
		return nil, err
	}
	return paginer.All(), nil
}
