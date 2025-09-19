package mllp

import (
	"bytes"
	"net"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	mockAck    = []byte("ack")
	encodedMsg = []byte{sb, 't', 'e', 's', 't', eb, cr}
	decodedMsg = []byte("test")
)

type mockSender struct {
	messages [][]byte
	mu       sync.Mutex
}

func (s *mockSender) Send(b []byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, b)
	return mockAck, nil
}

type transmission struct {
	inp      []byte
	wantAcks [][]byte
}

func TestReceiverE2E(t *testing.T) {
	tests := []struct {
		name     string
		trans    []transmission
		wantMsgs [][]byte
	}{
		{
			"one message",
			[]transmission{
				{encodedMsg, [][]byte{mockAck}},
			},
			[][]byte{decodedMsg},
		},
		{
			"two messages, two conns",
			[]transmission{
				{encodedMsg, [][]byte{mockAck}},
				{encodedMsg, [][]byte{mockAck}},
			},
			[][]byte{decodedMsg, decodedMsg},
		},
		{
			"two messages, one conn",
			[]transmission{
				{bytes.Join([][]byte{encodedMsg, encodedMsg}, nil), [][]byte{mockAck, mockAck}},
			},
			[][]byte{decodedMsg, decodedMsg},
		},
		{
			"one encoded, one decoded, two conns",
			[]transmission{
				{encodedMsg, [][]byte{mockAck}},
				{decodedMsg, nil},
			},
			[][]byte{decodedMsg},
		},
		{
			"one encoded, one decoded, one conn",
			[]transmission{
				{bytes.Join([][]byte{encodedMsg, decodedMsg}, nil), [][]byte{mockAck}},
			},
			[][]byte{decodedMsg},
		},
		{
			"ignored",
			[]transmission{
				{[]byte("ignore"), nil},
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, r := newMock(t)
			defer r.Stop()

			for _, tr := range tt.trans {
				conn := dial(t, r.port)

				_, err := conn.Write(tr.inp)
				require.NoError(t, err)

				rd := NewMsgReader(conn)
				for _, want := range tr.wantAcks {
					got, err := rd.Read()
					require.NoError(t, err)
					require.Equal(t, got, want)
				}

				conn.Close()
			}

			r.Wait()
			require.Equal(t, tt.wantMsgs, s.messages)
		})
	}
}

func TestMultiConn(t *testing.T) {
	s, r := newMock(t)
	defer r.Stop()

	conn1 := dial(t, r.port)
	conn2 := dial(t, r.port)
	conn3 := dial(t, r.port)

	Write(conn3, decodedMsg)
	conn2.Write(decodedMsg)
	Write(conn1, decodedMsg)

	for _, conn := range []net.Conn{conn1, conn2, conn3} {
		conn.Close()
	}

	r.Wait()
	want := [][]byte{decodedMsg, decodedMsg}
	require.Equal(t, want, s.messages)
}

func newMock(t *testing.T) (*mockSender, *Receiver) {
	t.Helper()

	s := &mockSender{}
	r, err := NewReceiver("127.0.0.1", 0, s)
	require.NoError(t, err)

	go func() {
		err := r.Run()
		require.NoError(t, err)
	}()
	return s, r
}

func dial(t *testing.T, port int) net.Conn {
	c, err := net.Dial("tcp", net.JoinHostPort("localhost", strconv.Itoa(port)))
	require.NoError(t, err)

	return c
}
