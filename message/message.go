package message

type Message struct {
	MessageTypeId   uint8
	MessageStreamId uint32
	Data            []byte
}

func NewMessage(messageTypeId uint8, messageStreamId uint32, data []byte) *Message {
	return &Message{
		MessageTypeId:   messageTypeId,
		MessageStreamId: messageStreamId,
		Data:            data,
	}
}
