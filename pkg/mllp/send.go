package mllp

import (
	"fmt"
	"net"
	"sync"
)

type Sender struct {
	outs []*out
	mu   sync.Mutex
	next int
}

type out struct {
	conn net.Conn
	mu   sync.Mutex
}

func NewSender(addr string, size int) (*Sender, error) {
	sender := &Sender{}
	for range size {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			sender.Close()
			return nil, err
		}

		sender.outs = append(sender.outs, &out{conn: conn})
	}

	return sender, nil
}

func (s *Sender) Send(msg []byte) ([]byte, error) {
	out := s.getNext()
	out.mu.Lock()
	defer out.mu.Unlock()

	conn := out.conn
	if err := Write(conn, msg); err != nil {
		return nil, fmt.Errorf("mllp.Write: %v", err)
	}

	return NewMsgReader(conn).Read()
}

func (s *Sender) getNext() *out {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := s.outs[s.next]
	s.next = (s.next + 1) % len(s.outs)
	return out
}

func (s *Sender) Close() {
	for _, out := range s.outs {
		out.conn.Close()
	}
}
