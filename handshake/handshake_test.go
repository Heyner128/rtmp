package handshake

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateTimestamp(t *testing.T) {
	timestamp := GenerateTimestamp()
	assert.Equal(t, uint32(0), timestamp.Timestamp)
	assert.Equal(t, uint32(0), timestamp.Zero)
}

func TestHandshake(t *testing.T) {
	address := acceptTestHandshake(t)
	conn, _ := net.Dial("tcp", address)
	hs, err := RequestTestHandshake(t, conn)
	assert.Nil(t, err)
	expectedServerVersion := newVersion(3)
	assert.Equal(t, expectedServerVersion.Version, hs.ServerVersion.Version)
	assert.Equal(t, uint32(0), hs.ServerTimestamp.Timestamp)
	assert.Equal(t, uint32(0), hs.ServerTimestamp.Zero)
	assert.Equal(t, hs.ClientTimestamp.Timestamp, hs.ServerEcho.Timestamp)
	assert.Equal(t, hs.ServerTimestamp.Random, hs.ClientEcho.Random)
}
