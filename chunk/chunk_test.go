package chunk

import (
	"encoding/binary"
	"math"
	"net"
	"rtmp/rtmpconn"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func AcceptTestChunk() (string, chan Chunk) {
	address := "127.0.0.1:0"

	listener, _ := net.Listen("tcp", address)

	rtmpConn := rtmpconn.NewRtmpConn(nil, 128, 10*time.Second)

	chunks := make(chan Chunk)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}
			rtmpConn.Conn = conn
			chunk, err := ReadChunk(rtmpConn)
			if err != nil || chunk == nil {
				continue
			}
			chunks <- *chunk
		}
	}()
	return listener.Addr().String(), chunks
}

func TestType0Chunk(t *testing.T) {
	address, chunks := AcceptTestChunk()
	conn, _ := net.Dial("tcp", address)
	basicHeader := NewBasicHeader(uint8(0), uint32(2))
	messageHeader := NewMessageHeader(uint32(0), uint32(32), uint8(1), uint32(123456))
	header := NewHeader(*basicHeader, *messageHeader, uint32(0))
	chunk := NewChunk(*header, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(256)))
	_, err := conn.Write(chunk.Buffer())
	assert.Nil(t, err)
	chunkReceived := <-chunks
	assert.NotNil(t, chunkReceived)
	assert.Equal(t, chunkReceived.Header.messageHeader.timestamp, chunk.Header.messageHeader.timestamp)
	assert.Equal(t, chunkReceived.Header.messageHeader.messageLength, chunk.Header.messageHeader.messageLength)
	assert.Equal(t, chunkReceived.Header.messageHeader.messageTypeId, chunk.Header.messageHeader.messageTypeId)
	assert.Equal(t, chunkReceived.Header.messageHeader.messageStreamId, chunk.Header.messageHeader.messageStreamId)
}

func TestType1Chunk(t *testing.T) {
	address, chunks := AcceptTestChunk()
	conn, _ := net.Dial("tcp", address)
	basicHeader := NewBasicHeader(uint8(1), uint32(2))
	messageHeader := NewMessageHeader(uint32(12), uint32(32), uint8(1), uint32(123456))
	header := NewHeader(*basicHeader, *messageHeader, uint32(0))
	chunk := NewChunk(*header, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(256)))
	_, err := conn.Write(chunk.Buffer())
	assert.Nil(t, err)
	chunkReceived := <-chunks
	assert.NotNil(t, chunkReceived)
	assert.Equal(t, chunkReceived.Header.messageHeader.timestamp, chunk.Header.messageHeader.timestamp)
	assert.Equal(t, chunkReceived.Header.messageHeader.messageLength, chunk.Header.messageHeader.messageLength)
	assert.Equal(t, chunkReceived.Header.messageHeader.messageTypeId, chunk.Header.messageHeader.messageTypeId)
	assert.NotEqual(t, chunkReceived.Header.messageHeader.messageStreamId, chunk.Header.messageHeader.messageStreamId)
}

func TestType2Chunk(t *testing.T) {
	address, chunks := AcceptTestChunk()
	conn, _ := net.Dial("tcp", address)
	basicHeader := NewBasicHeader(uint8(2), uint32(2))
	messageHeader := NewMessageHeader(uint32(12), uint32(32), uint8(1), uint32(123456))
	header := NewHeader(*basicHeader, *messageHeader, uint32(0))
	chunk := NewChunk(*header, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(256)))
	_, err := conn.Write(chunk.Buffer())
	assert.Nil(t, err)
	chunkReceived := <-chunks
	assert.NotNil(t, chunkReceived)
	assert.Equal(t, chunkReceived.Header.messageHeader.timestamp, chunk.Header.messageHeader.timestamp)
	assert.NotEqual(t, chunkReceived.Header.messageHeader.messageLength, chunk.Header.messageHeader.messageLength)
	assert.NotEqual(t, chunkReceived.Header.messageHeader.messageTypeId, chunk.Header.messageHeader.messageTypeId)
	assert.NotEqual(t, chunkReceived.Header.messageHeader.messageStreamId, chunk.Header.messageHeader.messageStreamId)
}

func TestType3Chunk(t *testing.T) {
	address, chunks := AcceptTestChunk()
	conn, _ := net.Dial("tcp", address)
	basicHeader := NewBasicHeader(uint8(3), uint32(2))
	messageHeader := NewMessageHeader(uint32(12), uint32(32), uint8(1), uint32(123456))
	header := NewHeader(*basicHeader, *messageHeader, uint32(0))
	chunk := NewChunk(*header, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(256)))
	_, err := conn.Write(chunk.Buffer())
	assert.Nil(t, err)
	chunkReceived := <-chunks
	assert.NotNil(t, chunkReceived)
	assert.NotEqual(t, chunkReceived.Header.messageHeader.timestamp, chunk.Header.messageHeader.timestamp)
	assert.NotEqual(t, chunkReceived.Header.messageHeader.messageLength, chunk.Header.messageHeader.messageLength)
	assert.NotEqual(t, chunkReceived.Header.messageHeader.messageTypeId, chunk.Header.messageHeader.messageTypeId)
	assert.NotEqual(t, chunkReceived.Header.messageHeader.messageStreamId, chunk.Header.messageHeader.messageStreamId)
}

func TestChunkExtendedTimestamp(t *testing.T) {
	address, chunks := AcceptTestChunk()
	conn, _ := net.Dial("tcp", address)
	basicHeader := NewBasicHeader(uint8(0), uint32(2))
	messageHeader := NewMessageHeader(math.MaxUint32, uint32(32), uint8(1), uint32(0))
	header := NewHeader(*basicHeader, *messageHeader, math.MaxUint32)
	chunk := NewChunk(*header, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(256)))
	_, err := conn.Write(chunk.Buffer())
	assert.Nil(t, err)
	chunkReceived := <-chunks
	assert.NotNil(t, chunkReceived)
	assert.Equal(t, chunkReceived.Header.messageHeader.timestamp, uint32(0xffffff))
	assert.Equal(t, chunkReceived.Header.extendedTimestamp, chunk.Header.extendedTimestamp)
}
