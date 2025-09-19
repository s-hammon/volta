package mllp

import (
	"bufio"
	"fmt"
	"io"
)

const (
	sb byte = 0x0B
	eb byte = 0x1C
	cr byte = 0x0D
)

type MsgReader struct {
	r *bufio.Reader
}

func NewMsgReader(rd io.Reader) *MsgReader {
	r := bufio.NewReader(rd)
	return &MsgReader{r}
}

// Reads content from buffer & decodes content from MLLP encoding
func (mr *MsgReader) Read() ([]byte, error) {
	if _, err := mr.r.ReadBytes(sb); err != nil {
		return nil, err
	}

	b, err := mr.r.ReadBytes(eb)
	if err != nil {
		return nil, err
	}

	end, err := mr.r.ReadByte()
	if err != nil {
		return nil, err
	}
	if end != cr {
		if err := mr.r.UnreadByte(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("expecting %c, got %d", cr, end)
	}

	return b[:len(b)-1], nil
}

// Write contents in MLLP encoding (SB + data + EB + CR)
func Write(w io.Writer, b []byte) error {
	if _, err := w.Write([]byte{sb}); err != nil {
		return fmt.Errorf("w.Write(sb): %v", err)
	}
	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("w.Write(data): %v", err)
	}
	if _, err := w.Write([]byte{eb, cr}); err != nil {
		return fmt.Errorf("w.Write(eb, cr): %v", err)
	}

	return nil
}

// Reads a singular message, used to read ACK from a sending conn
func ReadMsg(rd io.Reader) ([]byte, error) {
	return NewMsgReader(rd).Read()
}
