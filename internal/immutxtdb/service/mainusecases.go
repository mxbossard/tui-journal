package service

import (
	"bytes"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/index"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/mxbossard/utilz/ztring"
)

// Aggregation of layers
type Dump struct {
	Uid       string
	Metadata  model.Metadata
	LayerRefs []*model.LayerRef
}

type Doc struct {
	Uid      string
	Metadata model.Metadata
}

type Layer string

type Text string

type Topic string

/*
- What is the difference between dump and doc ?
  - Dump is an immutable journal log entry attached to only one day. Implemented by a collection of layers.
  - Doc is a named & editable document. Implemented by a collection of layers.
  - => No difference in implementation ?

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

func NewIdxService() *idxService {
	return &idxService{}
}

func ForgeDumpName(device string, when time.Time) string {
	return fmt.Sprintf("dump-%s-%d", device, when.Unix())
}

func RotatingHashString(s string) ([]byte, error) {
	panic("not implemented yet")
}

func GetBlocWriter(device string) (*model.BlocRef, io.Writer, int, error) {
	//blocs.GetLastBloc()
	panic("not implemented yet")
}

// ------------- Dumps ---------------

func UseCaseDump0_Create(device, txt string) (*Dump, error) {
	idxService := NewIdxService()
	now := time.Now()

	// 0- Forge dump name
	name := ForgeDumpName(device, now)

	// 1- Create a bucket
	err := idxService.bucketIdx.Add(dumpNewState, nil, name)
	if err != nil {
		return nil, err
	}

	// 2- Create a bucket-time idx entry
	bucketRhUid, err := RotatingHashString(name)
	if err != nil {
		return nil, err
	}
	idxService.bucketByTimeIdx.Add(dumpNewState, now, bucketRhUid)

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

	d := &Dump{
		Uid: name,
		Metadata: model.Metadata{
			Version:    0, // First layer
			Created:    now,
			Updated:    now,
			Labels:     nil,
			Commited:   false,
			Snapshoted: true, // First layer snapshoted by definition
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
	idxService := NewIdxService()
	dumpStateFilter := func(s idx.State, stop func()) bool {
		return bytes.Equal(s[0:3], dumpState[0:3])
	}
	k := 0
	var bucketRhUids [][128]byte
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
			array := [128]byte(entry.Val())
			bucketRhUids = append(bucketRhUids, array)
		}
	}

	layerStateFilter := func(s idx.State, stop func()) bool {
		return bytes.Equal(s[0:3], layerState[0:3])
	}
	var snapshotedLayersRhUids [][128]byte
	layerFilter := func(k [128]byte, s idx.State, stop func()) bool {
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
			layersByRhUid[[128]byte(entry.Key())] = append(layersByRhUid[[128]byte(entry.Key())], entry.Val())
		}
	}

	for rhUid, layerRefs := range layersByRhUid {
		// FIXME: need to find UID
		// FIXME: do not have metadata here ? Where are stored metadatas ? Do we need Metadata before projecting the document ?
		d := &Dump{
			Uid:       string(rhUid[:]),
			LayerRefs: layerRefs,
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
