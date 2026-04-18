package bucket

import (
	"fmt"
)

type BucketExport struct {
	name     string
	metadata *Metadata
	headers  []*Header
	layers   []*Layer
}

func (s *bucketService) Export(uid BucketUid, headerHist, dataHist, squash bool) (*BucketExport, error) {
	if headerHist {
		panic("not implemented yet")
	}

	if dataHist {
		panic("not implemented yet")
	}

	if squash {
		panic("not implemented yet")
	}

	// FIXME: do we export all names & headers or only last one ?
	// - 2 names means a bucket renaming Do we want to list all names ? No because we don't want to search over hidden names. But we can trace renaming history with header history.
	// - 2 headers means a header update
	// - We want to have controll over what we export => headerHistory

	// FIXME: use idx.Export ?
	// partitions, err := scanServicePartitions(s.dir)
	// for _, part := range partitions {
	// 	bNameIdx, hRefIdx, bRefIdx, err := getServiceIndexes(s.dir, s.salt, part)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	// FIXME: what to do with bucketNameIdx ? Need key (name) to make an export.
	// 	// May not necessary to scan becaus we just want to add last name in the idx.
	// 	//idx.NewExport(bNameIdx, idx.TopToBottom, "foo")
	// 	_ = bNameIdx

	// 	hRedExp, err := idx.NewExport(hRefIdx, idx.TopToBottom, uid)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	bRedExp, err := idx.NewExport(bRefIdx, idx.TopToBottom, uid)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	b, err := s.Get(uid)
	if err != nil {
		return nil, err
	}

	var headers []*Header
	var layers []*Layer

	// For now export only last header (no header history)
	headers = append(headers, &b.Header)

	// For now export all layers
	layerIt, err := b.LayerIt(LatestVersion)
	if err != nil {
		return nil, err
	}
	for err, l := range layerIt {
		if err != nil {
			return nil, err
		}
		layers = append(layers, l)
	}

	return &BucketExport{
		name:     b.Header.Name,
		metadata: b.Metadata,
		headers:  headers,
		layers:   layers,
	}, nil
}

func (s *bucketService) Import(export *BucketExport, partition string) error {
	// FIXME: what to do if bucket name already exists ?
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	// FIXME: use idx.Import ?

	// FIXME: keep uid ? Or generate a new one ?
	// FIXME: which timing to use ? Keep original timings ?
	b := newBucket(s)
	b.Header = *export.headers[0]
	b.Header.Name = export.name
	b.Metadata = export.metadata

	// 1- Store bucketName
	err := s.addName(partition, *b.Header.Created, b.Header.Name, b.Header.Uid)
	if err != nil {
		return fmt.Errorf("import: unable to add Name: %w", err)
	}

	// 2- Store header (COULD have multiple entries)
	_, err = s.addHeader(partition, dummyState, &b.Header)
	if err != nil {
		b.Header.Created = nil
		return fmt.Errorf("import: unable to add Header: %w", err)
	}

	// 3- Store All Layers & Metadatas
	for _, l := range export.layers {
		_, err := s.addLayer(partition, l.State, b.Header.Uid, l.Metadata, l.Content)
		if err != nil {
			return fmt.Errorf("create: unable to add Layer: %w", err)
		}
	}

	return nil
}
