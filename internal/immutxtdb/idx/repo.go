package idx

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/mxbossard/utilz/filez"
)

type blocsRepo struct {
	indexDir   string
	salt       []byte
	passphrase string
	encoder    IdxEncoder

	partitionIdxFiles []*filez.BlocsFile
	otherIdxFiles     []*filez.BlocsFile
}

func (r blocsRepo) SelectPartitionBlocFile(s State, k []byte) *filez.BlocsFile {
	return r.partitionIdxFiles[0]
}

func (r blocsRepo) Scan(ordering Order, scanner func([]byte, error) bool) {
	// TODO: cache all the bloc file content ?
	// TODO: call all the index content ?
	// FIXME : which order of idx files to iterate ?

	idxFiles := append(r.partitionIdxFiles, r.otherIdxFiles...)
	var loop bool
	for _, bf := range idxFiles {
		for err, b := range bf.All(filez.BlocOrdering(ordering)) {
			// fmt.Printf("filter: loop2 b: %v\n", b.Uid)
			// Pass err to scanner
			loop = scanner(b.Bytes(), err)
			if !loop {
				// Stop looping
				return
			}
		}
	}
}

func (r blocsRepo) ScanDecodeAll(ordering Order, callback func(seq int, t time.Time, s State, key []byte, val []byte, err error) bool) {
	loop := true
	r.Scan(ordering, func(b []byte, err error) bool {
		// Pass err to scanner
		if err != nil {
			loop = callback(-1, time.Time{}, nil, nil, nil, err)
		} else {
			wrapped := func(seq int, t time.Time, s State, key []byte, val []byte, err error) bool {
				loop = callback(seq, t, s, key, val, err)
				return loop
			}
			r.encoder.DecodeAll(ordering, b, wrapped)
		}
		return loop
	})
}

func (r blocsRepo) LoadSeqs() (map[string]int, error) {
	seqs := make(map[string]int)
	for _, bf := range r.partitionIdxFiles {
		// Decode last line of last bloc to get current seq
		bloc, err := bf.GetLastNonEmptyBloc()
		if err == filez.ErrNotExist {
			// No bloc to read
			err = nil
			continue
		} else if err != nil {
			return nil, fmt.Errorf("unable to get last bloc: %w", err)
		}
		if bloc.Len() > 0 {
			lastSeq, _, _, _, _, err := r.encoder.DecodeLastWord(bloc.Bytes())
			if err != nil {
				return nil, fmt.Errorf("unable to decode last word: %w", err)
			}
			seqs[bf.Name()] = lastSeq + 1
		}
	}
	return seqs, nil
}

func DefaultBlocsRepo[K comparable, V any](indexDir string, cfg config[K, V]) (r blocsRepo, err error) {
	firstPartitionFilepath := filepath.Join(indexDir, fmt.Sprintf("%s-%s-001.idx", cfg.name, cfg.partition))
	dbf1, err := filez.NewBlocsFile(firstPartitionFilepath, 256, 100)
	if err != nil {
		return r, fmt.Errorf("unable to build blocs file: %w", err)
	}
	enc := NewByteSliceEncoder(0, cfg.stateSize, cfg.keySize, cfg.valSize)

	r.indexDir = indexDir
	r.encoder = enc
	r.partitionIdxFiles = []*filez.BlocsFile{dbf1}
	r.otherIdxFiles = nil

	return r, nil
}
