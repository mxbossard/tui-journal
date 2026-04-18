package bucket

import (
	"bytes"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
)

type State idx.State

type Sorting string

const (
	YoungerFirst = Sorting("YoungerFirst")
	OlderFirst   = Sorting("OlderFirst")
)

type TimeMatcher = func(time.Time) bool
type BucketStateMatcher = func(State) bool
type BucketNameMatcher = func(string) bool
type BucketUidMatcher = func(BucketUid) bool

type Criteria interface {
	CreationTimeMatcher() TimeMatcher
	UpdateTimeMatcher() TimeMatcher
	BucketStateMatcher() BucketStateMatcher
	MatchingStates() []State
	BucketNameMatcher() BucketNameMatcher
	MatchingNames() []string
	BucketUidMatcher() BucketUidMatcher
	MatchingUids() []BucketUid
}

type basicCriterion struct {
	Criteria
	creationTimeAfter  *time.Time
	creationTimeBefore *time.Time
	updateTimeAfter    *time.Time
	updateTimeBefore   *time.Time
	bucketStates       []State
	bucketNames        []string
	bucketUids         []BucketUid
}

func (f basicCriterion) CreationTimeMatcher() TimeMatcher {
	if f.creationTimeAfter != nil || f.creationTimeBefore != nil {
		return f.matchCreationTime
	}
	return nil
}

func (f basicCriterion) UpdateTimeMatcher() TimeMatcher {
	if f.updateTimeAfter != nil || f.updateTimeBefore != nil {
		return f.matchUpdateTime
	}
	return nil
}

func (f basicCriterion) BucketStateMatcher() BucketStateMatcher {
	if len(f.bucketStates) > 0 {
		return f.matchBucketState
	}
	return nil
}

func (f basicCriterion) MatchingStates() []State {
	return f.bucketStates
}

func (f basicCriterion) BucketNameMatcher() BucketNameMatcher {
	if len(f.bucketNames) > 0 {
		return f.matchBucketName
	}
	return nil
}

func (f basicCriterion) MatchingNames() []string {
	return f.bucketNames
}

func (f basicCriterion) BucketUidMatcher() BucketUidMatcher {
	if len(f.bucketUids) > 0 {
		return f.matchBucketUid
	}
	return nil
}

func (f basicCriterion) MatchingUids() []BucketUid {
	return f.bucketUids
}

func (f basicCriterion) matchCreationTime(t time.Time) bool {
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

func (f basicCriterion) matchUpdateTime(t time.Time) bool {
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

func (f basicCriterion) matchBucketState(state State) bool {
	for _, s := range f.bucketStates {
		if bytes.Equal(s, state) {

			return true
		}
	}
	return false
}

func (f basicCriterion) matchBucketName(name string) bool {
	for _, s := range f.bucketNames {
		if s == name {
			return true
		}
	}
	return false
}

func (f basicCriterion) matchBucketUid(uid BucketUid) bool {
	for _, u := range f.bucketUids {
		if u == uid {
			return true
		}
	}
	return false
}

type unionCriteria struct {
	Criteria
	logicalOr bool
	criteria  []basicCriterion
}

func (f unionCriteria) logicalUnion(matcher func(basicCriterion) bool) bool {
	andOk := true
	if f.logicalOr {
		for _, filter := range f.criteria {
			ok := matcher(filter)
			if ok {
				return true
			}
		}
	} else {
		for _, filter := range f.criteria {
			ok := matcher(filter)
			andOk = andOk && ok
		}
	}
	return andOk
}

func (f unionCriteria) CreationTimeMatcher() TimeMatcher {
	return func(t time.Time) bool {
		return f.logicalUnion(func(bf basicCriterion) bool {
			return bf.matchCreationTime(t)
		})
	}
}

func (f unionCriteria) UpdateTimeMatcher() TimeMatcher {
	return func(t time.Time) bool {
		return f.logicalUnion(func(bf basicCriterion) bool {
			return bf.matchUpdateTime(t)
		})
	}
}

func (f unionCriteria) BucketStateMatcher() BucketStateMatcher {
	return func(s State) bool {
		return f.logicalUnion(func(bf basicCriterion) bool {
			return bf.matchBucketState(s)
		})
	}
}

func (f unionCriteria) MatchingStates() []State {
	// return first criterion not nil datas
	for _, c := range f.criteria {
		if m := c.MatchingStates(); len(m) > 0 {
			return m
		}
	}
	return nil
}

func (f unionCriteria) BucketNameMatcher() BucketNameMatcher {
	return func(name string) bool {
		return f.logicalUnion(func(bf basicCriterion) bool {
			return bf.matchBucketName(name)
		})
	}
}

func (f unionCriteria) MatchingNames() []string {
	// return first criterion not nil datas
	for _, c := range f.criteria {
		if m := c.MatchingNames(); len(m) > 0 {
			return m
		}
	}
	return nil
}

func (f unionCriteria) BucketUidMatcher() BucketUidMatcher {
	return func(uid BucketUid) bool {
		return f.logicalUnion(func(bf basicCriterion) bool {
			return bf.matchBucketUid(uid)
		})
	}
}

func (f unionCriteria) MatchingUids() []BucketUid {
	// return first criterion not nil datas
	for _, c := range f.criteria {
		if m := c.MatchingUids(); len(m) > 0 {
			return m
		}
	}
	return nil
}
