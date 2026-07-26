package idx

import (
	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx/idxcfg"
)

// Build Default idx config.
// Name should be a functionnal name
// Partition should be a technical qualifier (like a device)
func DefaultConfig[K comparable, V any](name, partition string) idxcfg.Config[K, V] {
	return idxcfg.DefaultConfig[K, V](name, partition)
}
