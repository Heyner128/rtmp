package chunk

import (
	"encoding/binary"
	"fmt"
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

func (chunk *Chunk) Encode() []byte {
	buffer := make([]byte, 0)
	basicHeader := chunk.encodeBasicHeader()
	messageHeader, extendedTimeStamp := chunk.encodeMessageHeader()
	buffer = append(buffer, basicHeader...)
	buffer = append(buffer, messageHeader...)
	buffer = append(buffer, extendedTimeStamp...)
	buffer = append(buffer, chunk.Data...)
	return buffer
}

func (chunk *Chunk) String() string {
	basicHeader := chunk.Header.BasicHeader
	messageHeader := chunk.Header.MessageHeader

	res := fmt.Sprintf("chunk fmt: %d - chunk stream id: %d", basicHeader.Fmt, basicHeader.ChunkStreamId)

	if basicHeader.Fmt <= 2 {
		res += fmt.Sprintf(" - message timestamp: %d", messageHeader.Timestamp)
	}
	if basicHeader.Fmt <= 1 {
		res += fmt.Sprintf(" - message length: %d - message type id: %d", messageHeader.MessageLength, messageHeader.MessageTypeId)
	}
	if basicHeader.Fmt == 0 {
		res += fmt.Sprintf(" - message stream id: %d", messageHeader.MessageStreamId)
	}
	res += fmt.Sprintf(" - chunk data size: %d", len(chunk.Data))
	return res
}

func (chunk *Chunk) encodeBasicHeader() []byte {
	basicHeader := make([]byte, 0)
	if chunk.Header.BasicHeader.ChunkStreamId >= 2 && chunk.Header.BasicHeader.ChunkStreamId <= 63 {
		basicHeader = append(basicHeader, chunk.Header.BasicHeader.Fmt<<6|uint8(chunk.Header.BasicHeader.ChunkStreamId))
	} else if chunk.Header.BasicHeader.ChunkStreamId >= 64 && chunk.Header.BasicHeader.ChunkStreamId <= 319 {
		basicHeader = append(basicHeader, chunk.Header.BasicHeader.Fmt<<6|0x00)
		basicHeader = append(basicHeader, uint8(chunk.Header.BasicHeader.ChunkStreamId-64))
	} else if chunk.Header.BasicHeader.ChunkStreamId >= 320 && chunk.Header.BasicHeader.ChunkStreamId <= 65599 {
		basicHeader = append(basicHeader, chunk.Header.BasicHeader.Fmt<<6|0x3F)
		basicHeader = binary.BigEndian.AppendUint16(basicHeader, uint16(chunk.Header.BasicHeader.ChunkStreamId-64))
	}
	return basicHeader
}

func (chunk *Chunk) encodeMessageHeader() ([]byte, []byte) {
	messageHeader := make([]byte, 0)
	extendedTimeStamp := make([]byte, 0)
	if chunk.Header.BasicHeader.Fmt <= 2 {
		if chunk.Header.MessageHeader.Timestamp >= 0xFFFFFF {
			messageHeader = append(messageHeader, []byte{0xFF, 0xFF, 0xFF}...)
			extendedTimeStamp = binary.BigEndian.AppendUint32(extendedTimeStamp, chunk.Header.ExtendedTimestamp)
		} else {
			messageHeader = append(messageHeader, binary.BigEndian.AppendUint32(make([]byte, 0), chunk.Header.MessageHeader.Timestamp)[1:]...)
		}
	}
	if chunk.Header.BasicHeader.Fmt <= 1 {
		messageHeader = append(messageHeader, binary.BigEndian.AppendUint32(make([]byte, 0), chunk.Header.MessageHeader.MessageLength)[1:]...)
		messageHeader = append(messageHeader, chunk.Header.MessageHeader.MessageTypeId)
	}
	if chunk.Header.BasicHeader.Fmt == 0 {
		messageHeader = append(messageHeader, binary.LittleEndian.AppendUint32(make([]byte, 0), chunk.Header.MessageHeader.MessageStreamId)...)
	}
	return messageHeader, extendedTimeStamp
}
