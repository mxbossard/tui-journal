package idx

import "github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"

const DefaultStateSize = 16
const DefaultKeySize = 16
const DefaultValSize = 64
const DefaultPageSize = 10
const DefaultPreloadPageCount = 1

type config[K comparable, V any] struct {
	name, partition string
	// salt            []byte
	stateSize int
	keySize   int
	valSize   int

	keySerializer serialize.Serializer2[K] // convert key to []byte & vice versa
	valSerializer serialize.Serializer2[V] // convert val to []byte & vice versa
	keyHasher     GlidingHasher
	valHasher     GlidingHasher

	pageSize         int
	preloadPageCount int
}

func (c *config[K, V]) SetStateSize(n int) *config[K, V] {
	c.stateSize = n
	return c
}

func (c *config[K, V]) SetKeySize(n int) *config[K, V] {
	c.keySize = n
	c.keySerializer = serialize.AnySerializer[K]{Length: n}
	return c
}

func (c *config[K, V]) SetValSize(n int) *config[K, V] {
	c.valSize = n
	c.valSerializer = serialize.AnySerializer[V]{Length: n}
	return c
}

func (c *config[K, V]) SetPageSize(n int) *config[K, V] {
	c.pageSize = n
	return c
}

func (c *config[K, V]) SetPreloadPageCount(n int) *config[K, V] {
	c.preloadPageCount = n
	return c
}

func (c *config[K, V]) SetKeySerializer0(s serialize.Serializer[K]) *config[K, V] {
	c.keySerializer = serialize.Serializer2Adapter[K]{Wrapped: s}
	return c
}

func (c *config[K, V]) SetValSerializer0(s serialize.Serializer[V]) *config[K, V] {
	c.valSerializer = serialize.Serializer2Adapter[V]{Wrapped: s}
	return c
}

func (c *config[K, V]) SetKeySerializer(s serialize.Serializer2[K]) *config[K, V] {
	c.keySerializer = s
	return c
}

func (c *config[K, V]) SetValSerializer(s serialize.Serializer2[V]) *config[K, V] {
	c.valSerializer = s
	return c
}

func (c *config[K, V]) EnableKeyHasher(salt []byte) *config[K, V] {
	c.keyHasher = NewRotatingHasher(salt, c.keySize)
	return c
}

func (c *config[K, V]) EnableValHasher(salt []byte) *config[K, V] {
	c.valHasher = NewRotatingHasher(salt, c.valSize)
	return c
}

func DefaultConfig[K comparable, V any](name, partition string) config[K, V] {
	keySize := DefaultKeySize
	valSize := DefaultValSize

	c := config[K, V]{
		name:      name,
		partition: partition,

		stateSize: DefaultStateSize,
		keySize:   DefaultKeySize,
		valSize:   DefaultValSize,

		keyHasher: nil,
		valHasher: nil,

		pageSize:         DefaultPageSize,
		preloadPageCount: DefaultPreloadPageCount,
	}

	c.keySerializer = serialize.AnySerializer[K]{Length: keySize}
	c.valSerializer = serialize.AnySerializer[V]{Length: valSize}

	return c
}
