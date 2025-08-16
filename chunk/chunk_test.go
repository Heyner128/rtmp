package chunk

import (
	"encoding/binary"
	"math"
	"math/rand"
	"miistream/handshake"
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
		err := handshake.Accept(conn)
		if err != nil {
			panic(err)
		}
	}()
	_, err := handshake.Request(address)
	if err != nil {
		panic(err)
	}
	m.Run()
}

func TestNewSetSizeChunk(t *testing.T) {
	const timestamp = 0
	const size = 128
	setSizeChunk := NewSetSizeChunk(uint32(timestamp), uint32(size))
	assert.Equal(t, []byte{0x02 << 6}, setSizeChunk.Header.BasicHeader)
	assert.Equal(t, binary.BigEndian.AppendUint32(make([]byte, 0), timestamp)[1:], setSizeChunk.Header.MessageHeader[:3])
	assert.Equal(t, binary.BigEndian.AppendUint32(make([]byte, 0), 32)[1:], setSizeChunk.Header.MessageHeader[3:6])
	assert.Equal(t, byte(0), setSizeChunk.Header.MessageHeader[6])
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00}, setSizeChunk.Header.MessageHeader[7:11])
	assert.Equal(t, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(size)), setSizeChunk.Data)
}

func TestNewSetSizeChunkExtendedTimestamp(t *testing.T) {
	extendedTimestamp := []byte{0x7F, 0xFF, 0xFF, 0xFF}
	setSizeChunk := NewSetSizeChunk(binary.BigEndian.Uint32(extendedTimestamp), uint32(0))
	assert.Equal(t, []byte{0xFF, 0xFF, 0xFF}, setSizeChunk.Header.MessageHeader[:3])
	assert.Equal(t, extendedTimestamp, setSizeChunk.Header.ExtendedTimestamp)
}

func TestNewSetSizeChunkSizeFirstBitZero(t *testing.T) {
	const size = math.MaxUint32
	setSizeChunk := NewSetSizeChunk(uint32(0), uint32(size))
	assert.Equal(t, uint32(0), binary.BigEndian.Uint32(setSizeChunk.Data)>>31)
}

func TestNewSetSizeChunkSizeZero(t *testing.T) {
	const size = 0
	setSizeChunk := NewSetSizeChunk(uint32(0), uint32(size))
	assert.Equal(t, uint32(128), binary.BigEndian.Uint32(setSizeChunk.Data))
}

func TestNewSetSizeChunkSizeNoLargerThanMaxMessageSize(t *testing.T) {
	maxMessageSize := binary.BigEndian.Uint32([]byte{0x00, 0xFF, 0xFF, 0xFF})
	const size = math.MaxUint32
	setSizeChunk := NewSetSizeChunk(uint32(0), size)
	assert.Equal(t, maxMessageSize, binary.BigEndian.Uint32(setSizeChunk.Data))
}
