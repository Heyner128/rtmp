package testutil

import (
	"net"
	"rtmp/chunk"
	"rtmp/conn"
	"rtmp/message"
	"testing"
	"time"
)

func AcceptTestChunk(t *testing.T) (string, chan chunk.Chunk) {
	t.Helper()
	address := "127.0.0.1:0"

	listener, _ := net.Listen("tcp", address)

	connection, _ := conn.NewConn(nil, 128, 10*time.Second)

	chunks := make(chan chunk.Chunk)

	go func() {
		for {
			netConnection, err := listener.Accept()
			if err != nil {
				continue
			}
			connection.Conn = netConnection
			receivedChunk, err := message.Accept(connection)
			if err != nil || receivedChunk == nil {
				continue
			}
			chunks <- *receivedChunk
		}
	}()
	return listener.Addr().String(), chunks
}
