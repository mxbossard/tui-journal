package index

import "github.com/mxbossard/tui-journal/internal/immutxtdb/model"

const (
	indexLineBufferSize        = 1000
	blocBufferSize             = 1000
	delimiterChar              = ','
	newLineChar                = '\n'
	defaultPageSize            = 10
	asciiEncoderStateSize      = 8
	asciiEncoderDataSize       = 80
	asciiEncoderDefaultVersion = 0
)

var (
	Document = model.BuildState(asciiEncoderStateSize, "document")
	Dump     = model.BuildState(asciiEncoderStateSize, "dump")
)
