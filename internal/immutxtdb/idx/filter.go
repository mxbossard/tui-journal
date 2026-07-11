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
type seqFilter func(pos int, o Order) (ok bool, loop bool)

type Filter interface {
	StateFilter() stateFilter
	TimeFilter() timeFilter
	SeqFilter() seqFilter
	KeyFilter() KeyFilter
}

type KeyFilter interface {
	MatchKeyFilter() matchKeyFilter
	ExactKeyFilter() *initedKeyFilter
}

type basicKeyFilter struct {
	KeyFilter
	matchKeyFilters matchKeyFilter
	exactKeyFilters *initedKeyFilter
}

func (f basicKeyFilter) MatchKeyFilter() matchKeyFilter {
	return f.matchKeyFilters
}

func (f basicKeyFilter) ExactKeyFilter() *initedKeyFilter {
	return f.exactKeyFilters
}

type aggFilter struct {
	Filter
	logicalOr    bool
	stateFilters []stateFilter
	timeFilters  []timeFilter
	seqFilters   []seqFilter
	keyFilter    KeyFilter
}

func (f *aggFilter) Add(filters ...Filter) *aggFilter {
	for _, filter := range filters {
		if filter != nil {
			f.AddStateFilter(filter.StateFilter())
			f.AddTimeFilter(filter.TimeFilter())
			f.AddSeqFilter(filter.SeqFilter())
			if filter.KeyFilter() != nil {
				// Keep only last KeyFilter
				f.keyFilter = filter.KeyFilter()
			}
		}
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

func (f *aggFilter) SetKeyFilter(filter KeyFilter) *aggFilter {
	f.keyFilter = filter
	return f
}

func (f aggFilter) KeyFilter() KeyFilter {
	return f.keyFilter
}

func NewFilter() *aggFilter {
	f := &aggFilter{logicalOr: false}
	return f
}

func DevelopFilter(filter Filter) *aggFilter {
	f := &aggFilter{logicalOr: false}
	if filter != nil {
		f.AddSeqFilter(filter.SeqFilter())
		f.AddStateFilter(filter.StateFilter())
		f.AddTimeFilter(filter.TimeFilter())
		f.SetKeyFilter(filter.KeyFilter())
	}
	return f
}

func AndFilter(filters ...*aggFilter) *aggFilter {
	filter := &aggFilter{logicalOr: false}
	for _, f := range filters {
		filter.Add(f)
	}
	return filter
}

func OrFilter(filters ...*aggFilter) *aggFilter {
	filter := &aggFilter{logicalOr: true}
	for _, f := range filters {
		filter.Add(f)
	}
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

func IsStateFilter(state State, stopAtFirst bool) *aggFilter {
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

func MatchKeyFilter[K comparable](mkf matchKeyFilter) *basicKeyFilter {
	filter := &basicKeyFilter{matchKeyFilters: mkf}
	return filter
}

type keyFilter struct {
	stopAtFirstMatch  bool
	hashedKeySupplied bool // true if filtering on hashed keys
	bytesKeys         [][]byte
	// keys              []K
	matcher matchKeyFilter
}

func (f keyFilter) match(decodedKey []byte, state State) (bool, bool) {
	return f.matcher(decodedKey, state)
}

type initedKeyFilter struct {
	keyFilter
	// encoder KeyEncoder[K]
	hasher GlidingHasher
	// ekf     exactKeyFilter
}

// Init the filter
func (f *initedKeyFilter) init(hasher GlidingHasher) {
	// f.encoder = encoder
	f.hasher = hasher
}

func (f initedKeyFilter) isExactly(pos int, decodedKey []byte) (bool, bool, error) {
	// build f.byteKeys if nil & panic if keys is nil
	if f.hashedKeySupplied {
		// Supplied keys are hashed
		// Panic if no hasher supplied => Idx is not hashing keys
		if f.hasher == nil {
			panic("cannot use hashed keys, idx does not use GlidingHash")
		}
		// compare each byteKey to decodedKey
		for _, bk := range f.bytesKeys {
			if bytes.Equal(decodedKey, bk) {
				return true, !f.stopAtFirstMatch, nil
			}
		}
	} else {
		// Supplied keys are not hashed
		// if hasher supplied => use hasher to hash each byteKey and compare it to decodedKey
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

func MatchPlainBytesKeysFilter(m matchKeyFilter) matchKeyFilter {
	// idx.filter must call mpbkf.match(decodeKey, state)
	// Panic if key is hashed
	kf := keyFilter{
		stopAtFirstMatch:  false,
		hashedKeySupplied: false,
		matcher:           m,
	}
	return func(decodedKey []byte, s State) (ok bool, loop bool) {
		return kf.match(decodedKey, s)
	}
}

// Filter using bytes key in Index.
// In case of key rotating hashed supplied key will be rotating hashed.
func ExactlyPlainBytesKeysFilter(stopAtFirstMatch bool, keys ...[]byte) *initedKeyFilter {
	// aggFilter must store all keys and compare each decoded key to supplied keys.
	// IN idx.filter epbkf.isExactly(pos, decodedKey)
	// SHOULD panic if keys are hashed
	ikf := initedKeyFilter{
		keyFilter: keyFilter{
			stopAtFirstMatch:  stopAtFirstMatch,
			hashedKeySupplied: false,
			bytesKeys:         keys,
		},
	}
	return &ikf
}

func ExactlyHashedBytesKeysFilter(stopAtFirstMatch bool, keys ...[]byte) *keyFilter {
	// aggFilter must store all keys and compare each decoded key to supplied keys.
	// SHOULD panic if keys are not hashed
	kf := keyFilter{
		stopAtFirstMatch:  stopAtFirstMatch,
		hashedKeySupplied: true,
		bytesKeys:         keys,
	}
	return &kf
}
