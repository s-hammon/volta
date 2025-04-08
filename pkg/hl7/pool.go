package hl7

import (
	"bytes"
	"sync"
)

var pool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}
