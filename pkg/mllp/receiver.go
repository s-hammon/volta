package mllp

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
)

// Sender sends each HL7 message and returns the ACK
type sender interface {
	Send([]byte) ([]byte, error)
}

type Receiver struct {
	listener net.Listener
	sender   sender
	port     int

	wg     sync.WaitGroup
	closed chan struct{}
	once   sync.Once
}

func NewReceiver(ip string, port int, sender sender) (*Receiver, error) {
	host := net.JoinHostPort(ip, strconv.Itoa(port))
	l, err := net.Listen("tcp", host)
	if err != nil {
		return nil, fmt.Errorf("net.Listen: %v", err)
	}

	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		return nil, fmt.Errorf("could not cast %v to TCPAddr: %v", l.Addr(), err)
	}

	return &Receiver{
		listener: l,
		sender:   sender,
		port:     addr.Port,
		closed:   make(chan struct{}),
	}, nil
}

func (rec *Receiver) Run() error {
	for {
		conn, err := rec.listener.(*net.TCPListener).AcceptTCP()
		if err != nil {
			select {
			case <-rec.closed:
				return nil
			default:
				return fmt.Errorf("listener.AcceptTCP: %v", err)
			}
		}

		rec.wg.Go(func() { rec.handle(conn) })
	}
}

func (rec *Receiver) Stop() {
	rec.once.Do(func() {
		close(rec.closed)
		if err := rec.listener.Close(); err != nil {
			log.Printf("listener.Close: %v", err)
		}
		rec.Wait()
	})
}

func (rec *Receiver) Wait() {
	rec.wg.Wait()
}

func (rec *Receiver) handle(conn *net.TCPConn) {
	defer conn.Close()

	r := NewMsgReader(conn)
	for {
		b, err := r.Read()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Printf("failed to read message: %v", err)
			}
			return
		}

		ack, err := rec.sender.Send(b)
		if err != nil {
			log.Printf("sender.Send: %v", err)
			return
		}

		if err := Write(conn, ack); err != nil {
			log.Printf("failed to write ACK: %v", err)
			return
		}
	}
}
