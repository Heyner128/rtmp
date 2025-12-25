package message

import (
	"encoding/binary"
	"errors"
	"math/rand"
	"rtmp/amf"
	"rtmp/chunk"
	"rtmp/conn"
	"rtmp/logger"
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

func NewWindowAcknowledgementSizeMessage(acknowledgementSize int) *Message {
	return NewMessage(TypeWindowAcknowledgementSize, 0, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(acknowledgementSize)))
}

func NewSetPeerBandwidthMessage(size int, limitType uint8) *Message {
	content := binary.BigEndian.AppendUint32(make([]byte, 0), uint32(size))
	content = append(content, limitType)
	return NewMessage(TypeSetPeerBandwidth, 0, content)
}

func NewAcknowledgementMessage(acknowledgementSize int) *Message {
	return NewMessage(TypeAcknowledgement, 0, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(acknowledgementSize)))
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
			ackChunk, err := Accept(conn)
			if err != nil {
				return 0, errors.New("bandwith exceeded")
			}
			if ackChunk.Header.MessageHeader.MessageTypeId == TypeAcknowledgement {
				conn.UnacknowledgedBytesSent = 0
			}
		}
		if nChunk.Header.BasicHeader.Fmt == 0 && nChunk.Header.MessageHeader.MessageTypeId == TypeSetChunkSize {
			conn.PeerMaxChunkSize = binary.BigEndian.Uint32(nChunk.Data[0:4]) & 0x7FFFFFFF
		}
		logger.Get().Debugf("sending chunk %v", nChunk)
		encoded := nChunk.Encode()
		logger.Get().Debugf("chunk bytes %x", encoded)
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
	dataSize, err := getNextChunkDataSize(connection, header)
	if err != nil {
		return nil, err
	}
	data := make(
		[]byte,
		dataSize,
	)
	_, err = connection.Read(data)
	if err != nil {
		return nil, err
	}
	receivedChunk := chunk.NewChunk(*header, data)
	connection.UnacknowledgedBytesReceived += uint32(len(receivedChunk.Encode()))
	if connection.WindowAcknowledgementSize > 0 && connection.UnacknowledgedBytesReceived >= connection.WindowAcknowledgementSize {
		acknowledgementMessage := NewAcknowledgementMessage(int(connection.UnacknowledgedBytesReceived))
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

func getNextChunkDataSize(connection *conn.Conn, header *chunk.Header) (uint32, error) {
	messageLength := connection.CurrentMessage.Length
	if header.BasicHeader.Fmt == 0 || header.BasicHeader.Fmt == 1 {
		messageLength = header.MessageHeader.MessageLength
	}
	remainingBytes := messageLength - connection.CurrentMessage.DataSize()
	if remainingBytes <= 0 {
		return 0, errors.New("no data to read")
	}
	return min(remainingBytes, connection.PeerMaxChunkSize), nil
}

func handleReceivedChunk(connection *conn.Conn, receivedChunk *chunk.Chunk) {
	logger.Get().Debugf("received chunk %v", receivedChunk)
	if receivedChunk.Header.BasicHeader.Fmt <= 1 {
		connection.CurrentMessage.Length = receivedChunk.Header.MessageHeader.MessageLength
		connection.CurrentMessage.TypeId = receivedChunk.Header.MessageHeader.MessageTypeId
	}
	if receivedChunk.Header.BasicHeader.Fmt == 0 {
		connection.CurrentMessage.StreamId = receivedChunk.Header.MessageHeader.MessageStreamId
	}
}

func handleCompletedMessage(connection *conn.Conn, completedMessage *conn.Message) error {
	logger.Get().Debugf("received completed message %v", completedMessage)
	if completedMessage.TypeId == TypeSetChunkSize {
		connection.MaxChunkSize = binary.BigEndian.Uint32(completedMessage.Data[0:4]) & 0x7FFFFFFF
		peerMaxChunkSizeMessage := NewMessage(TypeSetChunkSize, 0, binary.BigEndian.AppendUint32(make([]byte, 0), connection.MaxChunkSize))
		_, err := peerMaxChunkSizeMessage.Send(connection)
		if err != nil {
			return err
		}
	} else if completedMessage.TypeId == TypeAbortMessage {
		connection.CurrentMessage = nil
	} else if completedMessage.TypeId == TypeWindowAcknowledgementSize {
		connection.WindowAcknowledgementSize = binary.BigEndian.Uint32(completedMessage.Data[0:4])
	} else if completedMessage.TypeId == TypeSetPeerBandwidth {
		connection.PeerWindowAcknowledgementSize = binary.BigEndian.Uint32(connection.CurrentMessage.Data[0:4])
		windowAcknowledgementSizeMessage := NewWindowAcknowledgementSizeMessage(int(connection.PeerWindowAcknowledgementSize))
		_, err := windowAcknowledgementSizeMessage.Send(connection)
		if err != nil {
			return err
		}
	} else if completedMessage.TypeId == TypeAcknowledgement {
		connection.UnacknowledgedBytesSent = 0
	} else if completedMessage.TypeId == TypeCommandMessageAmf0 {
		command, err := amf.DecodeCommand(completedMessage.Data)
		if err != nil {
			return err
		}
		logger.Get().Debugf("Command received: %s\n", command)
		if len(command.Parts) > 0 && command.Parts[0] == amf.NewString("connect") {
			err = doConnectMessageFlow(connection, rand.Uint32())

		} else if len(command.Parts) > 0 && command.Parts[0] == amf.NewString("createStream") {
			err = doCreateStreamFlow(connection, rand.Uint32(), *command)
		} else if len(command.Parts) > 0 && command.Parts[0] == amf.NewString("publish") {
			err = doPublishFlow(connection, rand.Uint32(), *command)
		}
		if err != nil {
			return err
		}
	}
	select {
	case connection.Messages <- connection.CurrentMessage:
	default:
	}
	connection.CurrentMessage = new(conn.Message)
	// preserves the data of the last message
	connection.CurrentMessage.Length = completedMessage.Length
	connection.CurrentMessage.TypeId = completedMessage.TypeId
	connection.CurrentMessage.StreamId = completedMessage.StreamId
	return nil
}

func doConnectMessageFlow(connection *conn.Conn, messageStreamId uint32) error {
	// server sends window acknowledgement size
	windowAcknowledgementSizeMessage := NewWindowAcknowledgementSizeMessage(int(connection.PeerWindowAcknowledgementSize))
	_, err := windowAcknowledgementSizeMessage.Send(connection)
	if err != nil {
		return err
	}
	// server sends set peer bandwidth
	setPeerBandwidthMessage := NewSetPeerBandwidthMessage(int(connection.PeerWindowAcknowledgementSize), SetPeerBandwidthLimitTypeHard)
	_, err = setPeerBandwidthMessage.Send(connection)
	connection.WindowAcknowledgementSize = connection.PeerWindowAcknowledgementSize
	if err != nil {
		return err
	}
	// server sends stream begin message
	streamBeginMessage := NewStreamBeginMessage(messageStreamId)
	_, err = streamBeginMessage.Send(connection)
	if err != nil {
		return err
	}
	// server sends result command
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
	resultCommandMessage := NewMessage(TypeCommandMessageAmf0, uint32(0), resultCommand.Encode())
	resultCommandMessage.ChunkStreamId = uint32(3)
	_, err = resultCommandMessage.Send(connection)
	if err != nil {
		return err
	}
	return nil
}

func doCreateStreamFlow(connection *conn.Conn, messageStreamId uint32, command amf.Command) error {
	// sends result command
	transactionId := command.Parts[1]
	resultCommand := amf.NewCommand(amf.NewString("_result"), transactionId, amf.NewNull(), amf.NewNumber(float64(messageStreamId)))
	resultCommandMessage := NewMessage(TypeCommandMessageAmf0, uint32(0), resultCommand.Encode())
	resultCommandMessage.ChunkStreamId = uint32(3)
	_, err := resultCommandMessage.Send(connection)
	if err != nil {
		return err
	}
	return nil
}

func doPublishFlow(connection *conn.Conn, messageStreamId uint32, command amf.Command) error {
	infoProps := amf.NewObject(
		amf.ObjectProperty{Name: "level", Value: amf.NewString("status")},
		amf.ObjectProperty{Name: "code", Value: amf.NewString("NetConnection.Publish.Start")},
		amf.ObjectProperty{Name: "description", Value: amf.NewString("Publish flow started.")},
	)

	statusCommand := amf.NewCommand(amf.NewString("onStatus"), amf.NewNumber(0), amf.NewNull(), infoProps)
	statusCommandMessage := NewMessage(TypeCommandMessageAmf0, uint32(0), statusCommand.Encode())
	statusCommandMessage.ChunkStreamId = uint32(3)
	_, err := statusCommandMessage.Send(connection)
	// server sends stream begin message
	streamBeginMessage := NewStreamBeginMessage(messageStreamId)
	_, err = streamBeginMessage.Send(connection)
	if err != nil {
		return err
	}
	// sends result command
	transactionId := command.Parts[1]
	resultCommand := amf.NewCommand(amf.NewString("_result"), transactionId, amf.NewNull(), amf.NewNumber(float64(messageStreamId)))
	resultCommandMessage := NewMessage(TypeCommandMessageAmf0, uint32(0), resultCommand.Encode())
	resultCommandMessage.ChunkStreamId = uint32(3)
	_, err = resultCommandMessage.Send(connection)
	if err != nil {
		return err
	}
	//sets message data
	connection.CurrentMessage.StreamId = messageStreamId
	connection.CurrentMessage.TypeId = TypeVideo
	return nil
}
