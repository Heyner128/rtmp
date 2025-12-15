package chunk_test

import (
	"encoding/binary"
	"math"
	"net"
	"rtmp/chunk"
	"rtmp/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType0Chunk(t *testing.T) {
	address, chunks := testutil.AcceptTestChunk(t)
	conn, _ := net.Dial("tcp", address)
	basicHeader := chunk.NewBasicHeader(uint8(0), uint32(2))
	messageHeader := chunk.NewMessageHeader(uint32(0), uint32(4), uint8(1), uint32(123456))
	header := chunk.NewHeader(*basicHeader, *messageHeader, uint32(0))
	testChunk := chunk.NewChunk(*header, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(256)))
	_, err := conn.Write(testChunk.Encode())
	assert.Nil(t, err)
	chunkReceived := <-chunks
	assert.NotNil(t, chunkReceived)
	assert.Equal(t, chunkReceived.Header.MessageHeader.Timestamp, testChunk.Header.MessageHeader.Timestamp)
	assert.Equal(t, chunkReceived.Header.MessageHeader.MessageLength, testChunk.Header.MessageHeader.MessageLength)
	assert.Equal(t, chunkReceived.Header.MessageHeader.MessageTypeId, testChunk.Header.MessageHeader.MessageTypeId)
	assert.Equal(t, chunkReceived.Header.MessageHeader.MessageStreamId, testChunk.Header.MessageHeader.MessageStreamId)
}

func TestType1Chunk(t *testing.T) {
	address, chunks := testutil.AcceptTestChunk(t)
	conn, _ := net.Dial("tcp", address)
	basicHeader := chunk.NewBasicHeader(uint8(1), uint32(2))
	messageHeader := chunk.NewMessageHeader(uint32(12), uint32(32), uint8(1), uint32(123456))
	header := chunk.NewHeader(*basicHeader, *messageHeader, uint32(0))
	testChunk := chunk.NewChunk(*header, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(256)))
	_, err := conn.Write(testChunk.Encode())
	assert.Nil(t, err)
	chunkReceived := <-chunks
	assert.NotNil(t, chunkReceived)
	assert.Equal(t, chunkReceived.Header.MessageHeader.Timestamp, testChunk.Header.MessageHeader.Timestamp)
	assert.Equal(t, chunkReceived.Header.MessageHeader.MessageLength, testChunk.Header.MessageHeader.MessageLength)
	assert.Equal(t, chunkReceived.Header.MessageHeader.MessageTypeId, testChunk.Header.MessageHeader.MessageTypeId)
	assert.NotEqual(t, chunkReceived.Header.MessageHeader.MessageStreamId, testChunk.Header.MessageHeader.MessageStreamId)
}

func TestType2Chunk(t *testing.T) {
	address, chunks := testutil.AcceptTestChunk(t)
	conn, _ := net.Dial("tcp", address)
	basicHeader := chunk.NewBasicHeader(uint8(2), uint32(2))
	messageHeader := chunk.NewMessageHeader(uint32(12), uint32(32), uint8(1), uint32(123456))
	header := chunk.NewHeader(*basicHeader, *messageHeader, uint32(0))
	testChunk := chunk.NewChunk(*header, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(256)))
	_, err := conn.Write(testChunk.Encode())
	assert.Nil(t, err)
	chunkReceived := <-chunks
	assert.NotNil(t, chunkReceived)
	assert.Equal(t, chunkReceived.Header.MessageHeader.Timestamp, testChunk.Header.MessageHeader.Timestamp)
	assert.NotEqual(t, chunkReceived.Header.MessageHeader.MessageLength, testChunk.Header.MessageHeader.MessageLength)
	assert.NotEqual(t, chunkReceived.Header.MessageHeader.MessageTypeId, testChunk.Header.MessageHeader.MessageTypeId)
	assert.NotEqual(t, chunkReceived.Header.MessageHeader.MessageStreamId, testChunk.Header.MessageHeader.MessageStreamId)
}

func TestType3Chunk(t *testing.T) {
	address, chunks := testutil.AcceptTestChunk(t)
	conn, _ := net.Dial("tcp", address)
	basicHeader := chunk.NewBasicHeader(uint8(3), uint32(2))
	messageHeader := chunk.NewMessageHeader(uint32(12), uint32(32), uint8(1), uint32(123456))
	header := chunk.NewHeader(*basicHeader, *messageHeader, uint32(0))
	testChunk := chunk.NewChunk(*header, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(256)))
	_, err := conn.Write(testChunk.Encode())
	assert.Nil(t, err)
	chunkReceived := <-chunks
	assert.NotNil(t, chunkReceived)
	assert.NotEqual(t, chunkReceived.Header.MessageHeader.Timestamp, testChunk.Header.MessageHeader.Timestamp)
	assert.NotEqual(t, chunkReceived.Header.MessageHeader.MessageLength, testChunk.Header.MessageHeader.MessageLength)
	assert.NotEqual(t, chunkReceived.Header.MessageHeader.MessageTypeId, testChunk.Header.MessageHeader.MessageTypeId)
	assert.NotEqual(t, chunkReceived.Header.MessageHeader.MessageStreamId, testChunk.Header.MessageHeader.MessageStreamId)
}

func TestChunkExtendedTimestamp(t *testing.T) {
	address, chunks := testutil.AcceptTestChunk(t)
	conn, _ := net.Dial("tcp", address)
	basicHeader := chunk.NewBasicHeader(uint8(0), uint32(2))
	messageHeader := chunk.NewMessageHeader(math.MaxUint32, uint32(32), uint8(1), uint32(0))
	header := chunk.NewHeader(*basicHeader, *messageHeader, math.MaxUint32)
	testChunk := chunk.NewChunk(*header, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(256)))
	_, err := conn.Write(testChunk.Encode())
	assert.Nil(t, err)
	chunkReceived := <-chunks
	assert.NotNil(t, chunkReceived)
	assert.Equal(t, chunkReceived.Header.MessageHeader.Timestamp, uint32(0xffffff))
	assert.Equal(t, chunkReceived.Header.ExtendedTimestamp, testChunk.Header.ExtendedTimestamp)
}
