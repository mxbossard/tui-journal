package immutxtdb

type Query struct {
}

type DB struct {
	rootPath string

	bucketIdx *bucketIndex
	layerIdx  *layerIndex
}

func (d *DB) Bucket(uid string) (*Bucket, error) {
	p, errChan := d.layerIdx.Paginate(uid, BottomToTop, 100)
	if err != nil {
		return nil, err
	}
	b := &Bucket{
		db:  d,
		uid: uid,
	}
	if !ok {
		return b, nil
	} else {
		layers, err := d.layerIdx.list(uid, BottomToTop, -1)
		if err != nil {
			return nil, err
		}
		b.layers = layers
	}
	panic("not implemented yet")
}

func (d DB) Query(query Query) ([]Bucket, error) {
	panic("not implemented yet")
}
