package message

import (
	"rtmp/chunk"
	"rtmp/rtmpconn"
	"testing"
)

func SendMessage(t *testing.T, message Message, conn rtmpconn.RtmpConn) error {
	t.Helper()
	for _, nChunk := range GetMessageChunks(t, message, int(conn.MaxChunkSize)) {
		_, err := conn.Write(chunk.GetChunkBuffer(t, nChunk))
		if err != nil {
			return err
		}
	}
	return nil
}

func GetMessageChunks(t *testing.T, message Message, chunkSize int) []chunk.Chunk {
	t.Helper()
	numberOfChunks := len(message.Data) / chunkSize
	chunks := make([]chunk.Chunk, 0)
	for i := range numberOfChunks {
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
		bytes[i] = byte(i)
	}
	return bytes
}
