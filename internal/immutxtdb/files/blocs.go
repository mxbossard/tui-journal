package files

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mxbossard/utilz/filez"
)

const (
	DataBlocCapacity      = 256
	DataBlocThresholdSize = 100
)

type BlocRef struct {
	BlocsFilepath string
	BlocId        int
}

type BlocRefPart struct {
	*BlocRef
	Pos int
	Len int
}

type MetadataRef BlocRefPart
type LayerRef BlocRefPart
type HeaderRef BlocRefPart

// FIXME: NEED to synchronize blocs writes & reads in a dedicated service (which may cache BlocsFile).
func GetBlocWriter(dir, device, qualifier string) (*filez.BlocsFile, error) {
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

func StoreBytes(dir, device, qualifier string, data []byte) (*BlocRefPart, error) {
	blocWriter, err := GetBlocWriter(dir, device, qualifier)
	if err != nil {
		return nil, err
	}
	p, err := blocWriter.Write(data)
	if err != nil {
		return nil, err
	}
	n := len(data)
	if p != n {
		return nil, fmt.Errorf("written %d bytes instead of %d in bloc %s", p, n, blocWriter)
	}
	bloc, err := blocWriter.GetLastBloc()
	if err != nil {
		return nil, err
	}
	blocRefPart := BlocRefPart{
		BlocRef: &BlocRef{
			BlocsFilepath: blocWriter.Name(),
			BlocId:        bloc.Uid.Id,
		},
	}
	return &blocRefPart, nil
}
