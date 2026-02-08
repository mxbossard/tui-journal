package immutxtdb

import "time"

type Labels map[string]string

type Metadata struct {
	version    int
	created    time.Time
	updated    time.Time
	labels     Labels
	commited   bool
	snapshoted bool
}

type Document struct {
	content  string
	metadata *Metadata
}

type layer struct {
	content  string
	metadata *Metadata
}

type Bucket struct {
	uid    string
	layers []*layer
}

func (b Bucket) Project() (Document, error) {
	panic("not implemented yet")
}

func (b Bucket) Save(content string, labels Labels) error {
	panic("not implemented yet")
}

func (b Bucket) Commit() error {
	panic("not implemented yet")
}

func (b Bucket) Squash() error {
	panic("not implemented yet")
}
