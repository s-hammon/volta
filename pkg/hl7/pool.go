package hl7

import (
	"bytes"
	"sync"
)

var pool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}
