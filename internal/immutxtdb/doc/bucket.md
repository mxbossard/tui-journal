## Bucket Service conflicts

### Saving 2 layers with same version in same bucket service (same indexes files)
Bucket service MUST return a conflict error on second Save() invocation.
=> Must load all layers before performing Save().

### Saving 2 layers with same version in different bucket service (same indexes files but different partitions)
Bucket service MUST return a conflict error on second Save() invocation.
=> Must load all layers before performing Save().

### Saving 2 layers with same version in different bucket service (different indexes files)
Bucket service MUST not conflict on Save() invocations.

## Two Phase Store conflicts

### Saving 2 identical buckets in same store instance (same store dir)
Ephemeral store MUST return a conflict error on second Save() invocation.
=> Must load all layers before performing a save.

### Saving 2 identical buckets in two different store instance (same store dir)

### Saving 2 identical buckets in two different store instance (different store dirs

### Commiting 2 identical buckets in same store instance (same store dir)

### Commiting 2 identical buckets in two different store instance (same store dir)

### Commiting 2 identical buckets in two different store instance (different store dirs

