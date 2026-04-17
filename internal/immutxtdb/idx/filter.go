package idx

import (
	"bytes"
	"time"
)

// Return ok=true to select entry, return loop=false to stop iterating.
type stateFilter func(s State) (ok bool, loop bool)

// Return ok=true to select entry, return loop=false to stop iterating.
type timeFilter func(t time.Time) (ok bool, loop bool)

// Return ok=true to select entry, return loop=false to stop iterating.
type matchKeyFilter func(decodedKey []byte, s State) (ok bool, loop bool)

// Return ok=true to select entry, return loop=false to stop iterating.
type exactKeyFilter func(pos int, k []byte) (ok bool, loop bool)

// Return ok=true to select entry, return loop=false to stop iterating.
type seqFilter func(pos int, o Order) (ok bool, loop bool)

type Filter interface {
	StateFilter() stateFilter
	TimeFilter() timeFilter
	SeqFilter() seqFilter
}

type KeyFilter[K comparable] interface {
	MatchKeyFilter() matchKeyFilter
	ExactKeyFilter() initedKeyFilter[K]
	// KeyFilter() *keyFilter[any]
}

type basicKeyFilter[K comparable] struct {
	KeyFilter[K]
	matchKeyFilters matchKeyFilter
	exactKeyFilters *initedKeyFilter[K]
}

func (f basicKeyFilter[K]) MatchKeyFilter() matchKeyFilter {
	return f.matchKeyFilters
}

func (f basicKeyFilter[K]) ExactKeyFilter() *initedKeyFilter[K] {
	return f.exactKeyFilters
}

type aggFilter struct {
	Filter
	logicalOr    bool
	stateFilters []stateFilter
	timeFilters  []timeFilter
	seqFilters   []seqFilter
}

func (f *aggFilter) Add(filters ...*aggFilter) *aggFilter {
	for _, filter := range filters {
		f.AddStateFilter(filter.StateFilter())
		f.AddTimeFilter(filter.TimeFilter())
		f.AddSeqFilter(filter.SeqFilter())
	}
	return f
}

func (f *aggFilter) AddStateFilter(filter stateFilter) *aggFilter {
	if filter != nil {
		f.stateFilters = append(f.stateFilters, filter)
	}
	return f
}

func (f aggFilter) StateFilter() stateFilter {
	if len(f.stateFilters) == 0 {
		return nil
	}
	return func(s State) (ok bool, loop bool) {
		ok = !f.logicalOr
		loop = true
		for _, filter := range f.stateFilters {
			sok, sloop := filter(s)
			if f.logicalOr {
				ok = ok || sok
			} else {
				ok = ok && sok
			}
			loop = loop && sloop

		}
		return
	}
}

func (f *aggFilter) AddTimeFilter(filter timeFilter) *aggFilter {
	if filter != nil {
		f.timeFilters = append(f.timeFilters, filter)
	}
	return f
}

func (f aggFilter) TimeFilter() timeFilter {
	if len(f.timeFilters) == 0 {
		return nil
	}
	return func(t time.Time) (ok bool, loop bool) {
		ok = !f.logicalOr
		loop = true
		for _, filter := range f.timeFilters {
			sok, sloop := filter(t)
			if f.logicalOr {
				ok = ok || sok
			} else {
				ok = ok && sok
			}
			loop = loop && sloop

		}
		return
	}
}

func (f *aggFilter) AddSeqFilter(filter seqFilter) *aggFilter {
	if filter != nil {
		f.seqFilters = append(f.seqFilters, filter)
	}
	return f
}

func (f aggFilter) SeqFilter() seqFilter {
	if len(f.seqFilters) == 0 {
		return nil
	}
	return func(seq int, o Order) (ok bool, loop bool) {
		ok = !f.logicalOr
		loop = true
		for _, filter := range f.seqFilters {
			sok, sloop := filter(seq, o)
			if f.logicalOr {
				ok = ok || sok
			} else {
				ok = ok && sok
			}
			loop = loop && sloop

		}
		return
	}
}

func AndFilter(filters ...*aggFilter) *aggFilter {
	filter := &aggFilter{logicalOr: false}
	filter.Add(filters...)
	return filter
}

func OrFilter(filters ...*aggFilter) *aggFilter {
	filter := &aggFilter{logicalOr: true}
	filter.Add(filters...)
	return filter
}

func BeforeFilter(b time.Time) *aggFilter {
	f := &aggFilter{}
	f.AddTimeFilter(func(t time.Time) (ok bool, loop bool) {
		return t.Before(b), true
	})
	return f
}

func AfterFilter(a time.Time) *aggFilter {
	f := &aggFilter{}
	f.AddTimeFilter(func(t time.Time) (ok bool, loop bool) {
		return t.After(a), true
	})
	return f
}

func BetweenFilter(a, b time.Time) *aggFilter {
	if !a.Before(b) {
		panic("a must be before b")
	}
	f := AndFilter(BeforeFilter(b), AfterFilter(a))
	return f
}

func BeforeSeqFilter(seq int) *aggFilter {
	f := &aggFilter{}
	f.AddSeqFilter(func(s int, o Order) (ok bool, loop bool) {
		ok = s < seq
		switch o {
		case TopToBottom:
			loop = ok
		case BottomToTop:
			loop = true
		}
		return
	})
	return f
}

func AfterSeqFilter(seq int) *aggFilter {
	f := &aggFilter{}
	f.AddSeqFilter(func(s int, o Order) (ok bool, loop bool) {
		ok = s > seq
		switch o {
		case TopToBottom:
			loop = true
		case BottomToTop:
			loop = ok
		}
		return
	})
	return f
}

func BetweenSeqFilter(a, b int) *aggFilter {
	if a >= b {
		panic("a must be before b")
	}
	f := AndFilter(BeforeSeqFilter(b), AfterSeqFilter(a))
	return f
}

func TimeFilter(tf timeFilter) *aggFilter {
	filter := &aggFilter{logicalOr: false}
	if tf != nil {
		filter.AddTimeFilter(tf)
	}
	return filter
}

func StateFilter(sf stateFilter) *aggFilter {
	filter := &aggFilter{logicalOr: false}
	if sf != nil {
		filter.AddStateFilter(sf)
	}
	return filter
}

func MatchStateFilter(state State, stopAtFirst bool) *aggFilter {
	return StateFilter(func(s State) (ok bool, loop bool) {
		ok = bytes.Equal(s, state)
		loop = true
		if stopAtFirst {
			loop = !ok
		} else {
			loop = true
		}
		return
	})
}

func MatchKeyFilter(mkf matchKeyFilter) *basicKeyFilter[Void] {
	filter := &basicKeyFilter[Void]{matchKeyFilters: mkf}
	return filter
}

func ExactKeyFilter[K comparable](ekf *initedKeyFilter[K]) *basicKeyFilter[K] {
	filter := &basicKeyFilter[K]{exactKeyFilters: ekf}
	return filter
}

type keyFilter[K comparable] struct {
	stopAtFirstMatch  bool
	hashedKeySupplied bool // true if filtering on hashed keys
	bytesKeys         [][]byte
	keys              []K
	matcher           matchKeyFilter
}

func (f keyFilter[K]) match(decodedKey []byte, state State) (bool, bool) {
	return f.matcher(decodedKey, state)
}

type initedKeyFilter[K comparable] struct {
	keyFilter[K]
	encoder KeyEncoder[K]
	hasher  GlidingHasher
	// ekf     exactKeyFilter
}

// Init the filter
func (f *initedKeyFilter[K]) init(encoder KeyEncoder[K], hasher GlidingHasher) {
	f.encoder = encoder
	f.hasher = hasher
}

func (f initedKeyFilter[K]) isExactly(pos int, decodedKey []byte) (bool, bool, error) {
	// TODO: build f.byteKeys if nil & panic if keys is nil
	// FIXME: bytesKeys & keys could be supplied both need to build all bytesKeys at least once
	panic("need to implement this ^^^")
	if len(f.bytesKeys) == 0 {
		if len(f.keys) == 0 {
			panic("neither keys or bytesKeys supplied to keyFilter")
		}
		// Build all bytesKeys from supplied keys
		for _, key := range f.keys {
			bk, err := f.encoder(key)
			if err != nil {
				return false, true, err
			}
			f.bytesKeys = append(f.bytesKeys, bk)
		}
	}

	if f.hashedKeySupplied {
		// Supplied keys are hashed
		// TODO: Panic if no hasher supplied => Idx is not hashing keys
		if f.hasher == nil {
			panic("cannot use hashed keys, idx does not use GlidingHash")
		}
		// TODO: compare each byteKey to decodedKey
		for _, bk := range f.bytesKeys {
			if bytes.Equal(decodedKey, bk) {
				return true, !f.stopAtFirstMatch, nil
			}
		}
	} else {
		// Supplied keys are not hashed
		// TODO: if hasher supplied => use hasher to hash each byteKey and compare it to decodedKey
		var err error
		for _, bk := range f.bytesKeys {
			if f.hasher != nil {
				bk, err = f.hasher(pos, bk)
				if err != nil {
					return false, true, err
				}
			}
			if bytes.Equal(decodedKey, bk) {
				return true, !f.stopAtFirstMatch, nil
			}
		}

	}
	return false, true, nil
}

// Filter using bytes key in Index.
// In case of key rotating hashed supplied key will be rotating hashed.
func ExactlyPlainBytesKeysFilter(stopAtFirstMatch bool, keys ...[]byte) *initedKeyFilter[Void] {
	// TODO: aggFilter must store all keys and compare each decoded key to supplied keys.
	// IN idx.filter epbkf.isExactly(pos, decodedKey)
	// SHOULD panic if keys are hashed
	ikf := initedKeyFilter[Void]{
		keyFilter: keyFilter[Void]{
			stopAtFirstMatch:  stopAtFirstMatch,
			hashedKeySupplied: false,
			bytesKeys:         keys,
		},
	}
	return &ikf
	panic("not implemented yet")
}

func ExactlyHashedBytesKeysFilter(stopAtFirstMatch bool, keys ...[]byte) *keyFilter[Void] {
	// TODO: aggFilter must store all keys and compare each decoded key to supplied keys.
	// SHOULD panic if keys are not hashed
	kf := keyFilter[Void]{
		stopAtFirstMatch:  stopAtFirstMatch,
		hashedKeySupplied: true,
		bytesKeys:         keys,
	}
	panic("not implemented yet")
	return &kf

	// return MatchKeyFilter(func(k []byte, s State) (ok bool, loop bool) {
	// 	fmt.Printf("ExactlyHashedBytesKeysFilter: comparing %v with %v ...\n", k, key)
	// 	ok = bytes.Equal(k, key)
	// 	if stopAtFirst {
	// 		loop = !ok
	// 	} else {
	// 		loop = true
	// 	}
	// 	return
	// })
}

func KeysFilter[K comparable](stopAtFirstMatch bool, keys ...K) *keyFilter[K] {
	// TODO: CAN we use ExactlyPlainBytesKeysFilter to converting keys ?
	// Probably not, so we should store keys and let the idx.filter converting keys
	// Should works if keys are hashed or not
	kf := keyFilter[K]{
		stopAtFirstMatch:  stopAtFirstMatch,
		hashedKeySupplied: false,
		keys:              keys,
	}
	panic("not implemented yet")
	return &kf
}

func MatchPlainBytesKeysFilter(m matchKeyFilter) matchKeyFilter {
	// TODO: idx.filter must call mpbkf.match(decodeKey, state)
	// Panic if key is hashed
	kf := keyFilter[Void]{
		stopAtFirstMatch:  false,
		hashedKeySupplied: false,
		matcher:           m,
	}
	return func(decodedKey []byte, s State) (ok bool, loop bool) {
		return kf.match(decodedKey, s)
	}
	panic("not implemented yet")
}

type bar[K comparable] struct {
	k K
}

func foo[K comparable](b bar[K]) bar[K] {
	return bar[K]{}
}

func foo2() bar[int] {
	return bar[int]{}
}

func baz() {

	foo[string]()
}
