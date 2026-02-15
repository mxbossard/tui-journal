package immutxtdb

import "github.com/mxbossard/utilz/errorz"

type Query struct {
}

type DB struct {
	rootPath string

	bucketIdx *bucketIndex
	layerIdx  *layerIndex
}

func (d *DB) Bucket(uid string) (*Bucket, error) {
	p, errChan := d.layerIdx.Paginate(uid, BottomToTop, 100)

	var layers []*layer
	for page, ok := p.Next(); ok; {
		_ = page
		for _, entry := range page.Entries() {
			layers = append(layers, entry.Val())
		}
	}
	b := &Bucket{
		db:     d,
		uid:    uid,
		layers: layers,
	}

	return b, errorz.ConsumedAggregated(errChan)
}

func (d DB) Query(query Query) ([]Bucket, error) {
	panic("not implemented yet")
}
