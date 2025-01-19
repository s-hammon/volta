package jsonhl7

import (
	"bytes"
	"fmt"
)

const commonDelims = "|^~\\&"

type Delimiters map[int]byte

func getDelimiters(delimField []byte) (Delimiters, error) {
	if len(delimField) != 5 {
		return nil, fmt.Errorf("expected 5 delimiters, got %d", len(delimField))
	}
	fieldDelim := delimField[0]
	if !bytes.ContainsAny([]byte{fieldDelim}, commonDelims) {
		return nil, fmt.Errorf("invalid field delimiter '%c' at index 0", fieldDelim)
	}

	delims := Delimiters{0: fieldDelim}
	used := map[byte]struct{}{fieldDelim: {}}
	for i, c := range delimField[1:] {
		if c == '\x00' {
			continue
		}
		if !bytes.ContainsAny([]byte{c}, commonDelims) {
			return nil, fmt.Errorf("invalid delimiter '%c' at index %d", c, i)
		}
		if _, ok := used[c]; ok {
			return nil, fmt.Errorf("duplicate delimiter '%c' at index %d", c, i)
		}
		delims[i+1] = c
	}

	delims[1], delims[2] = delims[2], delims[1]
	delims[3], delims[4] = delims[4], delims[3]
	return delims, nil
}
