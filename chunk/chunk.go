package chunk

import (
	"encoding/binary"
	"miistream/rtmpconn"
)

type Chunk struct {
	Header Header
	Data   []byte
}

type Header struct {
	basicHeader   BasicHeader   // 1 to 3 bytes
	messageHeader MessageHeader // 0, 3, 7 or 11 bytes
}

type BasicHeader struct {
	fmt           uint8
	chunkStreamId uint32
}

type MessageHeader struct {
	timestamp       uint32
	messageLength   uint32
	messageTypeId   uint8
	messageStreamId uint32
}

func Accept(conn *rtmpconn.RtmpConn) error {
	_, err := ReadChunk(conn)
	return err
}

func ReadChunk(conn *rtmpconn.RtmpConn) (*Chunk, error) {
	fmt, chunkStreamId, err := ReadBasicHeader(conn)
	if err != nil {
		return nil, err
	}
	timestampBuffer := make([]byte, 3)
	_, err = conn.Read(timestampBuffer)
	// Type 0 chunk
	if *fmt == uint8(0) {
		messageLengthBuffer := make([]byte, 3)
		_, err = conn.Read(messageLengthBuffer)
		if err != nil {
			return nil, err
		}
		messageTypeIdBuffer := make([]byte, 1)
		_, err = conn.Read(messageTypeIdBuffer)
		if err != nil {
			return nil, err
		}
		messageStreamIdBuffer := make([]byte, 4)
		_, err = conn.Read(messageStreamIdBuffer)
		if err != nil {
			return nil, err
		}
		extendedTimestampBuffer := make([]byte, 0)
		if binary.BigEndian.Uint32(append([]byte{0x00}, timestampBuffer...)) == 16777215 {
			extendedTimestampBuffer = make([]byte, 4)
			_, err = conn.Read(extendedTimestampBuffer)
			if err != nil {
				return nil, err
			}
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
		timestamp := uint32(0)
		if len(extendedTimestampBuffer) > 0 {
			timestamp = binary.BigEndian.Uint32(extendedTimestampBuffer)
		} else {
			timestamp = binary.BigEndian.Uint32(append([]byte{0x00}, timestampBuffer...))
		}
		return &Chunk{
			Header: Header{
				basicHeader: BasicHeader{
					fmt:           *fmt,
					chunkStreamId: *chunkStreamId,
				},
				messageHeader: MessageHeader{
					timestamp:       timestamp,
					messageLength:   binary.BigEndian.Uint32(append([]byte{0x00}, messageLengthBuffer...)),
					messageTypeId:   messageTypeIdBuffer[0],
					messageStreamId: binary.BigEndian.Uint32(messageStreamIdBuffer),
				},
			},
			Data: data,
		}, nil
	}
	return &Chunk{}, nil
}

func ReadBasicHeader(conn *rtmpconn.RtmpConn) (*uint8, *uint32, error) {
	firstByte := make([]byte, 1)
	chunkStreamIdBuffer := make([]byte, 0)
	_, err := conn.Read(firstByte)
	if err != nil {
		return nil, nil, err
	}
	chunkStreamIdBuffer = []byte{firstByte[0] & 0x3F}
	fmt := (firstByte[0] & 0xC0) >> 6
	chunkStreamId := uint32(chunkStreamIdBuffer[0])
	if chunkStreamIdBuffer[0] == 0x00 {
		secondByte := make([]byte, 1)
		_, err := conn.Read(secondByte)
		if err != nil {
			return &fmt, nil, err
		}
		chunkStreamIdBuffer = binary.BigEndian.AppendUint32(make([]byte, 0), uint32(secondByte[0])+64)[1:]
		chunkStreamId = binary.BigEndian.Uint32(append([]byte{0x00, 0x00}, chunkStreamIdBuffer...))
	} else if chunkStreamIdBuffer[0] == 0x01 {
		secondByte := make([]byte, 1)
		_, err := conn.Read(secondByte)
		if err != nil {
			return &fmt, nil, err
		}
		thirdByte := make([]byte, 1)
		_, err = conn.Read(thirdByte)
		if err != nil {
			return &fmt, nil, err
		}
		chunkStreamIdBuffer = binary.BigEndian.AppendUint32(make([]byte, 0), uint32(thirdByte[0])*256+uint32(secondByte[0])+64)[1:]
		chunkStreamId = binary.BigEndian.Uint32(append([]byte{0x00}, chunkStreamIdBuffer...))
	}
	return &fmt, &chunkStreamId, nil
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
		extendedTimeStamp = binary.BigEndian.AppendUint32(extendedTimeStamp, chunk.Header.messageHeader.timestamp)
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
