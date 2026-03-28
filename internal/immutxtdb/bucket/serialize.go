package bucket

import "github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"

type BucketUidSerializer struct {
	serialize.Serializer[BucketUid]
}

func (s BucketUidSerializer) Serialize(i BucketUid, o []byte) (int, error) {
	for k := range len(i) {
		o[k] = (i)[k]
	}
	return len(i), nil
}

func (s BucketUidSerializer) Deserialize(i []byte) (BucketUid, error) {
	var o BucketUid
	for k := 0; k < len(o) && k < len(i); k++ {
		o[k] = i[k]
	}
	return o, nil
}

type HashedBucketUidSerializer struct {
	serialize.Serializer[HashedBucketUid]
}

func (s HashedBucketUidSerializer) Serialize(i HashedBucketUid, o []byte) (int, error) {
	for k := range len(i) {
		o[k] = (i)[k]
	}
	return len(i), nil
}

func (s HashedBucketUidSerializer) Deserialize(i []byte) (HashedBucketUid, error) {
	var o HashedBucketUid
	for k := 0; k < len(o) && k < len(i); k++ {
		o[k] = i[k]
	}
	return o, nil
}
