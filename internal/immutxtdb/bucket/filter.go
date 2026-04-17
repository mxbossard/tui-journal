package bucket

import (
	"bytes"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
)

type State idx.State

type Order string

const (
	YoungerFirst = Order("YoungerFirst")
	OlderFirst   = Order("OlderFirst")
)

type Filter interface {
	MatchCreationTime(time.Time) bool
	MatchUpdateTime(time.Time) bool
	MatchBucketState(State) bool
	MatchBucketName(string) bool
	MatchBucketUid(BucketUid) bool
}

type basicFilter struct {
	Filter
	creationTimeAfter  *time.Time
	creationTimeBefore *time.Time
	updateTimeAfter    *time.Time
	updateTimeBefore   *time.Time
	bucketStates       []State
	bucketNames        []string
	bucketUids         []BucketUid
}

func (f basicFilter) MatchCreationTime(t time.Time) bool {
	after := true
	before := true
	if f.creationTimeAfter != nil {
		after = t.After(*f.creationTimeAfter)
	}
	if f.creationTimeBefore != nil {
		before = t.Before(*f.creationTimeBefore)
	}
	return after && before
}

func (f basicFilter) MatchUpdateTime(t time.Time) bool {
	after := true
	before := true
	if f.updateTimeAfter != nil {
		after = t.After(*f.updateTimeAfter)
	}
	if f.updateTimeBefore != nil {
		before = t.Before(*f.updateTimeBefore)
	}
	return after && before
}

func (f basicFilter) MatchBucketState(state State) bool {
	for _, s := range f.bucketStates {
		if bytes.Equal(s, state) {

			return true
		}
	}
	return false
}

func (f basicFilter) MatchBucketName(name string) bool {
	for _, s := range f.bucketNames {
		if s == name {
			return true
		}
	}
	return false
}

func (f basicFilter) MatchBucketUid(uid BucketUid) bool {
	for _, u := range f.bucketUids {
		if u == uid {
			return true
		}
	}
	return false
}

type unionFilter struct {
	Filter
	logicalOr bool
	filters   []basicFilter
}

func (f unionFilter) logicalUnion(matcher func(basicFilter) bool) bool {
	andOk := true
	if f.logicalOr {
		for _, filter := range f.filters {
			ok := matcher(filter)
			if ok {
				return true
			}
		}
	} else {
		for _, filter := range f.filters {
			ok := matcher(filter)
			andOk = andOk && ok
		}
	}
	return andOk
}

func (f unionFilter) MatchCreationTime(t time.Time) bool {
	return f.logicalUnion(func(bf basicFilter) bool {
		return bf.MatchCreationTime(t)
	})
}

func (f unionFilter) MatchUpdateTime(t time.Time) bool {
	return f.logicalUnion(func(bf basicFilter) bool {
		return bf.MatchUpdateTime(t)
	})
}

func (f unionFilter) MatchBucketState(s State) bool {
	return f.logicalUnion(func(bf basicFilter) bool {
		return bf.MatchBucketState(s)
	})
}

func (f unionFilter) MatchBucketName(name string) bool {
	return f.logicalUnion(func(bf basicFilter) bool {
		return bf.MatchBucketName(name)
	})
}

func (f unionFilter) MatchBucketUid(uid BucketUid) bool {
	return f.logicalUnion(func(bf basicFilter) bool {
		return bf.MatchBucketUid(uid)
	})
}
