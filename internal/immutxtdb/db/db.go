package db

import (
	"github.com/mxbossard/tui-journal/internal/immutxtdb/index"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/mxbossard/utilz/errorz"
)

type Query struct {
}

type DB struct {
	rootPath string

	bucketIdx *index.BucketIndex
	layerIdx  *index.LayerIndex
}

func (d *DB) Bucket(uid string) (*model.Bucket, error) {
	p, errChan := d.layerIdx.Paginate(uid, model.BottomToTop, 100)

	var layers []*model.LayerRef
	for page, ok, err := p.Next(); ok; {
		if err != nil {
			return nil, err
		}
		_ = page
		for _, entry := range page.Entries() {
			layers = append(layers, entry.Val())
		}
	}
	b := model.NewBucket(uid, layers)

	return b, errorz.ConsumedAggregated(errChan)
}

func (d DB) Query(query Query) ([]model.Bucket, error) {
	panic("not implemented yet")
}
