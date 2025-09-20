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

func StartTestServer() *RtmpServer {
	address := "127.0.0.1:0"
	server := NewRtmpServer(address)
	go func() {
		server.Accept()
	}()
	return server
}

func TestStartServer(t *testing.T) {
	server := StartTestServer()
	_, err := net.Dial("tcp", server.listener.Addr().String())
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
	server := StartTestServer()
	server.DefaultNetworkTimeout = 1 * time.Second
	conn, _ := net.Dial("tcp", server.listener.Addr().String())
	_, err := conn.Write([]byte("test"))
	assert.Nil(t, err)
	time.Sleep(2 * time.Second)
	err = <-server.errors
	assert.NotNil(t, err)
}

func TestServerOneConnectionOnlyOneHandshake(t *testing.T) {
	server := StartTestServer()
	conn, err := net.Dial("tcp", server.listener.Addr().String())
	assert.Nil(t, err)
	_, err = handshake.Request(conn)
	assert.Nil(t, err)
	_, err = handshake.Request(conn)
	assert.NotNil(t, err)
}

func TestServerReceivesMultipleChunks(t *testing.T) {
	server := StartTestServer()
	conn, err := net.Dial("tcp", server.listener.Addr().String())
	assert.Nil(t, err)
	_, err = handshake.Request(conn)
	assert.Nil(t, err)
	for i := range 10 {
		chunkSent := chunk.NewChunk(
			*chunk.NewHeader(
				*chunk.NewBasicHeader(uint8(0), uint32(2)),
				*chunk.NewMessageHeader(uint32(0), uint32(4), uint8(0), uint32(123456)),
				uint32(0),
			),
			[]byte("test"),
		)
		_, err = conn.Write(chunkSent.Buffer())
		assert.Nil(t, err)
		select {
		case err = <-server.errors:
			assert.Nil(t, err)
		case <-time.After(10 * time.Millisecond):
			// No error received within timeout, assume success
		}
		fmt.Printf("Chunk: %d/n", i)
	}
}
