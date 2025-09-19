package mllp

import (
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/s-hammon/p"
	"github.com/stretchr/testify/require"
)

func TestSenderE2E(t *testing.T) {
	s, r := newMock(t)
	defer r.Stop()

	go func() {
		err := r.Run()
		require.NoError(t, err)
	}()
	time.Sleep(100 * time.Millisecond)

	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(r.port))

	sender, err := NewSender(addr, 1)
	require.NoError(t, err)
	defer sender.Close()

	ack, err := sender.Send(decodedMsg)
	require.NoError(t, err)
	require.Equal(t, mockAck, ack)

	s.mu.Lock()
	require.Len(t, s.messages, 1)
	require.Equal(t, decodedMsg, s.messages[0])
	s.mu.Unlock()
}

func TestSenderMultiple(t *testing.T) {
	s, r := newMock(t)
	defer r.Stop()

	go func() {
		err := r.Run()
		require.NoError(t, err)
	}()
	time.Sleep(100 * time.Millisecond)

	addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(r.port))

	sender, err := NewSender(addr, 1)
	require.NoError(t, err)
	defer sender.Close()

	for range 5 {
		ack, err := sender.Send(decodedMsg)
		require.NoError(t, err)
		require.Equal(t, mockAck, ack)
	}

	s.mu.Lock()
	require.Len(t, s.messages, 5)
	for _, m := range s.messages {
		require.Equal(t, decodedMsg, m)
	}

	s.mu.Unlock()
}

func TestSendError(t *testing.T) {
	_, err := NewSender("127.0.0.1:65535", 1)
	require.Error(t, err)
}

var benchMsg = make([]byte, 5600) // represents the average size of an ORU

func init() {
	for i := range benchMsg {
		benchMsg[i] = 'X'
	}
}

func BenchmarkSendParallel(b *testing.B) {
	for _, size := range []int{1, 2, 4, 8, 16} {
		b.Run(p.Format("pool=%d", size), func(b *testing.B) {
			s := &mockSender{}
			r, err := NewReceiver("127.0.0.1", 0, s)
			require.NoError(b, err)
			defer r.Stop()

			go func() {
				err := r.Run()
				require.NoError(b, err)
			}()
			time.Sleep(100 * time.Millisecond)

			addr := net.JoinHostPort("127.0.0.1", strconv.Itoa(r.port))
			sender, err := NewSender(addr, size)
			require.NoError(b, err)
			defer sender.Close()

			b.SetBytes(int64(len(benchMsg)))
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, err := sender.Send(benchMsg)
					require.NoError(b, err)
				}
			})
		})
	}
}
