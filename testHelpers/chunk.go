package testHelpers

import (
	"net"
	"rtmp/chunk"
	"rtmp/message"
	"rtmp/rtmpconn"
	"testing"
	"time"
)

func AcceptTestChunk(t *testing.T) (string, chan chunk.Chunk) {
	t.Helper()
	address := "127.0.0.1:0"

	listener, _ := net.Listen("tcp", address)

	rtmpConn := rtmpconn.NewRtmpConn(nil, 128, 10*time.Second)

	chunks := make(chan chunk.Chunk)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			rtmpConn.Conn = conn
			receivedChunk, err := message.Accept(rtmpConn)
			if err != nil || receivedChunk == nil {
				continue
			}
			chunks <- *receivedChunk
		}
	}()
	return listener.Addr().String(), chunks
}
