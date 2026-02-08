package immutxtdb

type Query struct {
}

type DB struct {
	rootPath string
}

func (d DB) Bucket(uid string) (Bucket, error) {
	panic("not implemented yet")
}

func (d DB) Query(query Query) ([]Bucket, error) {
	panic("not implemented yet")
}
