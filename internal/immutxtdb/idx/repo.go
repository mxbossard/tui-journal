package idx

type repo struct {
	dirPath    string
	salt       string
	passphrase string
	encoder    IdxEncoder
}
