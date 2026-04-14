package idx

type Export[K comparable, V any] struct {
	Key      K
	Ordering Order
	Entries  []Entry[K, V]
}

// Export all entries of supplied Key.
// FSS Key must be supplied because Key MAY be Hashed and we need it in order to import the export later.
func NewExport[K comparable, V any](idx Index[K, V], o Order, key K) (*Export[K, V], error) {
	paginer, err := idx.Paginate(key, o)
	if err != nil {
		return nil, err
	}

	var entries []Entry[K, V]
	for e := range paginer.All() {
		entries = append(entries, e)
	}
	return &Export[K, V]{
		Key:      key,
		Ordering: o,
		Entries:  entries,
	}, nil

}

func ImportAll[K comparable, V any](idx Index[K, V], exp *Export[K, V], o Order) ([]Entry[K, V], error) {
	var imported []Entry[K, V]
	if exp.Ordering == o {
		// Keep ordering
		for _, entry := range exp.Entries {
			e, err := idx.Add(entry.State(), entry.Time(), exp.Key, entry.Val())
			if err != nil {
				return imported, err
			}
			imported = append(imported, e)
		}
	} else {
		// Reverse ordering
		for k := len(exp.Entries) - 1; k >= 0; k-- {
			entry := exp.Entries[k]
			e, err := idx.Add(entry.State(), entry.Time(), exp.Key, entry.Val())
			if err != nil {
				return imported, err
			}
			imported = append(imported, e)
		}
	}
	return imported, nil
}
