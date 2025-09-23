package server

import (
	"fmt"
	"net"
	"rtmp/chunk"
	"rtmp/handshake"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStartServer(t *testing.T) {
	server := StartTestingServer(t)
	_, err := net.Dial("tcp", server.Listener.Addr().String())
	assert.Nil(t, err)
}

func TestStartServerFail(t *testing.T) {
	invalidAddress := "notanip:0000"
	go func() {
		assert.Panics(t, func() { NewRtmpServer(invalidAddress) })
	}()
	_, err := net.Dial("tcp", invalidAddress)
	assert.NotNil(t, err)
}

func TestServerDefaultSettings(t *testing.T) {
	server := NewRtmpServer("127.0.0.1:0")
	assert.Equal(t, uint32(128), server.DefaultMaxChunkSize)
	assert.Equal(t, 10*time.Second, server.DefaultNetworkTimeout)
}

func TestServerNetworkTimeout(t *testing.T) {
	server := StartTestingServer(t)
	server.DefaultNetworkTimeout = 1 * time.Second
	conn, _ := net.Dial("tcp", server.Listener.Addr().String())
	_, err := conn.Write([]byte("test"))
	assert.Nil(t, err)
	serverConn := <-server.Connections
	time.Sleep(2 * time.Second)
	err = <-serverConn.Errors
	assert.NotNil(t, err)
}

func TestServerOneConnectionOnlyOneHandshake(t *testing.T) {
	server := StartTestingServer(t)
	conn, err := net.Dial("tcp", server.Listener.Addr().String())
	assert.Nil(t, err)
	_, err = handshake.RequestTestHandshake(t, conn)
	assert.Nil(t, err)
	_, err = handshake.RequestTestHandshake(t, conn)
	assert.NotNil(t, err)
}

func TestServerReceivesMultipleChunks(t *testing.T) {
	server, conn := StartTestingServerWithHandshake(t)
	for i := range 10 {
		chunkSent := chunk.NewChunk(
			*chunk.NewHeader(
				*chunk.NewBasicHeader(uint8(0), uint32(2)),
				*chunk.NewMessageHeader(uint32(0), uint32(4), uint8(0), uint32(123456)),
				uint32(0),
			),
			[]byte("test"),
		)
		_, err := conn.Write(chunkSent.Buffer(t))
		assert.Nil(t, err)
		select {
		case serverConn := <-server.Connections:
			select {
			case err = <-serverConn.Errors:
				assert.Nil(t, err)
			case <-time.After(10 * time.Millisecond):
			}
		case <-time.After(10 * time.Millisecond):
		}
		fmt.Printf("Chunk: %d\n", i)
	}
}
