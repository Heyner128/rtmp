package chunk

import (
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
		min(header.MessageHeader.MessageLength-uint32(len(conn.CurrentMessage.Data)), conn.MaxChunkSize),
	)
	_, err = conn.Read(data)
	if err != nil {
		return nil, err
	}
	chunk := NewChunk(*header, data)
	fmt.Printf("chunk received, Fmt: %d - message type id: %d - message stream id: %d\n", chunk.Header.BasicHeader.Fmt, chunk.Header.MessageHeader.MessageTypeId, chunk.Header.MessageHeader.MessageStreamId)
	if chunk.Header.BasicHeader.Fmt == 0 {
		if conn.CurrentMessage.MessageStreamId != chunk.Header.MessageHeader.MessageStreamId {
			conn.CurrentMessage.Data = chunk.Data
		} else {
			conn.CurrentMessage.Data = append(conn.CurrentMessage.Data, chunk.Data...)
		}
		conn.CurrentMessage.MessageLength = chunk.Header.MessageHeader.MessageLength
		conn.CurrentMessage.MessageTypeId = chunk.Header.MessageHeader.MessageTypeId
		conn.CurrentMessage.MessageStreamId = chunk.Header.MessageHeader.MessageStreamId
		if conn.CurrentMessage.MessageLength == uint32(len(conn.CurrentMessage.Data)) && conn.CurrentMessage.MessageLength > 0 {
			conn.Messages <- conn.CurrentMessage
		}
	}
	return chunk, nil
}
