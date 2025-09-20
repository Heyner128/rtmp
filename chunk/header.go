package chunk

import (
	"encoding/binary"
	"rtmp/rtmpconn"
)

type Header struct {
	basicHeader       BasicHeader   // 1 to 3 bytes
	messageHeader     MessageHeader // 0, 3, 7 or 11 bytes
	extendedTimestamp uint32
}

func NewHeader(basicHeader BasicHeader, messageHeader MessageHeader, extendedTimestamp uint32) *Header {
	return &Header{
		basicHeader:       basicHeader,
		messageHeader:     messageHeader,
		extendedTimestamp: extendedTimestamp,
	}
}

func ReadHeader(conn *rtmpconn.RtmpConn) (*Header, error) {
	basicHeader, err := ReadBasicHeader(conn)
	if err != nil {
		return nil, err
	}
	messageHeader, err2 := ReadMessageHeader(conn, *basicHeader)
	if err2 != nil {
		return nil, err2
	}
	var extendedTimestamp uint32
	extendedTimestampBuffer := make([]byte, 0)
	if messageHeader.timestamp >= 16777215 {
		extendedTimestampBuffer = make([]byte, 4)
		_, err = conn.Read(extendedTimestampBuffer)
		if err != nil {
			return nil, err
		}
		extendedTimestamp = binary.BigEndian.Uint32(extendedTimestampBuffer)
	}
	return NewHeader(*basicHeader, *messageHeader, extendedTimestamp), nil
}

type BasicHeader struct {
	fmt           uint8
	chunkStreamId uint32
}

func NewBasicHeader(fmt uint8, chunkStreamId uint32) *BasicHeader {
	return &BasicHeader{
		fmt:           fmt,
		chunkStreamId: chunkStreamId,
	}
}

func ReadBasicHeader(conn *rtmpconn.RtmpConn) (*BasicHeader, error) {
	firstByte := make([]byte, 1)
	chunkStreamIdBuffer := make([]byte, 0)
	_, err := conn.Read(firstByte)
	if err != nil {
		return nil, err
	}
	chunkStreamIdBuffer = []byte{firstByte[0] & 0x3F}
	fmt := (firstByte[0] & 0xC0) >> 6
	chunkStreamId := uint32(chunkStreamIdBuffer[0])
	if chunkStreamIdBuffer[0] == 0x00 {
		secondByte := make([]byte, 1)
		_, err := conn.Read(secondByte)
		if err != nil {
			return nil, err
		}
		chunkStreamIdBuffer = binary.BigEndian.AppendUint32(make([]byte, 0), uint32(secondByte[0])+64)[1:]
		chunkStreamId = binary.BigEndian.Uint32(append([]byte{0x00, 0x00}, chunkStreamIdBuffer...))
	} else if chunkStreamIdBuffer[0] == 0x01 {
		secondByte := make([]byte, 1)
		_, err := conn.Read(secondByte)
		if err != nil {
			return nil, err
		}
		thirdByte := make([]byte, 1)
		_, err = conn.Read(thirdByte)
		if err != nil {
			return nil, err
		}
		chunkStreamIdBuffer = binary.BigEndian.AppendUint32(make([]byte, 0), uint32(thirdByte[0])*256+uint32(secondByte[0])+64)[1:]
		chunkStreamId = binary.BigEndian.Uint32(append([]byte{0x00}, chunkStreamIdBuffer...))
	}
	return &BasicHeader{
		fmt:           fmt,
		chunkStreamId: chunkStreamId,
	}, nil
}

type MessageHeader struct {
	timestamp       uint32
	messageLength   uint32
	messageTypeId   uint8
	messageStreamId uint32
}

func NewMessageHeader(timestamp uint32, messageLength uint32, messageTypeId uint8, messageStreamId uint32) *MessageHeader {
	return &MessageHeader{
		timestamp:       timestamp,
		messageLength:   messageLength,
		messageTypeId:   messageTypeId,
		messageStreamId: messageStreamId,
	}
}

func ReadMessageHeader(conn *rtmpconn.RtmpConn, basicHeader BasicHeader) (*MessageHeader, error) {
	timestampBuffer := make([]byte, 3)
	var timestamp uint32
	messageLengthBuffer := make([]byte, 0)
	var messageLength uint32
	messageTypeIdBuffer := make([]byte, 0)
	var messageTypeId uint8
	messageStreamIdBuffer := make([]byte, 0)
	var messageStreamId uint32
	_, err := conn.Read(timestampBuffer)
	if err != nil {
		return nil, err
	}
	if basicHeader.fmt <= 2 {
		timestamp = binary.BigEndian.Uint32(append([]byte{0x00}, timestampBuffer...))
	}
	if basicHeader.fmt <= 1 {
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
	if basicHeader.fmt == 0 {
		messageStreamIdBuffer = make([]byte, 4)
		_, err := conn.Read(messageStreamIdBuffer)
		if err != nil {
			return nil, err
		}
		messageStreamId = binary.LittleEndian.Uint32(messageStreamIdBuffer)
	}
	return &MessageHeader{
		timestamp:       timestamp,
		messageLength:   messageLength,
		messageTypeId:   messageTypeId,
		messageStreamId: messageStreamId,
	}, nil
}
