package message

import (
	"math/rand"
	"rtmp/chunk"
	"rtmp/rtmpconn"
	"testing"
)

type Message struct {
	MessageTypeId   uint8
	MessageStreamId uint32
	Data            []byte
}

func newMessage(messageTypeId uint8, messageStreamId uint32, data []byte) *Message {
	return &Message{
		MessageTypeId:   messageTypeId,
		MessageStreamId: messageStreamId,
		Data:            data,
	}
}

func (message *Message) Send(t *testing.T, conn rtmpconn.RtmpConn) error {
	t.Helper()
	for _, nChunk := range message.Chunks(t, int(conn.MaxChunkSize)) {
		_, err := conn.Write(nChunk.Encode(t))
		if err != nil {
			return err
		}
	}
	return nil
}

func (message *Message) Chunks(t *testing.T, chunkSize int) []chunk.Chunk {
	t.Helper()
	numberOfChunks := len(message.Data) / chunkSize
	chunks := make([]chunk.Chunk, 0)
	for i := range numberOfChunks + 1 {
		var basicHeader chunk.BasicHeader
		var messageHeader chunk.MessageHeader
		if i == 0 {
			basicHeader = *chunk.NewBasicHeader(uint8(0), uint32(2))
			messageHeader = *chunk.NewMessageHeader(uint32(0), uint32(len(message.Data)), message.MessageTypeId, message.MessageStreamId)
		} else {
			basicHeader = *chunk.NewBasicHeader(uint8(3), uint32(2))
		}
		header := chunk.NewHeader(basicHeader, messageHeader, uint32(0))
		chunks = append(
			chunks,
			*chunk.NewChunk(*header, message.Data[i*chunkSize:min((i+1)*chunkSize, len(message.Data))]),
		)
	}
	return chunks
}

func generateRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(rand.Intn(255))
	}
	return bytes
}
