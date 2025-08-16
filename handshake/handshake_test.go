package handshake

import (
	"math/rand"
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

var address = "127.0.0.1:" + strconv.Itoa(1000+rand.Intn(9999-1000))

var listener, _ = net.Listen("tcp", address)

func TestMain(m *testing.M) {
	defer listener.Close()
	go func() {
		conn, _ := listener.Accept()
		err := Accept(conn)
		if err != nil {
			panic(err)
		}
	}()
	m.Run()
}

func TestGenerateTimestampChunk(t *testing.T) {
	chunk := GenerateTimestamp()
	assert.Equal(t, uint32(0), chunk.Timestamp)
	assert.Equal(t, uint32(0), chunk.Zero)
}

func TestHandshake(t *testing.T) {
	handshake, err := Request(address)
	assert.Nil(t, err)
	assert.Equal(t, uint8(3), handshake.ServerVersion.Version)
	assert.Equal(t, uint32(0), handshake.ServerTimestamp.Timestamp)
	assert.Equal(t, uint32(0), handshake.ServerTimestamp.Zero)
	assert.Equal(t, handshake.ClientTimestamp.Timestamp, handshake.ServerEcho.Timestamp)
	assert.Equal(t, handshake.ServerTimestamp.Random, handshake.ClientEcho.Random)
}
