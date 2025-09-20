package handshake

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AcceptTestHandshake() string {
	address := "127.0.0.1:0"
	listener, _ := net.Listen("tcp", address)
	go func() {
		conn, _ := listener.Accept()
		err := Accept(conn)
		if err != nil {
			panic(err)
		}
	}()
	return listener.Addr().String()
}

func TestGenerateTimestampChunk(t *testing.T) {
	chunk := GenerateTimestamp()
	assert.Equal(t, uint32(0), chunk.Timestamp)
	assert.Equal(t, uint32(0), chunk.Zero)
}

func TestHandshake(t *testing.T) {
	address := AcceptTestHandshake()
	conn, _ := net.Dial("tcp", address)
	hs, err := Request(conn)
	assert.Nil(t, err)
	// Use constructor to create expected version for assertion context
	expectedServerVersion := NewVersion(3)
	assert.Equal(t, expectedServerVersion.Version, hs.ServerVersion.Version)
	assert.Equal(t, uint32(0), hs.ServerTimestamp.Timestamp)
	assert.Equal(t, uint32(0), hs.ServerTimestamp.Zero)
	assert.Equal(t, hs.ClientTimestamp.Timestamp, hs.ServerEcho.Timestamp)
	assert.Equal(t, hs.ServerTimestamp.Random, hs.ClientEcho.Random)
}

func TestConstructors(t *testing.T) {
	// Prepare a deterministic random array for tests
	var rnd [1528]byte
	for i := range rnd {
		rnd[i] = byte(i % 251)
	}

	v := NewVersion(7)
	assert.Equal(t, uint8(7), v.Version)

	ts := NewTimestamp(10, 20, rnd)
	assert.Equal(t, uint32(10), ts.Timestamp)
	assert.Equal(t, uint32(20), ts.Zero)
	assert.Equal(t, rnd, ts.Random)

	e := NewEcho(30, 40, rnd)
	assert.Equal(t, uint32(30), e.Timestamp)
	assert.Equal(t, uint32(40), e.TimeStamp2)
	assert.Equal(t, rnd, e.Random)

	h := NewHandshake(v, v, ts, ts, e, e)
	assert.Equal(t, v, h.ClientVersion)
	assert.Equal(t, v, h.ServerVersion)
	assert.Equal(t, ts, h.ClientTimestamp)
	assert.Equal(t, ts, h.ServerTimestamp)
	assert.Equal(t, e, h.ClientEcho)
	assert.Equal(t, e, h.ServerEcho)
}
