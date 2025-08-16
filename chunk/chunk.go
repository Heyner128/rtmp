package chunk

import (
	"encoding/binary"
)

type Chunk struct {
	Header Header
	Data   []byte
}

type Header struct {
	BasicHeader       []byte // 1 to 3 bytes
	MessageHeader     []byte // 0, 3, 7 or 11 bytes
	ExtendedTimestamp []byte // 0 to 4 bytes
}

func NewSetSizeChunk(timestampDelta uint32, size uint32) Chunk {
	chunk := new(Chunk)

	// Header
	chunk.Header.BasicHeader = []byte{0x02 << 6}
	var maxTimestampDelta = []byte{0x00, 0xFF, 0xFF, 0xFF}
	if timestampDelta > binary.BigEndian.Uint32(maxTimestampDelta) {
		chunk.Header.MessageHeader = append(chunk.Header.MessageHeader, maxTimestampDelta[1:]...)
		chunk.Header.ExtendedTimestamp = append(make([]byte, 0), binary.BigEndian.AppendUint32(make([]byte, 0), timestampDelta)...)
	} else {
		chunk.Header.MessageHeader = append(chunk.Header.MessageHeader, binary.BigEndian.AppendUint32(make([]byte, 0), timestampDelta)[1:]...)
	}
	chunk.Header.MessageHeader = append(chunk.Header.MessageHeader, binary.BigEndian.AppendUint32(make([]byte, 0), 32)[1:]...)
	chunk.Header.MessageHeader = append(chunk.Header.MessageHeader, 0x00)
	chunk.Header.MessageHeader = append(chunk.Header.MessageHeader, []byte{0x00, 0x00, 0x00, 0x00}...)
	// Data
	var maxChunkSize = []byte{0x00, 0xFF, 0xFF, 0xFF}
	if size == uint32(0) {
		chunk.Data = binary.BigEndian.AppendUint32(make([]byte, 0), 128)
	} else if size > binary.BigEndian.Uint32(maxChunkSize) {
		chunk.Data = append(chunk.Data, maxChunkSize...)
	} else {
		chunk.Data = append(chunk.Data, binary.BigEndian.AppendUint32(make([]byte, 0), size)...)
	}
	return *chunk
}
