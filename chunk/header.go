package chunk

import (
	"encoding/binary"
	"rtmp/conn"
)

type Header struct {
	BasicHeader       BasicHeader
	MessageHeader     MessageHeader
	ExtendedTimestamp uint32
}

func NewHeader(basicHeader BasicHeader, messageHeader MessageHeader, extendedTimestamp uint32) *Header {
	return &Header{
		BasicHeader:       basicHeader,
		MessageHeader:     messageHeader,
		ExtendedTimestamp: extendedTimestamp,
	}
}

func ReadChunkHeader(connection *conn.Conn) (*Header, error) {
	basicHeader, err := ReadBasicHeader(connection)
	if err != nil {
		return nil, err
	}
	messageHeader, err := ReadMessageHeader(connection, *basicHeader)
	if err != nil {
		return nil, err
	}
	var extendedTimestamp uint32
	extendedTimestampBuffer := make([]byte, 0)
	if messageHeader.Timestamp >= 16777215 {
		extendedTimestampBuffer = make([]byte, 4)
		_, err = connection.Read(extendedTimestampBuffer)
		if err != nil {
			return nil, err
		}
		extendedTimestamp = binary.BigEndian.Uint32(extendedTimestampBuffer)
	}
	return NewHeader(*basicHeader, *messageHeader, extendedTimestamp), nil
}

type BasicHeader struct {
	Fmt           uint8
	ChunkStreamId uint32
}

func NewBasicHeader(fmt uint8, chunkStreamId uint32) *BasicHeader {
	return &BasicHeader{
		Fmt:           fmt,
		ChunkStreamId: chunkStreamId,
	}
}

func ReadBasicHeader(connection *conn.Conn) (*BasicHeader, error) {
	firstByte := make([]byte, 1)
	chunkStreamIdBuffer := make([]byte, 0)
	_, err := connection.Read(firstByte)
	if err != nil {
		return nil, err
	}
	chunkStreamIdBuffer = []byte{firstByte[0] & 0x3F}
	fmt := (firstByte[0] & 0xC0) >> 6
	chunkStreamId := uint32(chunkStreamIdBuffer[0])
	if chunkStreamIdBuffer[0] == 0x00 {
		secondByte := make([]byte, 1)
		_, err := connection.Read(secondByte)
		if err != nil {
			return nil, err
		}
		chunkStreamIdBuffer = binary.BigEndian.AppendUint32(make([]byte, 0), uint32(secondByte[0])+64)[1:]
		chunkStreamId = binary.BigEndian.Uint32(append([]byte{0x00, 0x00}, chunkStreamIdBuffer...))
	} else if chunkStreamIdBuffer[0] == 0x3F {
		secondByte := make([]byte, 1)
		_, err := connection.Read(secondByte)
		if err != nil {
			return nil, err
		}
		thirdByte := make([]byte, 1)
		_, err = connection.Read(thirdByte)
		if err != nil {
			return nil, err
		}
		chunkStreamIdBuffer = binary.BigEndian.AppendUint32(make([]byte, 0), uint32(thirdByte[0])*256+uint32(secondByte[0])+64)[1:]
		chunkStreamId = binary.BigEndian.Uint32(append([]byte{0x00}, chunkStreamIdBuffer...))
	}
	return &BasicHeader{
		Fmt:           fmt,
		ChunkStreamId: chunkStreamId,
	}, nil
}

type MessageHeader struct {
	Timestamp       uint32
	MessageLength   uint32
	MessageTypeId   uint8
	MessageStreamId uint32
}

func NewMessageHeader(timestamp uint32, messageLength uint32, messageTypeId uint8, messageStreamId uint32) *MessageHeader {
	return &MessageHeader{
		Timestamp:       timestamp,
		MessageLength:   messageLength,
		MessageTypeId:   messageTypeId,
		MessageStreamId: messageStreamId,
	}
}

func ReadMessageHeader(conn *conn.Conn, basicHeader BasicHeader) (*MessageHeader, error) {
	timestampBuffer := make([]byte, 3)
	var timestamp uint32
	messageLengthBuffer := make([]byte, 0)
	var messageLength uint32
	messageTypeIdBuffer := make([]byte, 0)
	var messageTypeId uint8
	messageStreamIdBuffer := make([]byte, 0)
	var messageStreamId uint32
	if basicHeader.Fmt <= 2 {
		_, err := conn.Read(timestampBuffer)
		if err != nil {
			return nil, err
		}
		timestamp = binary.BigEndian.Uint32(append([]byte{0x00}, timestampBuffer...))
	}
	if basicHeader.Fmt <= 1 {
		messageLengthBuffer = make([]byte, 3)
		_, err := conn.Read(messageLengthBuffer)
		if err != nil {
			return nil, err
		}
		messageLength = binary.BigEndian.Uint32(append([]byte{0x00}, messageLengthBuffer...))
		messageTypeIdBuffer = make([]byte, 1)
		_, err = conn.Read(messageTypeIdBuffer)
		if err != nil {
			return nil, err
		}
		messageTypeId = messageTypeIdBuffer[0]
	}
	if basicHeader.Fmt == 0 {
		messageStreamIdBuffer = make([]byte, 4)
		_, err := conn.Read(messageStreamIdBuffer)
		if err != nil {
			return nil, err
		}
		messageStreamId = binary.LittleEndian.Uint32(messageStreamIdBuffer)
	}
	return &MessageHeader{
		Timestamp:       timestamp,
		MessageLength:   messageLength,
		MessageTypeId:   messageTypeId,
		MessageStreamId: messageStreamId,
	}, nil
}
