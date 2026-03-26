package db

import (
	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/index"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
)

type Query struct {
}

type DB struct {
	rootPath string

	bucketIdx index.BucketIndex
	layerIdx  index.LayerIndex
}

func RotatingHashString(s string) *model.HashedBucketUid {
	panic("not implemented yet")
}

func (d *DB) Bucket(uid string) (*model.Bucket, error) {
	rhUid := RotatingHashString(uid)
	p, err := d.layerIdx.Paginate(rhUid, idx.BottomToTop)
	if err != nil {
		return nil, err
	}

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

	panic("not implemented yet")
	b := &model.Bucket{Uid: model.BucketUid(uid)}
	return b, nil
}

func (d DB) Query(query Query) ([]model.Bucket, error) {
	panic("not implemented yet")
}
