package immutxtdb

type Bloc struct {
}

type FileEntry struct {
	Filepath string
	Content  string
	Pos      int
	Len      int
}

type BlocFile struct {
	entries []FileEntry
}

func (f BlocFile) ReadEntry(passphrase string, n int) FileEntry {
	panic("not implemented yet")
}

func (f BlocFile) WriteEntry(passphrase string, n int) FileEntry {
	panic("not implemented yet")
}

func (f BlocFile) ReadBloc(passphrase string, n, limit int) []Bloc {
	panic("not implemented yet")
}
