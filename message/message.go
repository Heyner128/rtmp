package message

import (
	"encoding/binary"
	"log"
	"math/rand"
	"rtmp/amf"
	"rtmp/chunk"
	"rtmp/conn"
)

const (
	TypeSetChunkSize              = uint8(1)
	TypeAbortMessage              = uint8(2)
	TypeAcknowledgement           = uint8(3)
	TypeWindowAcknowledgementSize = uint8(5)
	TypeSetPeerBandwidth          = uint8(6)
	TypeVideo                     = uint8(9)
	TypeCommandMessageAmf0        = uint8(20)
)

const (
	SetPeerBandwidthLimitTypeHard = uint8(0)
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
	return NewMessage(TypeWindowAcknowledgementSize, binary.BigEndian.AppendUint32(make([]byte, 0), acknowledgementSize))
}

func newSetPeerBandwidthMessage(size uint32, limitType uint8) *Message {
	content := binary.BigEndian.AppendUint32(make([]byte, 0), size)
	content = append(content, limitType)
	return NewMessage(TypeSetPeerBandwidth, content)
}

func newAcknowledgementMessage(acknowledgementSize uint32) *Message {
	return NewMessage(TypeAcknowledgement, binary.BigEndian.AppendUint32(make([]byte, 0), acknowledgementSize))

}

func (message *Message) Send(conn conn.Conn) (int, error) {
	bytesSent := 0
	for _, nChunk := range message.BuildChunks(int(conn.MaxChunkSize)) {
		n, err := conn.Write(nChunk.Encode())
		bytesSent += n
		if err != nil {
			return 0, err
		}
	}
	return bytesSent, nil
}

func (message *Message) BuildChunks(chunkSize int) []chunk.Chunk {
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

func Accept(connection *conn.Conn) (*chunk.Chunk, error) {
	header, err := chunk.ReadChunkHeader(connection)
	if err != nil {
		return nil, err
	}
	data := make(
		[]byte,
		getNextChunkDataSize(connection, header),
	)
	_, err = connection.Read(data)
	if err != nil {
		return nil, err
	}
	receivedChunk := chunk.NewChunk(*header, data)
	connection.UnacknowledgedBytes += uint32(len(receivedChunk.Encode()))
	if connection.WindowAcknowledgementSize > 0 && connection.UnacknowledgedBytes >= connection.WindowAcknowledgementSize {
		acknowledgementMessage := newAcknowledgementMessage(connection.UnacknowledgedBytes)
		_, err = acknowledgementMessage.Send(*connection)
		connection.UnacknowledgedBytes = 0
		if err != nil {
			return nil, err
		}
	}
	handleReceivedChunk(connection, receivedChunk)
	connection.CurrentMessage.Data = append(connection.CurrentMessage.Data, receivedChunk.Data...)
	if connection.CurrentMessage.Length == connection.CurrentMessage.DataSize() && connection.CurrentMessage.Length > 0 {
		err = handleCompletedMessage(connection, connection.CurrentMessage)
		if err != nil {
			return nil, err
		}
	}
	return receivedChunk, nil
}

func getNextChunkDataSize(connection *conn.Conn, header *chunk.Header) uint32 {
	return min(max(header.MessageHeader.MessageLength, connection.CurrentMessage.Length)-connection.CurrentMessage.DataSize(), connection.MaxChunkSize)
}

func handleReceivedChunk(connection *conn.Conn, receivedChunk *chunk.Chunk) {
	log.Printf(
		"receivedChunk received Fmt: %d - "+
			"receivedChunk stream id: %d - "+
			"message type id: %d - "+
			"message stream id: %d - "+
			"message length: %d",
		receivedChunk.Header.BasicHeader.Fmt,
		receivedChunk.Header.BasicHeader.ChunkStreamId,
		receivedChunk.Header.MessageHeader.MessageTypeId,
		receivedChunk.Header.MessageHeader.MessageStreamId,
		receivedChunk.Header.MessageHeader.MessageLength,
	)
	if receivedChunk.Header.BasicHeader.Fmt <= 1 {
		connection.CurrentMessage.Length = receivedChunk.Header.MessageHeader.MessageLength
		connection.CurrentMessage.TypeId = receivedChunk.Header.MessageHeader.MessageTypeId
	}
	if receivedChunk.Header.BasicHeader.Fmt == 0 {
		connection.CurrentMessage.StreamId = receivedChunk.Header.MessageHeader.MessageStreamId
	}
}

func handleCompletedMessage(connection *conn.Conn, completedMessage *conn.Message) error {
	if completedMessage.TypeId == TypeSetChunkSize {
		connection.MaxChunkSize = binary.BigEndian.Uint32(connection.CurrentMessage.Data[0:4]) >> 1
	} else if completedMessage.TypeId == TypeAbortMessage {
		connection.CurrentMessage = nil
	} else if completedMessage.TypeId == TypeWindowAcknowledgementSize {
		connection.WindowAcknowledgementSize = binary.BigEndian.Uint32(connection.CurrentMessage.Data[0:4])
	} else if completedMessage.TypeId == TypeCommandMessageAmf0 {
		command, err := amf.DecodeCommand(connection.CurrentMessage.Data)
		if err != nil {
			return err
		}
		log.Printf("Command received: %s\n", command)
		if len(command.Parts) > 0 && command.Parts[0] == amf.NewString("connect") {
			err = doConnectMessageFlow(connection, err)
			if err != nil {
				return err
			}
		}
	}
	connection.Messages <- completedMessage
	connection.CurrentMessage = new(conn.Message)
	return nil
}

func doConnectMessageFlow(connection *conn.Conn, err error) error {
	// server sends window acknowledgement size
	sendWindowAcknowledgementSize := newWindowAcknowledgementSizeMessage(connection.SendWindowAcknowledgementSize)
	_, err = sendWindowAcknowledgementSize.Send(*connection)
	if err != nil {
		return err
	}
	// server sends set peer bandwidth
	setPeerBandwidthMessage := newSetPeerBandwidthMessage(connection.SendWindowAcknowledgementSize, SetPeerBandwidthLimitTypeHard)
	_, err = setPeerBandwidthMessage.Send(*connection)
	if err != nil {
		return err
	}
	return nil
}
