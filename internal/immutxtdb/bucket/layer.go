package bucket

import (
	"bytes"

	"github.com/mxbossard/tui-journal/internal/immutxtdb/idx"
)

type Layer struct {
	Metadata *Metadata
	Content  []byte
	State    idx.State
}

func (l *Layer) Equals(l2 *Layer) bool {
	if l2 == nil {
		return false
	}
	if !bytes.Equal(l.Content, l2.Content) {
		return false
	}
	if !bytes.Equal(l.State, l2.State) {
		return false
	}
	if l.Metadata != nil {
		if l2.Metadata == nil {
			return false
		}
		if l.Metadata.Size != l2.Metadata.Size {
			return false
		}
		if l.Metadata.Updated != nil {
			if l2.Metadata.Updated == nil {
				return false
			}
			if l.Metadata.Updated.Compare(*l2.Metadata.Updated) != 0 {
				return false
			}
		}
		if l.Metadata.Version != l2.Metadata.Version {
			return false
		}
		if l.Metadata.Uid != l2.Metadata.Uid {
			return false
		}
	}

	return true
}
