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

// type BlocRef struct {
// 	BlocsFilepath string
// 	BlocId        int
// }

// type BlocRefPart struct {
// 	*BlocRef
// 	Pos int
// 	Len int
// }

type MetadataRef filez.BlocPart
type LayerRef filez.BlocPart
type HeaderRef filez.BlocPart

// Get a BlocWriter to write data
// FIXME: NEED to synchronize blocs writes & reads in a dedicated service (which may cache BlocsFile).
func GetBlocWriter(dir, partition, qualifier string) (*filez.BlocWriter, error) {
	dir = filepath.Join(dir, qualifier)
	err := os.MkdirAll(dir, 0700)
	if err != nil {
		return nil, fmt.Errorf("unable to get bloc writer [%s] [%s] [%s]: %w", dir, partition, qualifier, err)
	}
	firstPartitionFilepath := filepath.Join(dir, fmt.Sprintf("%s-%s-001.idx", qualifier, partition))
	dbf1, err := filez.NewBlocsFile(firstPartitionFilepath, DataBlocCapacity, DataBlocThresholdSize)
	if err != nil {
		return nil, fmt.Errorf("unable to get bloc writer [%s] [%s] [%s]: %w", dir, partition, qualifier, err)
	}

	return dbf1.Writer(), nil
}

// Get a Bloc to read data
func GetBloc(path string, blocId int) (*filez.Bloc, error) {
	dbf1, err := filez.OpenBlocsFile(path)
	if err != nil {
		return nil, fmt.Errorf("unable to get bloc reader [%s#%d]: %w", path, blocId, err)
	}
	return dbf1.Get(blocId)
}

func StoreBlocData(dir, partition, qualifier string, data []byte) (*filez.VirtualBloc, error) {
	blocWriter, err := GetBlocWriter(dir, partition, qualifier)
	if err != nil {
		return nil, fmt.Errorf("unable to store blocRef [%s] [%s] [%s]: %w", dir, partition, qualifier, err)
	}

	p, err := blocWriter.Write(data)
	if err != nil {
		return nil, fmt.Errorf("unable to store blocRef [%s] [%s] [%s]: %w", dir, partition, qualifier, err)
	}
	n := len(data)
	if p != n {
		return nil, fmt.Errorf("written %d bytes instead of %d", p, n)
	}

	return blocWriter.WritenBloc(), nil
}

func LoadBlocPart(brp *filez.BlocPart) ([]byte, error) {
	br, err := GetBloc(brp.Uid.Filepath, brp.Uid.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to load blocRef [%s#%d] [%d/%d]: %w", brp.Uid.Filepath, brp.Uid.Id, brp.Pos, brp.Len, err)
	}
	b := br.Bytes()
	return b[brp.Pos : brp.Pos+brp.Len], nil
}
