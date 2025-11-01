package amf

type AmfValueType interface {
	encode() []byte
}

type AmfMessage struct {
	parts []AmfValueType
}

func (msg AmfMessage) encode() []byte {
	bytes := make([]byte, 0)
	for _, part := range msg.parts {
		bytes = append(bytes, part.encode()...)
	}
	return bytes
}

func newAmfMessage(parts ...AmfValueType) AmfMessage {
	return AmfMessage{
		parts: parts,
	}
}
