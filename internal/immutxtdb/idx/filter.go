package idx

import "time"

// Return ok=true to select entry, return loop=false to stop iterating.
type stateFilter func(s State) (ok bool, loop bool)

// Return ok=true to select entry, return loop=false to stop iterating.
type timeFilter func(t time.Time) (ok bool, loop bool)

// Return ok=true to select entry, return loop=false to stop iterating.
type keyFilter func(k []byte, s State) (ok bool, loop bool)

type Filter interface {
	StateFilter() stateFilter
	TimeFilter() timeFilter
	KeyFilter() keyFilter
}

type aggFilter struct {
	Filter
	logicalOr    bool
	stateFilters []stateFilter
	timeFilters  []timeFilter
	keyFilters   []keyFilter
}

func (f *aggFilter) Add(filters ...*aggFilter) *aggFilter {
	for _, filter := range filters {
		f.AddStateFilter(filter.StateFilter())
		f.AddTimeFilter(filter.TimeFilter())
		f.AddKeyFilter(filter.KeyFilter())
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

func (f *aggFilter) AddKeyFilter(filter keyFilter) *aggFilter {
	if filter != nil {
		f.keyFilters = append(f.keyFilters, filter)
	}
	return f
}

func (f aggFilter) KeyFilter() keyFilter {
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

func KeyFilter(kf keyFilter) *aggFilter {
	filter := &aggFilter{logicalOr: false}
	if kf != nil {
		filter.AddKeyFilter(kf)
	}
	return filter
}

type filterBuilder struct {
}

func (b filterBuilder) AddOr() {
	panic("not implemented yet")
}

func NewFilter() filterBuilder {
	panic("not implemented yet")
}
