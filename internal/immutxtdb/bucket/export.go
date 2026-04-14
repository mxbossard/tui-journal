package bucket

type BucketExport struct {
	b *Bucket
}

func (s *bucketService) Export(uid BucketUid, squash bool) (*BucketExport, error) {
	b, err := s.Get(uid)
	if err != nil {
		return nil, err
	}
	return &BucketExport{b}, nil
}

func (s *bucketService) Import(export *BucketExport) error {
	panic("not implemented yet")
}
