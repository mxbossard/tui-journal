package immutxtdb

type cursor[T any] struct {
}

func (c cursor[T]) HasNext() bool {
	panic("not implemented yet")
}

func (c cursor[T]) Next() []string {
	panic("not implemented yet")
}

type bucketIndex struct {
	filepathes []string
}

func (i bucketIndex) add(uid string) error {
	panic("not implemented yet")
}

func (i bucketIndex) count() (int, error) {
	panic("not implemented yet")
}

// Manage sorting
func (i bucketIndex) listAll(uid string, limit int) (cursor[Bucket], error) {
	panic("not implemented yet")
}

type layerIndex struct {
	filepathes []string
}

func (i layerIndex) add(bucketUid string, filepath string, pos, len int) error {
	panic("not implemented yet")
}

func (i layerIndex) count() (int, error) {
	panic("not implemented yet")
}

// Manage sorting
func (i layerIndex) listAll(uid string, limit int) (cursor[layer], error) {
	panic("not implemented yet")
}

type documentIndex struct {
	filepathes []string
}

func (i documentIndex) add(topic string, bucketUid string) error {
	panic("not implemented yet")
}

func (i documentIndex) count() (int, error) {
	panic("not implemented yet")
}

// Manage sorting
func (i documentIndex) listAll(uid string, limit int) (cursor[Bucket], error) {
	panic("not implemented yet")
}

type textIndex struct {
	filepathes []string
}

func (i textIndex) add(topic string, bucketUid string, pos, len int) error {
	panic("not implemented yet")
}

func (i textIndex) count() (int, error) {
	panic("not implemented yet")
}

// Manage sorting
func (i textIndex) listAll(uid string, limit int) (cursor[Bucket], error) {
	panic("not implemented yet")
}
