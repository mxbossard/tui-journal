## Bucket Service conflicts

### Conflict sources
- Saving 2 instances of Buckets concurrently
  - 2 Concurrent save in same partition => ErrVersionMissmatch
  - 2 concurrent save in different partitions => ErrVersionMissmatch
- Importing a Bucket which already exists
  - imported version must be greater than stored version OR => ErrVersionMissmatch
- Projecting a Bucket with 2 layers of identical version (following merged files of different partitions)

## Two Phase Store conflicts

### Conflict sources
- Saving 2 instances of Buckets concurrently
- Commiting 2 instances of Bucket concurrently




### Saving 2 identical buckets in same store instance (same store dir)
Ephemeral store MUST return a conflict error on second Save() invocation.
=> Must load all layers before performing a save.

### Saving 2 identical buckets in two different store instance (same store dir)

### Saving 2 identical buckets in two different store instance (different store dirs)

### Commiting 2 identical buckets in same store instance (same store dir)

### Commiting 2 identical buckets in two different store instance (same store dir)

### Commiting 2 identical buckets in two different store instance (different store dirs)

