package idx

import "time"

// Return ok=true to select entry, return loop=false to stop iterating.
type StateFilter func(s State) (ok bool, loop bool)

// Return ok=true to select entry, return loop=false to stop iterating.
type TimeFilter func(t time.Time) (ok bool, loop bool)

// Return ok=true to select entry, return loop=false to stop iterating.
type KeyFilter func(k []byte, s State) (ok bool, loop bool)

type Filter interface {
	StateFilter() StateFilter
	TimeFilter() TimeFilter
	KeyFilter() KeyFilter
}

type aggFilter struct {
	Filter
	logicalOr    bool
	stateFilters []StateFilter
	timeFilters  []TimeFilter
	keyFilters   []KeyFilter
}

func (f *aggFilter) Add(filters ...*aggFilter) *aggFilter {
	for _, filter := range filters {
		f.AddStateFilter(filter.StateFilter())
		f.AddTimeFilter(filter.TimeFilter())
		f.AddKeyFilter(filter.KeyFilter())
	}
	return f
}

func (f *aggFilter) AddStateFilter(filter StateFilter) *aggFilter {
	f.stateFilters = append(f.stateFilters, filter)
	return f
}

func (f aggFilter) StateFilter() StateFilter {
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

func (f *aggFilter) AddTimeFilter(filter TimeFilter) *aggFilter {
	f.timeFilters = append(f.timeFilters, filter)
	return f
}

func (f aggFilter) TimeFilter() TimeFilter {
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

func (f *aggFilter) AddKeyFilter(filter KeyFilter) *aggFilter {
	f.keyFilters = append(f.keyFilters, filter)
	return f
}

func (f aggFilter) KeyFilter() KeyFilter {
	if len(f.keyFilters) == 0 {
		return nil
	}
	return func(k []byte, s State) (ok bool, loop bool) {
		ok = !f.logicalOr
		loop = true
		for _, filter := range f.keyFilters {
			sok, sloop := filter(k, s)
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

type filterBuilder struct {
}

func (b filterBuilder) AddOr() {
	panic("not implemented yet")
}

func NewFilter() filterBuilder {
	panic("not implemented yet")
}
