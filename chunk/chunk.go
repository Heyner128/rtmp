package chunk

import (
	"encoding/binary"
	"fmt"
	"rtmp/rtmpconn"
)

type Chunk struct {
	Header Header
	Data   []byte
}

func NewChunk(header Header, data []byte) *Chunk {
	return &Chunk{
		Header: header,
		Data:   data,
	}
}

func Accept(conn *rtmpconn.RtmpConn) (*Chunk, error) {
	header, err := ReadHeader(conn)
	if err != nil {
		return nil, err
	}
	data := make(
		[]byte,
		min(max(header.MessageHeader.MessageLength, conn.CurrentMessage.Length)-conn.CurrentMessage.DataSize(), conn.MaxChunkSize),
	)
	_, err = conn.Read(data)
	if err != nil {
		return nil, err
	}
	chunk := NewChunk(*header, data)
	fmt.Printf("chunk received, Fmt: %d - chunk stream id: %d - message type id: %d - message stream id: %d", chunk.Header.BasicHeader.Fmt, chunk.Header.BasicHeader.ChunkStreamId, chunk.Header.MessageHeader.MessageTypeId, chunk.Header.MessageHeader.MessageStreamId)
	fmt.Printf(" - message length: %d\n", chunk.Header.MessageHeader.MessageLength)
	if chunk.Header.BasicHeader.Fmt <= 1 {
		conn.CurrentMessage.Length = chunk.Header.MessageHeader.MessageLength
		conn.CurrentMessage.TypeId = chunk.Header.MessageHeader.MessageTypeId
	}
	if chunk.Header.BasicHeader.Fmt == 0 {
		conn.CurrentMessage.StreamId = chunk.Header.MessageHeader.MessageStreamId
	}
	conn.CurrentMessage.Data = append(conn.CurrentMessage.Data, chunk.Data...)
	if conn.CurrentMessage.Length == conn.CurrentMessage.DataSize() && conn.CurrentMessage.Length > 0 {
		if conn.CurrentMessage.TypeId == 1 {
			conn.MaxChunkSize = binary.BigEndian.Uint32(conn.CurrentMessage.Data[0:4]) >> 1
		} else if conn.CurrentMessage.TypeId == 2 {
			conn.CurrentMessage = nil
		} else if conn.CurrentMessage.TypeId == 5 {
			conn.WindowAcknowledgementSize = binary.BigEndian.Uint32(conn.CurrentMessage.Data[0:4])
		}
		conn.Messages <- conn.CurrentMessage
		conn.CurrentMessage = new(rtmpconn.Message)
	}
	return chunk, nil
}
