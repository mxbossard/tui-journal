package service

import (
	"bytes"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/index"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/ztring"
)

// Aggregation of layers
type Dump struct {
	model.Bucket
}

type Doc struct {
	model.Bucket
}

type Layer string

type Text string

type Topic string

/*
## What is the difference between dump and doc ?
  - Dump is an immutable journal log entry attached to only one day. Implemented by a collection of layers.
  - Doc is a named & editable document. Implemented by a collection of layers.
  - => No difference in implementation ?

## Main questions :
- What Metadata ?
- Where to store Metadata ?
- On what attach Metadata ? Layer ? Bucket ? Doc ?

## Model
- Doc
- Dump ? is it a Doc ?
- Metadata
- Bucket
- Layer
- Where are stored Metadata ?
- Do we need Metadata when listing docs/dumps ?
- Do we want to attach metadata to Layer or to Doc ?

### Metadata

*/

var (
	dump1 = ztring.LoremIpsumWords(10)
	dump2 = ztring.LoremIpsumWords(20)
	dump3 = ztring.LoremIpsumWords(30)

	dumpState          = idx.BuildState(index.BucketIdxStateSize, "dmp")
	dumpNewState       = idx.BuildState(index.BucketIdxStateSize, "dmpNew")
	layerState         = idx.BuildState(index.BucketIdxStateSize, "lyr")
	layerSnapshotState = idx.BuildState(index.BucketIdxStateSize, "lyrSnap")
)

type idxService struct {
	bucketIdx       index.BucketIndex
	bucketByTimeIdx index.DocByTimeIndex
	layerIdx        index.LayerIndex
}

func NewIdxService(dir, device string) (*idxService, error) {
	bucketIdxDir := filepath.Join(dir, "bucketIdx")
	bucketByTimeIdxDir := filepath.Join(dir, "bucketByTimeIdx")
	layerIdxDir := filepath.Join(dir, "layerIdx")

	err := os.MkdirAll(bucketIdxDir, 0700)
	if err != nil {
		return nil, err
	}
	bucketIdx, err := index.NewBucketIndex(bucketIdxDir, device)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(bucketByTimeIdxDir, 0700)
	if err != nil {
		return nil, err
	}
	bucketByTimeIdx, err := index.NewCreationTimeIndex(bucketByTimeIdxDir, device)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(layerIdxDir, 0700)
	if err != nil {
		return nil, err
	}
	layerIdx, err := index.NewLayerIndex(layerIdxDir, device)
	if err != nil {
		return nil, err
	}

	return &idxService{
		bucketIdx:       bucketIdx,
		bucketByTimeIdx: bucketByTimeIdx,
		layerIdx:        layerIdx,
	}, nil
}

func ForgeDumpName(device string, when time.Time) string {
	return fmt.Sprintf("dump-%s-%d", device, when.Unix())
}

func RotatingHashString(salt []byte, pos uint32, s string) (*index.HashedBucketUid, error) {
	hash := sha512.New()
	hash.Write(salt)
	err := binary.Write(hash, binary.BigEndian, pos)
	if err != nil {
		return nil, err
	}
	hash.Write([]byte(s))
	hashed := hash.Sum(nil)
	uid := index.HashedBucketUid(hashed[:index.LayerIdxKeySize])
	return &uid, nil
}

func GetBlocWriter(device string) (*model.BlocRef, io.Writer, int, error) {
	//blocs.GetLastBloc()
	panic("not implemented yet")
}

// ------------- Dumps ---------------

func UseCaseDump0_Create(device, txt string) (*Dump, error) {
	tmpDir := filez.MkTempOrPanic("UseCaseDump0_Create")
	idxService, err := NewIdxService(tmpDir, device)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	// 0- Forge dump name
	name := ForgeDumpName(device, now)

	// 1- Create a bucket
	err = idxService.bucketIdx.Add(dumpNewState, nil, name)
	if err != nil {
		return nil, err
	}

	// 2- Create a bucket-time idx entry
	bucketRhUid, err := RotatingHashString(name)
	if err != nil {
		return nil, err
	}
	idxService.bucketByTimeIdx.Add(dumpNewState, now, bucketRhUid[:])

	// 3- Store the content
	blocRef, blocWriter, pos, err := GetBlocWriter(device)
	if err != nil {
		return nil, err
	}
	n, err := blocWriter.Write([]byte(txt))
	if err != nil {
		return nil, err
	}
	layerRef := &model.LayerRef{
		BlocRef: blocRef,
		Pos:     pos,
		Len:     n,
	}

	// 4- Create a layer idx entry
	rootLayerState := idx.BuildState(index.BucketIdxStateSize, "root")
	rhBucketUid, err := RotatingHashString(name)
	if err != nil {
		return nil, err
	}
	idxService.layerIdx.Add(rootLayerState, rhBucketUid, layerRef)

	// rootLayer := model.Layer{
	// 	Metadata: model.LayerMetadata{
	// 		Version: 0, // First layer
	// 		Created: now,
	// 	},
	// 	Commited:   false,
	// 	Snapshoted: true, // First layer snapshoted by definition
	// }

	var layerRefIt iter.Seq[*model.LayerRef] = func(yield func(*model.LayerRef) bool) {
		yield(layerRef)
	}

	d := &Dump{
		Bucket: model.Bucket{
			Uid: model.BucketUid(name),
			Metadata: model.BucketMetadata{
				Created: now,
				Updated: now,
				Labels:  nil,
			},
			LayerRefIt: layerRefIt,
		},
	}
	return d, nil
}

func ConsumeErrorIfAny(errChan chan error) (err error) {
	// Non blocking chan consumption
	select {
	case err = <-errChan:
	default:
	}
	return
}

// Must return a list of dumps able to lazy load their layers.
func UseCaseDump1_ListLast(count int) ([]*Dump, error) {
	// 1- Browse bucket-time idx to find last dumps ref
	tmpDir := filez.MkTempOrPanic("UseCaseDump0_Create")
	device := "pif"
	idxService, err := NewIdxService(tmpDir, device)
	if err != nil {
		return nil, err
	}
	dumpStateFilter := func(s idx.State, stop func()) bool {
		return bytes.Equal(s[0:3], dumpState[0:3])
	}
	k := 0
	var bucketRhUids []*index.HashedBucketUid
	paginer, errChan := idxService.bucketByTimeIdx.FilterAll(idx.BottomToTop, 10, dumpStateFilter, nil)
BucketLoop:
	for err, page := range paginer.All() {
		// FIXME: what is this error ?
		if err != nil {
			return nil, err
		}
		if err := ConsumeErrorIfAny(errChan); err != nil {
			return nil, err
		}
		// FIXME what is this pos ? is it seq ?
		for pos, entry := range page.All() {
			_ = pos
			if k == count {
				break BucketLoop
			}
			// var array [128]byte
			// copy(array[:], entry.Val())
			array := index.HashedBucketUid(entry.Val())
			bucketRhUids = append(bucketRhUids, &array)
		}
	}

	layerStateFilter := func(s idx.State, stop func()) bool {
		return bytes.Equal(s[0:3], layerState[0:3])
	}
	var snapshotedLayersRhUids []*index.HashedBucketUid
	layerFilter := func(k *index.HashedBucketUid, s idx.State, stop func()) bool {
		if slices.Contains(bucketRhUids, k) {
			if slices.Contains(snapshotedLayersRhUids, k) {
				return false
			}
			if bytes.Equal(s, layerSnapshotState) {
				// Layer is a snapshot
				snapshotedLayersRhUids = append(snapshotedLayersRhUids, k)
			}
			return true
		}
		return false
	}
	var dumps []*Dump
	var layersByRhUid map[[128]byte][]*model.LayerRef
	paginer2, errChan := idxService.layerIdx.FilterAll(idx.BottomToTop, 10, layerStateFilter, layerFilter)
	for err, page := range paginer2.All() {
		// FIXME: what is this error ?
		if err != nil {
			return nil, err
		}
		if err := ConsumeErrorIfAny(errChan); err != nil {
			return nil, err
		}
		// FIXME what is this pos ? is it seq ?
		for pos, entry := range page.All() {
			_ = pos
			if k == count {
				return dumps, nil
			}
			layersByRhUid[*entry.Key()] = append(layersByRhUid[*entry.Key()], entry.Val())
		}
	}

	var layerRefIt iter.Seq[*model.LayerRef]

	for rhUid, layerRefs := range layersByRhUid {
		// FIXME: need to find UID
		// FIXME: do not have metadata here ? Where are stored metadatas ? Do we need Metadata before projecting the document ?
		_ = layerRefs
		d := &Dump{
			Bucket: model.Bucket{
				Uid:        model.BucketUid((rhUid[:])),
				LayerRefIt: layerRefIt,
			},
		}
		dumps = append(dumps, d)
	}

	panic("not implemented yet")
}

func UseCaseDump2_Get(d Dump) (string, error) {
	// 1- Browse dump associated bucket
	// 2- Aggregate dump layers
	// 3- Project embeded document
	panic("not implemented yet")
}

func UseCaseDump3_Add(d Dump, txt string) ([]Dump, error) {
	// 1- Create a layer
	// 2- Store the "Add" layer content
	panic("not implemented yet")
}

func UseCaseDump4_Update(d Dump, txt string) ([]Dump, error) {
	// 1- Create a layer
	// 2- Store the "Diff" layer content
	panic("not implemented yet")
}

// ------------- Docs ---------------

func UseCaseDoc0_Create(uid, txt string) (Doc, error) {
	panic("not implemented yet")
}

func UseCaseDoc1_Update(uid, txt string) (Doc, error) {
	panic("not implemented yet")
}

// ------------- Topics ---------------

func UseCaseTopic0_Create(topic Topic, ref any) error {
	panic("not implemented yet")
}

func UseCaseTopic1_List(count int) ([]Topic, error) {
	panic("not implemented yet")
}

func UseCaseTopic2_Search(query string, count int) ([]Topic, error) {
	panic("not implemented yet")
}

func UseCaseTopic3_ListDocs(topic string, count int) ([]Doc, error) {
	panic("not implemented yet")
}

func UseCaseTopic4_ListTexts(topic string, count int) ([]Text, error) {
	panic("not implemented yet")
}
