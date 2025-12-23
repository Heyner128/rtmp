package message

import (
	"encoding/binary"
	"errors"
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
	TypeUserControl               = uint8(4)
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

func NewMessage(messageTypeId uint8, messageStreamId uint32, data []byte) *Message {
	var chunkStreamId uint32
	if isProtocolControlMessage(messageTypeId) {
		chunkStreamId = 2
	} else {
		chunkStreamId = generateChunkStreamId()
	}
	return &Message{
		ChunkStreamId:   chunkStreamId,
		MessageTypeId:   messageTypeId,
		MessageStreamId: messageStreamId,
		Data:            data,
	}
}

func generateChunkStreamId() uint32 {
	minId := 3
	maxId := 65599
	return uint32(rand.Intn(maxId-minId) + minId)
}

func NewWindowAcknowledgementSizeMessage(messageStreamId uint32, acknowledgementSize int) *Message {
	return NewMessage(TypeWindowAcknowledgementSize, messageStreamId, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(acknowledgementSize)))
}

func NewSetPeerBandwidthMessage(messageStreamId uint32, size int, limitType uint8) *Message {
	content := binary.BigEndian.AppendUint32(make([]byte, 0), uint32(size))
	content = append(content, limitType)
	return NewMessage(TypeSetPeerBandwidth, messageStreamId, content)
}

func NewAcknowledgementMessage(messageStreamId uint32, acknowledgementSize int) *Message {
	return NewMessage(TypeAcknowledgement, messageStreamId, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(acknowledgementSize)))
}

func NewStreamBeginMessage(messageStreamId uint32) *Message {
	contents := make([]byte, 6)
	binary.BigEndian.PutUint16(contents[0:2], 0)
	binary.BigEndian.PutUint32(contents[2:6], messageStreamId)
	return NewMessage(TypeUserControl, 0, contents)
}

func (message *Message) Send(conn *conn.Conn) (int, error) {
	bytesSent := 0
	for _, nChunk := range message.BuildChunks(int(conn.MaxChunkSize)) {

		if conn.UnacknowledgedBytesSent > 0 && conn.PeerWindowAcknowledgementSize > 0 && conn.UnacknowledgedBytesSent >= conn.PeerWindowAcknowledgementSize {
			select {
			case clientAcknowledgement := <-conn.Messages:
				if clientAcknowledgement.TypeId == TypeAcknowledgement {
					conn.UnacknowledgedBytesSent = 0
				}
			case <-conn.Errors:
				return 0, errors.New("bandwidth size exceeded")
			}
		}
		encoded := nChunk.Encode()
		n, err := conn.Write(encoded)
		bytesSent += n
		conn.UnacknowledgedBytesSent += uint32(n)
		if err != nil {
			return 0, err
		}
	}
	return bytesSent, nil
}

func (message *Message) BuildChunks(chunkSize int) []chunk.Chunk {
	chunks := make([]chunk.Chunk, 0)
	for i := 0; i < len(message.Data); i += chunkSize {
		end := i + chunkSize
		if end > len(message.Data) {
			end = len(message.Data)
		}
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
			*chunk.NewChunk(*header, message.Data[i:end]),
		)
	}
	return chunks
}

func isProtocolControlMessage(messageTypeId uint8) bool {
	return messageTypeId == TypeSetChunkSize ||
		messageTypeId == TypeAbortMessage ||
		messageTypeId == TypeAcknowledgement ||
		messageTypeId == TypeWindowAcknowledgementSize ||
		messageTypeId == TypeSetPeerBandwidth ||
		messageTypeId == TypeUserControl
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
	connection.UnacknowledgedBytesReceived += uint32(len(receivedChunk.Encode()))
	if connection.WindowAcknowledgementSize > 0 && connection.UnacknowledgedBytesReceived >= connection.WindowAcknowledgementSize {
		acknowledgementMessage := NewAcknowledgementMessage(receivedChunk.Header.MessageHeader.MessageStreamId, int(connection.UnacknowledgedBytesReceived))
		_, err = acknowledgementMessage.Send(connection)
		connection.UnacknowledgedBytesReceived = 0
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
	messageLength := connection.CurrentMessage.Length
	if header.BasicHeader.Fmt == 0 || header.BasicHeader.Fmt == 1 {
		messageLength = header.MessageHeader.MessageLength
	}
	remainingBytes := messageLength - connection.CurrentMessage.DataSize()
	return min(remainingBytes, connection.MaxChunkSize)
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
		connection.MaxChunkSize = binary.BigEndian.Uint32(completedMessage.Data[0:4]) & 0x7FFFFFFF
	} else if completedMessage.TypeId == TypeAbortMessage {
		connection.CurrentMessage = nil
	} else if completedMessage.TypeId == TypeWindowAcknowledgementSize {
		connection.WindowAcknowledgementSize = binary.BigEndian.Uint32(completedMessage.Data[0:4])
	} else if completedMessage.TypeId == TypeSetPeerBandwidth {
		connection.PeerWindowAcknowledgementSize = binary.BigEndian.Uint32(connection.CurrentMessage.Data[0:4])
		windowAcknowledgementSizeMessage := NewWindowAcknowledgementSizeMessage(completedMessage.StreamId, int(connection.PeerWindowAcknowledgementSize))
		_, err := windowAcknowledgementSizeMessage.Send(connection)
		if err != nil {
			return err
		}
	} else if completedMessage.TypeId == TypeCommandMessageAmf0 {
		command, err := amf.DecodeCommand(completedMessage.Data)
		if err != nil {
			return err
		}
		log.Printf("Command received: %s\n", command)
		if len(command.Parts) > 0 && command.Parts[0] == amf.NewString("connect") {
			err = doConnectMessageFlow(connection, completedMessage.StreamId, err)
			if err != nil {
				return err
			}
		}
	}
	connection.Messages <- completedMessage
	connection.CurrentMessage = new(conn.Message)
	return nil
}

func doConnectMessageFlow(connection *conn.Conn, messageStreamId uint32, err error) error {
	// server sends window acknowledgement size
	windowAcknowledgementSizeMessage := NewWindowAcknowledgementSizeMessage(messageStreamId, int(connection.PeerWindowAcknowledgementSize))
	_, err = windowAcknowledgementSizeMessage.Send(connection)
	if err != nil {
		return err
	}
	// server sends set peer bandwidth
	setPeerBandwidthMessage := NewSetPeerBandwidthMessage(messageStreamId, int(connection.PeerWindowAcknowledgementSize), SetPeerBandwidthLimitTypeHard)
	_, err = setPeerBandwidthMessage.Send(connection)
	if err != nil {
		return err
	}
	streamBeginMessage := NewStreamBeginMessage(messageStreamId)
	_, err = streamBeginMessage.Send(connection)
	if err != nil {
		return err
	}
	serverProps := amf.NewObject(
		amf.ObjectProperty{Name: "fmsVer", Value: amf.NewString("FMS/3,0,1,123")},
		amf.ObjectProperty{Name: "capabilities", Value: amf.NewNumber(31)},
	)
	infoProps := amf.NewObject(
		amf.ObjectProperty{Name: "level", Value: amf.NewString("status")},
		amf.ObjectProperty{Name: "code", Value: amf.NewString("NetConnection.Connect.Success")},
		amf.ObjectProperty{Name: "description", Value: amf.NewString("Connection succeeded.")},
		amf.ObjectProperty{Name: "objectEncoding", Value: amf.NewNumber(0)},
	)

	resultCommand := amf.NewCommand(amf.NewString("_result"), amf.NewNumber(1), serverProps, infoProps)
	resultCommandMessage := NewMessage(TypeCommandMessageAmf0, messageStreamId, resultCommand.Encode())
	resultCommandMessage.ChunkStreamId = 3
	_, err = resultCommandMessage.Send(connection)
	if err != nil {
		return err
	}
	return nil
}
