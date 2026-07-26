package idx

import (
	"fmt"
	"iter"
	"sync"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx/idxcfg"
	idxhash "github.com/mxbossard/tui-journal/internal/immutxtdb/idx/idxhash"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx/idxrepo"
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

type Order = idxrepo.Order

const (
	TopToBottom Order = idxrepo.TopToBottom
	BottomToTop Order = idxrepo.BottomToTop
)

type void struct{}
type Void *void

// FIXME: Paginer SHOULD return KV Entries also embedding seq, time and state ?
type Index[K comparable, V any] interface {
	// Add a KV entry
	Add(s State, t time.Time, key K, val V) (Entry[K, V], error)
	// Return last entry seq
	LastSeq() (int, error)

	// ----- Browsing Methods -----
	// Paginate all KV entries matching supplied key & Filter
	Filter(key K, order Order, f Filter) (Paginer[K, V], error)
	// Paginate all KV entries matching supplied hashed key & Filter
	// HashedFilter(key K, order Order, f Filter) (Paginer[K, V], error)
	// Paginate all KV entries matching supplied Filter
	FilterAll(order Order, f Filter) (Paginer[K, V], error)
	// Paginate all KV entries exactly matching supplied key
	Paginate(key K, order Order) (Paginer[K, V], error)
	// Paginate all KV entries matching supplied hashed key which will be rotating hashed
	// HashedPaginate(key K, order Order) (Paginer[K, V], error)
	// Paginate all KV entries
	PaginateAll(order Order) (Paginer[K, V], error)
	// Return an iterator of all KV entries
	All(order Order) (iter.Seq[Entry[K, V]], error)

	// Build an ExactKeyFilter
	KeysFilter(stopAtFirstMatch bool, keys ...K) (*aggFilter, error)
}

type basicIndex[K comparable, V any] struct {
	Index[K, V]
	*sync.Mutex
	// FIXME: add a filelock

	cfg  idxcfg.Config[K, V]
	repo idxrepo.BlocsRepo

	// FIXME: encoder must be attached to each BlocsFile or to each Bloc !
	// encoder IdxEncoder // encode an entry into []byte ready to store & vice versa
	// partitionIdxFiles []*filez.BlocsFile
	// otherIdxFiles     []*filez.BlocsFile

	seqs map[string]int
}

// Create an Index with supplied config.
func NewBasicIndex[K comparable, V any](indexDir string, cfg idxcfg.Config[K, V]) (*basicIndex[K, V], error) {
	// FIXME: Move all files and encoder management in "repo" struct
	// FIXME: manage multiple idx files (rotation)
	// FIXME: add a filelock

	// firstPartitionFilepath := filepath.Join(indexDir, fmt.Sprintf("%s-%s-001.idx", cfg.name, cfg.partition))
	// dbf1, err := filez.NewBlocsFile(firstPartitionFilepath, 256, 100)
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to build blocs file: %w", err)
	// }
	// enc := NewByteSliceEncoder(0, cfg.stateSize, cfg.keySize, cfg.valSize)
	blocsRepo, err := idxrepo.DefaultBlocsRepo[K, V](indexDir, cfg)
	if err != nil {
		return nil, err
	}

	idx := &basicIndex[K, V]{
		Mutex: &sync.Mutex{},
		cfg:   cfg,
		repo:  blocsRepo,
		// encoder: enc,

		// partitionIdxFiles: []*filez.BlocsFile{dbf1},
		// otherIdxFiles:     nil,
		// seqs:              make(map[string]int),
	}

	// FIXME: need to setup the encoder!
	//e.Setup()

	// for _, bf := range idx.partitionIdxFiles {
	// 	// Decode last line of last bloc to get current seq
	// 	bloc, err := bf.GetLastNonEmptyBloc()
	// 	if err == filez.ErrNotExist {
	// 		// No bloc to read
	// 		err = nil
	// 		continue
	// 	} else if err != nil {
	// 		return nil, fmt.Errorf("unable to get last bloc: %w", err)
	// 	}
	// 	if bloc.Len() > 0 {
	// 		lastSeq, _, _, _, _, err := enc.DecodeLastWord(bloc.Bytes())
	// 		if err != nil {
	// 			return nil, fmt.Errorf("unable to decode last word: %w", err)
	// 		}
	// 		idx.seqs[bf.Name()] = lastSeq + 1
	// 	}
	// }

	idx.seqs, err = blocsRepo.LoadSeqs()
	return idx, err
}

// Create an Index with idx.DefaultConfig().
// Name should be a functionnal name
// Partition should be a technical qualifier (like a device)
func NewDefaultIndex[K comparable, V any](indexDir, name, partition string) (*basicIndex[K, V], error) {
	cfg := DefaultConfig[K, V](name, partition)
	return NewBasicIndex(indexDir, cfg)
}

func NewBasicIndex0[K comparable, V any](indexDir, name, partition string,
	keySer serialize.Serializer[K], valSer serialize.Serializer[V],
	salt []byte, keyH, valH idxhash.GlidingHasher, enc idxrepo.IdxEncoder,
	pageSize, preloadPageCount int) (*basicIndex[K, V], error) {
	// Init bucketIndex
	// FIXME: manage multiple idx files (rotation)
	// FIXME: add a filelock
	// FIXME: addRotatingHash ?

	// firstPartitionFilepath := filepath.Join(indexDir, fmt.Sprintf("%s-%s-001.idx", name, partition))
	// dbf1, err := filez.NewBlocsFile(firstPartitionFilepath, 256, 100)
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to build blocs file: %w", err)
	// }

	cfg := DefaultConfig[K, V](name, partition)
	cfg.SetStateSize(enc.StateSize())
	cfg.SetKeySize(enc.KeySize())
	cfg.SetValSize(enc.ValSize())
	cfg.SetPageSize(pageSize)
	cfg.SetPreloadPageCount(preloadPageCount)
	if keySer != nil {
		cfg.SetKeySerializer0(keySer)
	} else {
		cfg.SetKeySerializer(nil)
	}
	if valSer != nil {
		cfg.SetValSerializer0(valSer)
	} else {
		cfg.SetValSerializer(nil)
	}
	if keyH != nil {
		cfg.EnableKeyHasher(salt)
	}
	if valH != nil {
		cfg.EnableValHasher(salt)
	}

	r, err := idxrepo.DefaultBlocsRepo[K, V](indexDir, cfg)
	if err != nil {
		return nil, err
	}

	idx := &basicIndex[K, V]{
		Mutex: &sync.Mutex{},
		cfg:   cfg,

		// repo: repo.BlocsRepo{
		// 	indexDir:          indexDir,
		// 	encoder:           enc,
		// 	partitionIdxFiles: []*filez.BlocsFile{dbf1},
		// },
		repo: r,

		seqs: make(map[string]int),
	}

	// FIXME: need to setup the encoder!
	//e.Setup()

	// TODO: need to load idx.seqs !
	for _, bf := range r.BlocsFiles() {
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

func (i *basicIndex[K, V]) Add(s State, t time.Time, k K, v V) (Entry[K, V], error) {
	i.Lock()
	defer i.Unlock()

	truncatedTime := t.Truncate(24 * time.Hour)
	normalizedState := CatState(i.cfg.StateSize, s)

	var err error
	var ok bool
	var key []byte
	if _, ok = any(k).(Void); ok {
		key = make([]byte, i.cfg.KeySize)
	} else if i.cfg.KeySerializer() != nil {
		key = make([]byte, i.cfg.KeySize)
		_, err := i.cfg.KeySerializer().Serialize(k, &key)
		if err != nil {
			return nil, fmt.Errorf("error serializing key: %w", err)
		}
	} else if key, ok = any(k).([]byte); !ok {
		key = make([]byte, i.cfg.KeySize)
		bs := serialize.BinarySerializer{}
		_, err := bs.Serialize(k, &key)
		if err != nil {
			panic(fmt.Errorf("cannot convert key of type %T to []byte, need a keySerializer (err: %w)", key, err))
		}
	}

	if len(key) > i.cfg.KeySize {
		err = fmt.Errorf("supplied key: %v overflow key size: %d", k, i.cfg.KeySize)
		return nil, err
	} else if len(key) < i.cfg.KeySize {
		// FIXME: SHOULD copy key into right size of []byte
		panic("bad key size")
	}

	// bf := i.selectPartitionBlocFile(normalizedState, k)
	bf := i.repo.SelectPartitionBlocFile(normalizedState, key)
	bfName := bf.Name()
	seq := i.seqs[bfName]

	var val []byte
	if i.cfg.ValSerializer() != nil {
		val = make([]byte, i.cfg.ValSize)
		_, err = i.cfg.ValSerializer().Serialize(v, &val)
		if err != nil {
			return nil, fmt.Errorf("error serializing val: %w", err)
		}
	} else if val, ok = any(v).([]byte); !ok {
		panic("cannot convert val to []byte, need a valSerializer")
	}

	// Rotating Hash
	if i.cfg.KeyHasher() != nil {
		key, err = i.cfg.KeyHasher()(seq, key)
		if err != nil {
			return nil, fmt.Errorf("error hashing key: %w", err)
		}
	}
	if i.cfg.ValHasher() != nil {
		val, err = i.cfg.ValHasher()(seq, val)
		if err != nil {
			return nil, fmt.Errorf("error hashing val: %w", err)
		}
	}

	data, err := i.repo.Encoder().Encode(seq, truncatedTime, normalizedState, key, val)
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

func (i basicIndex[K, V]) LastSeq() (int, error) {
	// Should be performent and not read all the index to count lines.
	i.Lock()
	defer i.Unlock()
	count := 0
	for _, k := range i.seqs {
		count += k
	}
	return count, nil
}

func (i basicIndex[K, V]) filter(order Order, f Filter) (Paginer[K, V], error) {
	p := NewPaginer(i.cfg.PageSize, i.cfg.PreloadPageCount, func(push func(Entry[K, V]) bool) {
		callback := i.filteringDecoderCallback(order, f, push)
		i.repo.ScanDecodeAll(order, callback)
	})

	return p, nil
}

func (i basicIndex[K, V]) filteringDecoderCallback(order Order, f Filter, push func(Entry[K, V]) bool) func(seq int, t time.Time, s State, key []byte, val []byte, err error) bool {

	var sf stateFilter
	var tf timeFilter
	var qf seqFilter
	var kf KeyFilter
	var mkf matchKeyFilter
	var ekf *initedKeyFilter
	if f != nil {
		sf = f.StateFilter()
		tf = f.TimeFilter()
		qf = f.SeqFilter()
		kf = f.KeyFilter()
		if kf != nil {
			mkf = kf.MatchKeyFilter()
			ekf = kf.ExactKeyFilter()
			if ekf != nil {
				ekf.init(i.cfg.KeyHasher())
			}
		}
	}

	return func(seq int, t time.Time, s State, key []byte, val []byte, err error) bool {
		loop := true
		// callback func impl for each decoded line
		if err != nil {
			// decoding err => we want to push it and keep iterating
			e := NewErrEntry[K, V](err)
			if !push(e) {
				// we want to stop iterating and then stop decoding
				return false
			}
			return true
		}

		if sf != nil {
			// If StateFilter does not match ignore the entry
			ok, iloop := sf(s)
			if !ok {
				return true
			}
			loop = loop && iloop
		}
		if tf != nil {
			// If TimeFilter does not match ignore the entry
			ok, iloop := tf(t)
			if !ok {
				return true
			}
			loop = loop && iloop
		}

		matchingKeyFilter := false
		if mkf != nil {
			// If MatchKeyFilter does not match ignore the entry
			ok, iloop := mkf(key, s)
			if !ok {
				return true
			}
			loop = loop && iloop
			matchingKeyFilter = ok
		}

		if ekf != nil {
			// If MatchKeyFilter does not match ignore the entry
			// fmt.Printf("ekf: isExactly? %d %v\n", seq, key)
			ok, iloop, err := ekf.isExactly(seq, key)
			if err != nil {
				// decoding err => we want to push it and keep iterating
				e := NewErrEntry[K, V](err)
				if !push(e) {
					// we want to stop iterating and then stop decoding
					return false
				}
				return true
			}
			if !ok {
				return true
			}
			loop = loop && iloop
			matchingKeyFilter = matchingKeyFilter || ok
		}

		if qf != nil {
			// If SeqFilter does not match ignore the entry
			ok, iloop := qf(seq, order)
			if !ok {
				return true
			}
			loop = loop && iloop
		}

		// FIXME: do not use serializer if K or V is of []byte type.
		if matchingKeyFilter || kf == nil { //|| !keyFiltering || bytes.Equal(hashedK, key) {
			// FIXME: if key was hashed => cannot be deserialized ! => return nil ?
			var k K
			if i.cfg.KeySerializer() != nil {
				k, err = i.cfg.KeySerializer().Deserialize(key)
			}
			var v V
			if i.cfg.ValSerializer() != nil {
				v, err = i.cfg.ValSerializer().Deserialize(val)
			}
			e := NewEntry(k, v, seq, t, s, err, key)
			if !push(e) {
				return false
			}
		}
		return loop
	}
}

func (i basicIndex[K, V]) Filter(suppliedKey K, order Order, f Filter) (Paginer[K, V], error) {
	kf, err := i.KeysFilter(false, suppliedKey)
	if err != nil {
		return nil, err
	}
	kf.Add(f)
	// var noKey K
	// return i.filter(noKey, true, false, order, kf)
	return i.filter(order, kf)
}

// func (i *basicIndex[K, V]) HashedFilter(suppliedKey K, order Order, f Filter) (Paginer[K, V], error) {
// 	kf, err := i.KeysFilter(false, suppliedKey)
// 	if err != nil {
// 		return nil, err
// 	}
// 	kf.Add(f)
// 	// var noKey K
// 	// return i.filter(noKey, true, true, order, kf)
// 	return i.filter(order, kf)
// }

func (i *basicIndex[K, V]) FilterAll(order Order, f Filter) (Paginer[K, V], error) {
	// var noKey K
	// return i.filter(noKey, false, false, order, f)
	return i.filter(order, f)
}

func (i *basicIndex[K, V]) Paginate(key K, order Order) (Paginer[K, V], error) {
	kf, err := i.KeysFilter(false, key)
	if err != nil {
		return nil, err
	}
	// var noKey K
	// return i.filter(noKey, true, false, order, kf)
	return i.filter(order, kf)
}

// func (i *basicIndex[K, V]) HashedPaginate(key K, order Order) (Paginer[K, V], error) {
// 	kf, err := i.KeysFilter(false, key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// var noKey K
// 	// return i.filter(noKey, true, true, order, kf)
// 	return i.filter(order, kf)
// }

func (i *basicIndex[K, V]) PaginateAll(order Order) (Paginer[K, V], error) {
	// var noKey K
	// return i.filter(noKey, false, false, order, nil)
	return i.filter(order, nil)
}

func (i *basicIndex[K, V]) All(order Order) (iter.Seq[Entry[K, V]], error) {
	paginer, err := i.PaginateAll(order)
	if err != nil {
		return nil, err
	}
	return paginer.All(), nil
}

func (i *basicIndex[K, V]) KeysFilter(stopAtFirstMatch bool, keys ...K) (*aggFilter, error) {
	// Build all bytesKeys from supplied keys
	var bytesKeys [][]byte
	for _, key := range keys {
		bk := make([]byte, i.cfg.KeySize)
		_, err := i.cfg.KeySerializer().Serialize(key, &bk)
		if err != nil {
			return nil, err
		}
		bytesKeys = append(bytesKeys, bk)
	}
	// fmt.Printf("serialized keys: %v => %v\n", keys, bytesKeys)
	bkf := basicKeyFilter{
		exactKeyFilters: &initedKeyFilter{
			keyFilter: keyFilter{
				stopAtFirstMatch:  stopAtFirstMatch,
				hashedKeySupplied: false,
				bytesKeys:         bytesKeys,
			},
		},
	}
	f := NewFilter()
	f.SetKeyFilter(bkf)
	return f, nil
}
