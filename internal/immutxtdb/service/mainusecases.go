package service

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/bucket"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/index"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/model"
	"github.com/mxbossard/tui-journal/internal/immutxtdb/store"
	"github.com/mxbossard/utilz/filez"
	"github.com/mxbossard/utilz/ztring"
)

// Aggregation of layers
type Dump struct {
	*bucket.Bucket
}

type Doc struct {
	*bucket.Bucket
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

	dumpType  = idx.BuildStringState(TypeStateLen, "dmp")
	docType   = idx.BuildStringState(TypeStateLen, "doc")
	layerForm = idx.BuildStringState(FormStateLen, "lyr")
	rootKind  = idx.BuildStringState(KindStateLen, "root")
	diffKind  = idx.BuildStringState(KindStateLen, "diff")

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

func ForgeDumpName(when time.Time) string {
	return fmt.Sprintf("dump-%d", when.Unix())
}

func ForgeDumpUid(when time.Time) bucket.BucketUid {
	name := ForgeDumpName(when)
	uid := bucket.DeterministicUid(name)
	return uid
}

func GetBlocReader(ref *model.BlocRef) (*filez.Bloc, error) {
	bf, err := filez.OpenBlocsFile(ref.BlocsFilepath)
	if err != nil {
		return nil, err
	}
	bloc, err := bf.Get(ref.BlocId)
	if err != nil {
		return nil, fmt.Errorf("error getting bloc #%d of file [%s]: %w", ref.BlocId, ref.BlocsFilepath, err)
	}
	return bloc, nil
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
	length := ref.Pos + ref.Len
	b := make([]byte, length)
	n, err := bloc.Read(b)
	if err != nil {
		return "", fmt.Errorf("error reading diff of lenth %d: %w", length, err)
	}
	txt := string(b[ref.Pos:n])
	// fmt.Printf("read layer diff: [%s] from pos: %d of len: %d\n", txt, ref.Pos, n-ref.Pos)
	return txt, nil
}

func getStore(dir, salt, device string) (*store.TwoPhasesStore, error) {
	// Get store
	eDir := filepath.Join(dir, "ephemeral")
	rDir := filepath.Join(dir, "rested")
	s, err := store.NewTwoPhasesStore(eDir, rDir, device, salt)
	return s, err
}

// ------------- Dumps ---------------

func UseCaseDump0_Create(dir, salt, device, txt string) (*Dump, error) {
	s, err := getStore(dir, salt, device)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// 0- Forge dump name
	bUid := ForgeDumpUid(now)
	name := ForgeDumpName(now)
	var labels bucket.Labels

	// FIXME: do we want each device to create a different bucket for the same daily Dump ?
	// FIXME: do we want each device share the same bucket for same daily Dump ?
	// TODO: we want a deterministic way to build the daily Dump bucket uid because bucket manage partitions for us.
	// We want a way to pass a bucketUid to create a Bucket. We need to pass a name with it.
	// Do we need to check if the bucketUid already exists ? Or do we return the bucket if it already exists ?
	//     => We could use CreateOrGet(uid, name, labels) returning the existing bucket should the bucket be renamed ?
	//        If the bucket we want to create already exists but have a different name, could return an error to prevent errors
	// Bucket pkg SHOULD supply a way to generate good bucketUid supplying a Name
	// If so, should we allow user to supply it's own bucketUid ?

	// 1- Check if bucket already exists !
	bkt, err := s.Get(bUid)
	if err == store.ErrNotExist {
		// swallow the error
		err = nil
	} else if err != nil {
		return nil, err
	}

	// 2- Create a bucket
	if bkt == nil {
		// bkt do not exists yet
		bkt = s.CreateOrGetBucket(bUid, name, labels)
	}

	// 3- Store the content
	err = bkt.SetText(txt)
	if err != nil {
		return nil, err
	}
	err = s.Save(bkt)
	if err != nil {
		return nil, err
	}

	// 4- Build the entity to return
	d := &Dump{
		Bucket: bkt,
	}
	return d, nil
}

func byteSliceInArray0(a [][]byte, s []byte) bool {
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
	s, err := getStore(dir, salt, device)
	if err != nil {
		return nil, err
	}

	paginer, err := s.Filter(bucket.YoungerFirst, nil, count, 1)
	if err != nil {
		return nil, err
	}

	var dumps []*Dump
	for err, page := range paginer.Pages() {
		if err != nil {
			return nil, err
		}
		for _, entry := range page.Entries() {
			if entry.Error() != nil {
				return nil, err
			}
			d := &Dump{
				Bucket: entry.Val(),
			}
			dumps = append(dumps, d)
		}
		// Parse only first page
		break
	}

	return dumps, nil
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
