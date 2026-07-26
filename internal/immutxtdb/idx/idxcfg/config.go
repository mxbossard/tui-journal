package idxcfg

import (
	idxhash "github.com/mxbossard/tui-journal/internal/immutxtdb/idx/idxhash"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/serialize"
)

const DefaultStateSize = 16
const DefaultKeySize = 16
const DefaultValSize = 64
const DefaultPageSize = 10
const DefaultPreloadPageCount = 1

type Config[K comparable, V any] struct {
	Name, Partition string
	// salt            []byte
	StateSize int
	KeySize   int
	ValSize   int

	keySerializer serialize.Serializer2[K] // convert key to []byte & vice versa
	valSerializer serialize.Serializer2[V] // convert val to []byte & vice versa
	keyHasher     idxhash.GlidingHasher
	valHasher     idxhash.GlidingHasher

	PageSize         int
	PreloadPageCount int
}

func (c Config[K, V]) KeySerializer() serialize.Serializer2[K] {
	return c.keySerializer
}

func (c Config[K, V]) ValSerializer() serialize.Serializer2[V] {
	return c.valSerializer
}

func (c Config[K, V]) KeyHasher() idxhash.GlidingHasher {
	return c.keyHasher
}

func (c Config[K, V]) ValHasher() idxhash.GlidingHasher {
	return c.valHasher
}

func (c *Config[K, V]) SetStateSize(n int) *Config[K, V] {
	c.StateSize = n
	return c
}

func (c *Config[K, V]) SetKeySize(n int) *Config[K, V] {
	c.KeySize = n
	c.keySerializer = serialize.AnySerializer[K]{Length: n}
	return c
}

func (c *Config[K, V]) SetValSize(n int) *Config[K, V] {
	c.ValSize = n
	c.valSerializer = serialize.AnySerializer[V]{Length: n}
	return c
}

func (c *Config[K, V]) SetPageSize(n int) *Config[K, V] {
	c.PageSize = n
	return c
}

func (c *Config[K, V]) SetPreloadPageCount(n int) *Config[K, V] {
	c.PreloadPageCount = n
	return c
}

func (c *Config[K, V]) SetKeySerializer0(s serialize.Serializer[K]) *Config[K, V] {
	c.keySerializer = serialize.Serializer2Adapter[K]{Wrapped: s}
	return c
}

func (c *Config[K, V]) SetValSerializer0(s serialize.Serializer[V]) *Config[K, V] {
	c.valSerializer = serialize.Serializer2Adapter[V]{Wrapped: s}
	return c
}

func (c *Config[K, V]) SetKeySerializer(s serialize.Serializer2[K]) *Config[K, V] {
	c.keySerializer = s
	return c
}

func (c *Config[K, V]) SetValSerializer(s serialize.Serializer2[V]) *Config[K, V] {
	c.valSerializer = s
	return c
}

func (c *Config[K, V]) EnableKeyHasher(salt []byte) *Config[K, V] {
	c.keyHasher = idxhash.NewRotatingHasher(salt, c.KeySize)
	return c
}

func (c *Config[K, V]) EnableValHasher(salt []byte) *Config[K, V] {
	c.valHasher = idxhash.NewRotatingHasher(salt, c.ValSize)
	return c
}

// Build Default idx config.
// Name should be a functionnal name
// Partition should be a technical qualifier (like a device)
func DefaultConfig[K comparable, V any](name, partition string) Config[K, V] {
	keySize := DefaultKeySize
	valSize := DefaultValSize

	c := Config[K, V]{
		Name:      name,
		Partition: partition,

		StateSize: DefaultStateSize,
		KeySize:   DefaultKeySize,
		ValSize:   DefaultValSize,

		keyHasher: nil,
		valHasher: nil,

		PageSize:         DefaultPageSize,
		PreloadPageCount: DefaultPreloadPageCount,
	}

	c.keySerializer = serialize.AnySerializer[K]{Length: keySize}
	c.valSerializer = serialize.AnySerializer[V]{Length: valSize}

	return c
}
