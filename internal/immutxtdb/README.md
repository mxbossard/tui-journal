# Journal DB

## TODO
- [x] BlocsFile first impl
- [x] Bucket & Layer indexs first impl
- [x] Document & Text indexs first impl
- [x] Time index first impl
- [_] Test main usecase : Create a document and index it (bucket, layers, document, text, time)
- [_] Implement RotatingHash, who's responsability ?
- [_] Validate idx errChan usage
- [_] Implem idx Filter methods
- [_] A first text diff/layering impl (use a version / impl qualifier ?)
- [_] Manage preloading of idx files ?
- [_] Do we need to optimize "file reading stop" at snapshot layer ? Could provide a func to decide "preloading stop".
- [_] Encryption of BlocsFiles impl
- [_] Rotating Hash impl
- [_] Randomly generated SecretKey ciphered with user passphrase


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
- paging : index results MUST be returned aggregated by pages of finite size (never return all results simultaneously)
- cursor : ?
- search : 

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
- Paging
- Search

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

### Index
- Index --- Paginer --- Page --- BlocsFile
- One Index reference multiple BlocsFiles
- One Index return a Paginer
- One Paginer iterate over all the BlocsFile
- Index serialization format must be the same in all the file. The serialization version MUST be written on file head.
- HEAD: [VERSION,LINE_SIZE,FORMAT?...]
- LINE: [SEQ,STATE,DATA]
- uid SHOULD contains only lowercase alphanum plus some special chars ex:[-a-z0-9_/:<>=+#*.] 
- example of doc uid: project_electronic_analogic_ampli-op_foobarbaz : 46 chars
- how to compress efficiently text in a byte stream ? For base64 1 byte can encode 4 chars => 128 chars could be encoded in 32 bytes
- Could store UTF-8 chars for universality and simplicity => x12 bytes
- 10.000 items of 40 bytes in idx : 400 kB
- 10.000 items of 300 bytes in idx : 3 MB
- First simple implem: use ASCII lower case only each char encoded on 1 byte.

## Indexing

### Use Cases
- List dumps by time (asc or desc)
- Search dumps in time window
- List topics
- Search topics (wildcard ?)
- List dumps by topics
- List new topics
- List documents by topics
- Search documents

#### Use case 1 : list last dumps (each dump is in a separate bucket)
- read bucket-time index FROM end => get buckets uids
- read bucket index ALL ? => get bucket locations 
- read layer index => get bucket layers (file+blocId)
- build documents

#### Use case 2 : search topics
- read topic index ALL => get all topics (need to load all topics to order and search)
- serach algorithm on cached topics

#### Use case 3 : list topic dumps (each dump is in a separate bucket)
- read bucket-topic index FROM end => get buckets uids of kind dump
- read layer index => get bucket layers (file+blocId)
- build documents

#### Use case 4 : list topic dump references (each dump reference is in one on more dumps)
- read bucket-topic-ref index FROM end => get buckets uids of kind dump + inDoc ref
- read layer index => get bucket layers (file+blocId)
- build documents

#### Use case 5 : list topic documents (each document is in a separate bucket)
- read bucket-topic index FROM end => get buckets uids of kind document
- read layer index => get bucket layers (file+blocId)
- build documents

#### Use case 6 : get document in a bucket
- read layer index => get bucket layers (file+blocId)
- build documents

### Index needed properties
- Readable by start or by end, fully or by page
- Associate each index entry with a seq for fast count, and cache optimizations
- Key, Value store
- Associte entries with a State for Filtering purpose (stop index read, differentiate object kinds)

### Indexes
#### Bucket index
- Format: CYPHER(BUCKET_UID, STATE_PRIVATE_DATA)
- Encrypt all the files
- STATE_PRIVATE_DATA 

#### Layer index
- Format: PLAIN(RH(BUCKET_UID), RH(LAYER_FILE), BLOC_ID, STATE_PUBLIC_DATA)
- Use "Rotating Hash" to cipher bucket and layer to mitigate "usage data inference".
- Store layer state (Snapshoted or not) to reduce read count if possible.
- Store item states in index ?

#### Document index
- Format: PLAIN(RH(TOPIC), RH(BUCKET_UID), STATE_PUBLIC_DATA)
- Use "Rotating Hash" to cipher bucket and layer to mitigate "usage data inference".

#### Text index
- Format: PLAIN(RH(TOPIC), RH(BUCKET_UID), POS, LEN, STATE_PUBLIC_DATA)
- Use "Rotating Hash" to cipher bucket and layer to mitigate "usage data inference".

#### Time indexs
- Format: PLAIN(TIME, RH(BUCKET_UID))

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