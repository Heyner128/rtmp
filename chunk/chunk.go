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

func Accept(conn *rtmpconn.RtmpConn) error {
	chunk, err := ReadChunk(conn)
	if err == nil {
		fmt.Printf("chunk received, fmt: %d - message type id: %d\n", chunk.Header.basicHeader.fmt, chunk.Header.messageHeader.messageTypeId)
	}
	return err
}

func ReadChunk(conn *rtmpconn.RtmpConn) (*Chunk, error) {
	header, err := ReadHeader(conn)
	if err != nil {
		return nil, err
	}
	// here is not max it depends on how much data is left to finish the current message
	data := make(
		[]byte,
		conn.MaxChunkSize,
	)
	_, err = conn.Read(data)
	if err != nil {
		return nil, err
	}
	return &Chunk{
		Header: *header,
		Data:   data,
	}, nil
}

func (chunk Chunk) Buffer() []byte {
	buffer := make([]byte, 0)
	//basic header
	basicHeader := make([]byte, 0)
	if chunk.Header.basicHeader.chunkStreamId >= 2 && chunk.Header.basicHeader.chunkStreamId <= 63 {
		basicHeader = append(basicHeader, chunk.Header.basicHeader.fmt<<6|uint8(chunk.Header.basicHeader.chunkStreamId))
	} else if chunk.Header.basicHeader.chunkStreamId >= 64 && chunk.Header.basicHeader.chunkStreamId <= 319 {
		basicHeader = append(basicHeader, uint8(chunk.Header.basicHeader.chunkStreamId-64))
		basicHeader = append(basicHeader, chunk.Header.basicHeader.fmt<<6)
	} else if chunk.Header.basicHeader.chunkStreamId >= 320 && chunk.Header.basicHeader.chunkStreamId <= 65599 {
		basicHeader = binary.BigEndian.AppendUint16(basicHeader, uint16(chunk.Header.basicHeader.chunkStreamId-64))
		basicHeader = append(basicHeader, chunk.Header.basicHeader.fmt<<6|0x3F)
	}
	//timestamp
	messageHeader := make([]byte, 0)
	extendedTimeStamp := make([]byte, 0)
	if chunk.Header.messageHeader.timestamp >= 16777215 {
		messageHeader = append(messageHeader, []byte{0xFF, 0xFF, 0xFF}...)
		extendedTimeStamp = binary.BigEndian.AppendUint32(extendedTimeStamp, chunk.Header.extendedTimestamp)
	} else {
		messageHeader = append(messageHeader, binary.BigEndian.AppendUint32(make([]byte, 0), chunk.Header.messageHeader.timestamp)[1:]...)
	}
	//Message length
	messageHeader = append(messageHeader, binary.BigEndian.AppendUint32(make([]byte, 0), chunk.Header.messageHeader.messageLength)[1:]...)
	//Message type
	messageHeader = append(messageHeader, chunk.Header.messageHeader.messageTypeId)
	//Message stream id
	messageHeader = append(messageHeader, binary.LittleEndian.AppendUint32(make([]byte, 0), chunk.Header.messageHeader.messageStreamId)...)
	// TODO append according type id spec
	buffer = append(buffer, basicHeader...)
	buffer = append(buffer, messageHeader...)
	buffer = append(buffer, extendedTimeStamp...)
	buffer = append(buffer, chunk.Data...)
	return buffer
}
