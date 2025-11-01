package chunk

import (
	"encoding/binary"
	"net"
	"rtmp/rtmpconn"
	"testing"
	"time"
)

func acceptChunk(t *testing.T) (string, chan Chunk) {
	t.Helper()
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
			chunk, err := Accept(rtmpConn)
			if err != nil || chunk == nil {
				continue
			}
			chunks <- *chunk
		}
	}()
	return listener.Addr().String(), chunks
}

func (chunk *Chunk) Encode(t *testing.T) []byte {
	t.Helper()
	buffer := make([]byte, 0)
	//basic header
	BasicHeader := make([]byte, 0)
	if chunk.Header.BasicHeader.ChunkStreamId >= 2 && chunk.Header.BasicHeader.ChunkStreamId <= 63 {
		BasicHeader = append(BasicHeader, chunk.Header.BasicHeader.Fmt<<6|uint8(chunk.Header.BasicHeader.ChunkStreamId))
	} else if chunk.Header.BasicHeader.ChunkStreamId >= 64 && chunk.Header.BasicHeader.ChunkStreamId <= 319 {
		BasicHeader = append(BasicHeader, uint8(chunk.Header.BasicHeader.ChunkStreamId-64))
		BasicHeader = append(BasicHeader, chunk.Header.BasicHeader.Fmt<<6)
	} else if chunk.Header.BasicHeader.ChunkStreamId >= 320 && chunk.Header.BasicHeader.ChunkStreamId <= 65599 {
		BasicHeader = binary.BigEndian.AppendUint16(BasicHeader, uint16(chunk.Header.BasicHeader.ChunkStreamId-64))
		BasicHeader = append(BasicHeader, chunk.Header.BasicHeader.Fmt<<6|0x3F)
	}
	//timestamp
	MessageHeader := make([]byte, 0)
	extendedTimeStamp := make([]byte, 0)
	if chunk.Header.BasicHeader.Fmt <= 2 {
		// timestamp
		if chunk.Header.MessageHeader.Timestamp >= 16777215 {
			MessageHeader = append(MessageHeader, []byte{0xFF, 0xFF, 0xFF}...)
			extendedTimeStamp = binary.BigEndian.AppendUint32(extendedTimeStamp, chunk.Header.ExtendedTimestamp)
		} else {
			MessageHeader = append(MessageHeader, binary.BigEndian.AppendUint32(make([]byte, 0), chunk.Header.MessageHeader.Timestamp)[1:]...)
		}
	}
	if chunk.Header.BasicHeader.Fmt <= 1 {
		//Message length
		MessageHeader = append(MessageHeader, binary.BigEndian.AppendUint32(make([]byte, 0), chunk.Header.MessageHeader.MessageLength)[1:]...)
		//Message type
		MessageHeader = append(MessageHeader, chunk.Header.MessageHeader.MessageTypeId)
	}
	if chunk.Header.BasicHeader.Fmt == 0 {
		//Message stream id
		MessageHeader = append(MessageHeader, binary.LittleEndian.AppendUint32(make([]byte, 0), chunk.Header.MessageHeader.MessageStreamId)...)
	}
	buffer = append(buffer, BasicHeader...)
	buffer = append(buffer, MessageHeader...)
	buffer = append(buffer, extendedTimeStamp...)
	buffer = append(buffer, chunk.Data...)
	return buffer
}
