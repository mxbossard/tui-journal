package service

import (
	"bytes"
	"fmt"
	"io"
	"iter"
	"os"
	"path/filepath"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/index"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/mxbossard/utilz/errorz"
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

const (
	DataBlocCapacity      = 256
	DataBlocThresholdSize = 100
	TypeStateLen          = 3
	FormStateLen          = 3
	KindStateLen          = 4
)

var (
	dump1 = ztring.LoremIpsumWords(10)
	dump2 = ztring.LoremIpsumWords(20)
	dump3 = ztring.LoremIpsumWords(30)

	// dumpState          = idx.BuildState(index.BucketIdxStateSize, "dmp")
	// dumpNewState       = idx.BuildState(index.BucketIdxStateSize, "dmp", "new ")
	// rootLayerState     = idx.BuildState(index.LayerIdxStateSize, "lyr", "root")
	// layerSnapshotState = idx.BuildState(index.BucketIdxStateSize, "lyr", "snap")
	// dumpLayerRootState = idx.BuildState(index.LayerIdxStateSize, "dmp", "lyr", "root")
	// dumpLayerDiffState = idx.BuildState(index.LayerIdxStateSize, "dmp", "lyr", "diff")

	dumpType  = idx.BuildState(TypeStateLen, "dmp")
	docType   = idx.BuildState(TypeStateLen, "doc")
	layerForm = idx.BuildState(FormStateLen, "lyr")
	rootKind  = idx.BuildState(KindStateLen, "root")
	diffKind  = idx.BuildState(KindStateLen, "diff")

	dumpState          = idx.CatState(index.BucketIdxStateSize, dumpType)
	docState           = idx.CatState(index.BucketIdxStateSize, docType)
	dumpRootLayerState = idx.CatState(index.LayerIdxStateSize, dumpType, layerForm, rootKind)
	dumpDiffLayerState = idx.CatState(index.LayerIdxStateSize, dumpType, layerForm, diffKind)
)

var dumpTypeFilter = idx.StateFilter(func(s idx.State) (bool, bool) {
	fmt.Printf("dumpTypeFilter s: %v / dumpType: %v\n", s, dumpType)
	return bytes.Equal(s[0:TypeStateLen], dumpType), true
})

var docTypeFilter = idx.StateFilter(func(s idx.State) (bool, bool) {
	return bytes.Equal(s[0:TypeStateLen], docType), true
})

var layerFormFilter = idx.StateFilter(func(s idx.State) (bool, bool) {
	fmt.Printf("layerFormFilter s: %v / dumpType: %v\n", s, layerForm)
	return bytes.Equal(s[TypeStateLen:TypeStateLen+FormStateLen], layerForm), true
})

var rootKindFilter = idx.StateFilter(func(s idx.State) (bool, bool) {
	fmt.Printf("rootKindFilter s: %v / dumpType: %v\n", s, rootKind)
	return bytes.Equal(s[TypeStateLen+FormStateLen:TypeStateLen+FormStateLen+KindStateLen], rootKind), true
})

var diffKindFilter = idx.StateFilter(func(s idx.State) (bool, bool) {
	return bytes.Equal(s[TypeStateLen+FormStateLen:TypeStateLen+FormStateLen+KindStateLen], diffKind), true
})

type idxService struct {
	bucketIdx       index.BucketIndex
	bucketByTimeIdx index.DocByTimeIndex
	layerIdx        index.LayerIndex
}

func NewIdxService(dir, device, salt string) (*idxService, error) {
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
	bucketByTimeIdx, err := index.NewCreationTimeIndex(bucketByTimeIdxDir, device, salt)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(layerIdxDir, 0700)
	if err != nil {
		return nil, err
	}
	layerIdx, err := index.NewLayerIndex(layerIdxDir, device, salt)
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

func GetBlocReader(ref *model.BlocRef) (*filez.Bloc, error) {
	bf, err := filez.OpenBlocsFile(ref.BlocsFilepath)
	if err != nil {
		return nil, err
	}
	bloc, err := bf.Get(ref.BlocId)
	return bloc, err
}

// FIXME: NEED to synchronize blocs writes & reads in a dedicated service (which may cache BlocsFile).
func GetBlocWriter(dir, device string) (*filez.BlocsFile, error) {
	qualifier := "data"
	dir = filepath.Join(dir, qualifier)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return nil, err
	}
	firstDeviceFilepath := filepath.Join(dir, fmt.Sprintf("%s-%s-001.idx", qualifier, device))
	dbf1, err := filez.NewBlocsFile(firstDeviceFilepath, DataBlocCapacity, DataBlocThresholdSize)
	if err != nil {
		return nil, err
	}

	return dbf1, nil
}

func readLayerDiff(ref *model.LayerRef) (string, error) {
	bloc, err := GetBlocReader(ref.BlocRef)
	if err != nil {
		return "", err
	}
	b := make([]byte, ref.Pos+ref.Len)
	n, err := bloc.Read(b)
	if err != nil {
		return "", err
	}
	txt := string(b[ref.Pos:n])
	// fmt.Printf("read layer diff: [%s] from pos: %d of len: %d\n", txt, ref.Pos, n-ref.Pos)
	return txt, nil
}

func project(b *model.Bucket) (txt string, err error) {
	// FIXME: implements diff aggregation
	for err2, entry := range b.LayerRefIt {
		if err2 != nil {
			return "", err2
		}
		txt, err = readLayerDiff(entry.Val())
		if err != nil && err != io.EOF {
			return
		}
	}
	return txt, nil
	// panic("not implemented yet")
}

// ------------- Dumps ---------------

func UseCaseDump0_Create(dir, salt, device, txt string) (*Dump, error) {
	idxService, err := NewIdxService(dir, device, salt)
	if err != nil {
		return nil, err
	}
	now := time.Now()

	// 0- Forge dump name
	name := ForgeDumpName(device, now)

	// FIXME: 1- Check if bucket already exists !

	// 2- Create a bucket
	err = idxService.bucketIdx.Add(dumpState, now, nil, name)
	if err != nil {
		return nil, err
	}

	// 3- Create a bucket-time idx entry
	// idxService.bucketByTimeIdx.Add(dumpNewState, now, []byte(name))

	// 4- Store the content
	blocWriter, err := GetBlocWriter(dir, device)
	if err != nil {
		return nil, err
	}
	data := []byte(txt)
	n, err := blocWriter.Write(data)
	if err != nil {
		return nil, err
	}
	bloc, err := blocWriter.GetLastBloc()
	if err != nil {
		return nil, err
	}

	blocRef := model.BlocRef{
		BlocsFilepath: blocWriter.Name(),
		BlocId:        bloc.Uid.Id,
	}

	pos := bloc.Len() - n
	rootLayerRef := &model.LayerRef{
		BlocRef: &blocRef,
		Pos:     pos,
		Len:     n,
	}
	// fmt.Printf("Written data: %v of len: %d at pos: %d\n", data, n, pos)

	// 5- Create a layer idx entry
	hUid := model.HashedBucketUid(index.StringToBucketUid(name))
	idxService.layerIdx.Add(dumpRootLayerState, now, &hUid, rootLayerRef)

	// rootLayer := model.Layer{
	// 	Metadata: model.LayerMetadata{
	// 		Version: 0, // First layer
	// 		Created: now,
	// 	},
	// 	Commited:   false,
	// 	Snapshoted: true, // First layer snapshoted by definition
	// }

	// 6- Forge the root layer iterator
	var layerRefIt iter.Seq2[error, idx.Entry[*model.HashedBucketUid, *model.LayerRef]] = func(yield func(error, idx.Entry[*model.HashedBucketUid, *model.LayerRef]) bool) {
		entry := idx.NewEntry(&hUid, rootLayerRef, -1, now, dumpRootLayerState, nil)
		yield(nil, entry)
	}

	// 7- Build the entity to return
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

func ByteSliceInArray(a [][]byte, s []byte) bool {
	for _, slice := range a {
		if bytes.Equal(s, slice) {
			return true
		}
	}
	return false
}

// Must return a list of dumps able to lazy load their layers.
// Find count dump root layers
func UseCaseDump1_ListLast(dir, device, salt string, count int) ([]*Dump, error) {
	// 1- Browse bucket-time idx to find last dumps ref
	idxService, err := NewIdxService(dir, device, salt)
	if err != nil {
		return nil, err
	}
	// dumpStateFilter := idx.StateFilter(func(s idx.State) (bool, bool) {
	// 	fmt.Printf("dumpStateFilter s: %v / dumpState: %v\n", s[0:3], dumpState[0:3])
	// 	return bytes.Equal(s[0:3], dumpState[0:3]), true
	// })

	// 2- Find last buckets layers in bucketByTimeIdx
	// 	k := 0
	// 	var bucketRhUids [][]byte
	// 	paginer, errChan := idxService.bucketByTimeIdx.FilterAll(idx.BottomToTop, dumpStateFilter)
	// BucketLoop:
	// 	for err, page := range paginer.All() {
	// 		// FIXME: what is this error ?
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		if err := errorz.ChanCollect(errChan); err.GotError() {
	// 			return nil, err
	// 		}
	// 		// FIXME what is this pos ? is it seq ?
	// 		for pos, entry := range page.All() {
	// 			_ = pos
	// 			if k == count {
	// 				// FIXME: must break only if count DISTINCT bucket uids reached.
	// 				break BucketLoop
	// 			}
	// 			fmt.Printf("ranging over bucketByTimeIdx, hUid: %v\n", entry.Val())
	// 			// var array [128]byte
	// 			// copy(array[:], entry.Val())
	// 			// array := index.ByteSliceToBucketUid()
	// 			bucketRhUids = append(bucketRhUids, entry.Val())
	// 			k++
	// 		}
	// 	}

	// 	fmt.Printf("found last hashed bucket uids: %v\n", bucketRhUids)

	// 3- Find corresponding layers in layerIdx
	// var snapshotedLayersRhUids [][]byte
	// layerKeyFilter := idx.KeyFilter(func(k []byte, s idx.State) (bool, bool) {
	// 	// hashedUid := index.HashedBucketUid(k)
	// 	fmt.Printf("layerKeyFilter k: %v / state: %v / bucketRhUids: %v\n", k, s, bucketRhUids)
	// 	for _, hUid := range bucketRhUids {
	// 		if bytes.Equal(hUid, k) {
	// 			if ByteSliceInArray(snapshotedLayersRhUids, hUid) {
	// 				// RhUid already seen go on looping
	// 				return false, true
	// 			}
	// 			if bytes.Equal(s, layerSnapshotState) {
	// 				// Layer is a snapshot
	// 				snapshotedLayersRhUids = append(snapshotedLayersRhUids, hUid)
	// 			}
	// 			return true, true
	// 		}
	// 	}
	// 	return false, true
	// })

	k := 0
	var dumps []*Dump
	layersByRhUid := make(map[[128]byte][]*model.LayerRef)
	paginer2, errChan := idxService.layerIdx.FilterAll(idx.BottomToTop, idx.AndFilter(dumpTypeFilter, layerFormFilter, rootKindFilter))
Loop:
	for err, page := range paginer2.Pages() {
		// FIXME: what is this error ?
		if err != nil {
			return nil, err
		}
		if err := errorz.ChanCollect(errChan); err.GotError() {
			return nil, err
		}
		for pos, entry := range page.All() {
			_ = pos
			if k == count {
				break Loop
			}
			fmt.Printf("ranging over layerIdx, layerRef: %v\n", entry.Val())
			layersByRhUid[*entry.Key()] = append(layersByRhUid[*entry.Key()], entry.Val())
			k++
		}
	}

	for rhUid, layerRefs := range layersByRhUid {
		// FIXME: need to find UID
		// FIXME: do not have metadata here ? Where are stored metadatas ? Do we need Metadata before projecting the document ?
		_ = layerRefs
		layerPager, errChan := idxService.layerIdx.Paginate((*model.HashedBucketUid)(rhUid[:]), idx.BottomToTop)
		_ = layerPager
		// FIXME: what to do with errChan ?
		if err := errorz.ChanCollect(errChan); err.GotError() {
			return nil, err
		}
		d := &Dump{
			Bucket: model.Bucket{
				Uid:        model.BucketUid((rhUid[:])),
				LayerRefIt: layerPager.All(),
			},
		}
		dumps = append(dumps, d)
	}

	return dumps, nil
	// panic("not implemented yet")
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
