# Journal DB

## Purpose
DB to store text documents

## Objectives
- Minimize file updates: prefer to write a new file than append to a file.
- Never delete datas.
- Quick text research in all stored text

## Definitions
- document : a text document stored in DB.
- bucket : a collection of layers mergeable into a document (a projection).
- layer : patch of document attached with metadata and a version
- version : version of the document
- metadata : document qualifying data, creation timestamp, ...
- squash : merge last layers
- commit : immune a layer to squashing
- snapshot : create a new terminal layer for perf purpose allowing to ignore previous layers (concept from event sourcing).
- compact : gather multiple buckets in minimum of files
- offset? : 
- index : 

## Requirements
- Too much files > bad for perf and for git db
- reencrypt appended files > replace all the file > bad for git which cannot leverage diffs
- squash function to merge documents layers in bucket to reduce layers count (need more thinking)
- "commit" function to forbid layer squashing further than commited layers
- compact function to reduce file count keeping all document layers in a small amount of files.
- 2 phases file writing : 
  - concurrent write buckets new layers into encrypted temp files
  - compacting buckets append temp files into a minimum of already compacted files.
- store multiple buckets in one file.
- "compacting offset" to ensure compacting is done into dedicated filesDb appended files are exclusive to this device.

## Services needed
- save content into a "bucket" by "path"
- read a "bucket" content from "path"
- aggregate buckets by path ?
- list bucket pathes ?
- search buckets by path wildcard (need wildcard path indexes ?)
- search buckets by metadata (need specific indexes ?)
- text search in all buckets
- pagination
- create an index ? (bucked list could be in a special index ?)
- append index entry: (bucket, start, end)
- manage bucket parts (read, write, aggregate bucket parts) ?
- patch document on bucket content updates ? (do not edit initial bucket content but add a new layer to the bucket)
    - layered buckets ?
    - bucket versioning 
    - which format for diff ?
    - copy on write ?
    - parted buckets (index listing bucket parts) ?
    - metadata attach to each bucket version / layer
    - describe all bucket layers
- immutability
    - never delete anything : 
    - Write once read many (WORM) ?
    - each updates or erase is not deletes
- list history of a bucket (display all layers)


## Implem ideas
- Store Text Document
- Index Document parts
- A document is stored in a bucket
- A bucket contain multiple Read Only version of a document called Layers
- A layer is stored in a layer file, a layer file can contains multiple layers encrypted sequentialy in a "bloc encrypted file".
- Reading a bucket read all the layers concurrently.
- If too much layer COULD "smash" all layers in one new layer.
- SHOULD index files be stored in only one files ?
  - CAN we update the index just appending files ?
- SHOULD we authorize file appending in buckets under a size or time frame ?
  - small layers could be appended in same bucket file for a small timeframe
- COULD append encrypted data into files, not modifying previous parts of the file. need a header to know size of ciphered text before or after nonce.
- CAN we store an index in a bucket ?
  - may not be performant
  - appending in index CAN be performed writing a new layer.
  - updating an index CAN be performed writing a new layer.
  - fat index CAN be squashed 


## Implem details
### Bucket
- Ref all layers of a document
- Project a document
- Can be commited
- Can be squashed
- Can be snapshoted
- ? Where is it stored ? => in an index file

### Layer
- Layer ref a content, a version & a metadata
- Is written appending into a file
- Can be marked commited
- Can be marked snapshoted

### Metadata
- Ref a version
- Ref a creation time
- Ref labels

### Version
- Numerical increment ?

### Indexes
#### Bucket index
- Format: (BUCKET_UID)
- Encrypt all the files

#### Layer index
- Format: (RH(BUCKET_UID), RH(LAYER_FILE), POS, LEN)
- Use "Rotating Hash" to cipher bucket and layer to mitigate "usage data inference".
- Store layer state (Snapshoted or not) to reduce read count if possible.

#### Document index
- Format: (RH(TOPIC), RH(BUCKET_UID))
- Use "Rotating Hash" to cipher bucket and layer to mitigate "usage data inference".

#### Text index
- Format: (RH(TOPIC), RH(BUCKET_UID), POS, LEN)
- Use "Rotating Hash" to cipher bucket and layer to mitigate "usage data inference".

#### Ideas
- Only commited layers are referenced in document & text index ?
- Not commited files could be stored temporarilly ?
- Remote sync of files need files to be commited ?
- Do not remote sync temp files ?
- Layer state: snapshoted or not
- To mitigate fequency uage data inference use a "Rolling Hash" function using a salt based on the line number in the index file.
- Reading index files by bloc starting at bottom.
- One Index will span over multiple files.

#### Separating index files for each device
- Purpose: concurrent writes on different devices are stored with concurrent states.
- If a buckets doublon exists it MUST not be a problem
- If two concurrent layers exists, attempt to auto merge it
- If cannot auto merge 2 concurrent layers report conflict to user to merge in a new layer.

### Files
#### Bloc readable from bottom
Need a to know the entry count (one entry by line for exemple)
Need to be readable by entry from bottom (entry size at the end of the entry ?)

#### Rolling Hashed append only files (plaintext)
File can only be appended. It contains only hashed data or not secret data.

#### Fully encrypted files
File can writen and rewritten it contains plaintext and must be entirely ciphered

#### Bloc encrypted files
File contains sequential blocs each ciphered with same key but with different nonce. Each bloc is independant and can be read or updated independently.