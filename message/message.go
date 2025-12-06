package message

import (
	"encoding/binary"
	"log"
	"math/rand"
	"rtmp/amf"
	"rtmp/chunk"
	"rtmp/rtmpconn"
)

type Message struct {
	MessageTypeId   uint8
	MessageStreamId uint32
	ChunkStreamId   uint32
	Data            []byte
}

func NewMessage(messageTypeId uint8, data []byte) *Message {
	var chunkStreamId uint32
	if isProtocolControlMessage(messageTypeId) {
		chunkStreamId = 2
	} else {
		chunkStreamId = generateChunkStreamId()
	}
	return &Message{
		ChunkStreamId:   chunkStreamId,
		MessageTypeId:   messageTypeId,
		MessageStreamId: rand.Uint32(),
		Data:            data,
	}
}

func generateChunkStreamId() uint32 {
	minId := 3
	maxId := 65599
	return uint32(rand.Intn(maxId-minId) + minId)
}

func newWindowAcknowledgementSizeMessage(acknowledgementSize uint32) *Message {
	return NewMessage(uint8(5), binary.BigEndian.AppendUint32(make([]byte, 0), acknowledgementSize))
}

func (message *Message) Send(conn rtmpconn.RtmpConn) error {
	for _, nChunk := range message.buildChunks(int(conn.MaxChunkSize)) {
		_, err := conn.Write(nChunk.Encode())
		if err != nil {
			return err
		}
	}
	return nil
}

func (message *Message) buildChunks(chunkSize int) []chunk.Chunk {
	numberOfChunks := len(message.Data) / chunkSize
	chunks := make([]chunk.Chunk, 0)
	for i := range numberOfChunks + 1 {
		var basicHeader chunk.BasicHeader
		var messageHeader chunk.MessageHeader
		if i == 0 {
			basicHeader = *chunk.NewBasicHeader(uint8(0), message.ChunkStreamId)
			messageHeader = *chunk.NewMessageHeader(uint32(0), uint32(len(message.Data)), message.MessageTypeId, message.MessageStreamId)
		} else {
			basicHeader = *chunk.NewBasicHeader(uint8(3), message.ChunkStreamId)
		}
		header := chunk.NewHeader(basicHeader, messageHeader, uint32(0))
		chunks = append(
			chunks,
			*chunk.NewChunk(*header, message.Data[i*chunkSize:min((i+1)*chunkSize, len(message.Data))]),
		)
	}
	return chunks
}

func isProtocolControlMessage(messageTypeId uint8) bool {
	return messageTypeId == 0x01 || messageTypeId == 0x02 || messageTypeId == 0x03 || messageTypeId == 0x05 || messageTypeId == 0x06
}

// TODO refactor me pls
func Accept(conn *rtmpconn.RtmpConn) (*chunk.Chunk, error) {
	header, err := chunk.ReadChunkHeader(conn)
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
	receivedChunk := chunk.NewChunk(*header, data)
	log.Printf("receivedChunk received, Fmt: %d - receivedChunk stream id: %d - message type id: %d - message stream id: %d", receivedChunk.Header.BasicHeader.Fmt, receivedChunk.Header.BasicHeader.ChunkStreamId, receivedChunk.Header.MessageHeader.MessageTypeId, receivedChunk.Header.MessageHeader.MessageStreamId)
	log.Printf(" - message length: %d\n", receivedChunk.Header.MessageHeader.MessageLength)
	if receivedChunk.Header.BasicHeader.Fmt <= 1 {
		conn.CurrentMessage.Length = receivedChunk.Header.MessageHeader.MessageLength
		conn.CurrentMessage.TypeId = receivedChunk.Header.MessageHeader.MessageTypeId
	}
	if receivedChunk.Header.BasicHeader.Fmt == 0 {
		conn.CurrentMessage.StreamId = receivedChunk.Header.MessageHeader.MessageStreamId
	}
	conn.CurrentMessage.Data = append(conn.CurrentMessage.Data, receivedChunk.Data...)
	if conn.CurrentMessage.Length == conn.CurrentMessage.DataSize() && conn.CurrentMessage.Length > 0 {
		if conn.CurrentMessage.TypeId == 1 {
			conn.MaxChunkSize = binary.BigEndian.Uint32(conn.CurrentMessage.Data[0:4]) >> 1
		} else if conn.CurrentMessage.TypeId == 2 {
			conn.CurrentMessage = nil
		} else if conn.CurrentMessage.TypeId == 5 {
			conn.WindowAcknowledgementSize = binary.BigEndian.Uint32(conn.CurrentMessage.Data[0:4])
		} else if conn.CurrentMessage.TypeId == 20 {
			command, err := amf.DecodeAmfCommand(conn.CurrentMessage.Data)
			if err != nil {
				return nil, err
			}
			log.Printf("Command received: %s\n", command)
			sendWindowAcknowledgementSize := newWindowAcknowledgementSizeMessage(conn.SendWindowAcknowledgementSize)
			err = sendWindowAcknowledgementSize.Send(*conn)
			if err != nil {
				return nil, err
			}

		}
		conn.Messages <- conn.CurrentMessage
		conn.CurrentMessage = new(rtmpconn.Message)
	}
	return receivedChunk, nil
}
