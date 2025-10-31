package amf0

var stringMarker = byte(0x02)

var numberMarker = byte(0x00)

type AmfType interface {
	buffer() []byte
}

type AmfString string

type AmfNumber float64

func NewAmfNumber(number float64) AmfNumber {
	return AmfNumber(number)
}

func (number AmfNumber) buffer() []byte {
	return []byte{0x00}
}

func NewAmfString(str string) AmfString {
	return AmfString(str)
}

func (str AmfString) buffer() []byte {
	return []byte(str)
}

type AmfMessage struct {
	part AmfType
}

func (msg AmfMessage) buffer() []byte {
	var marker byte
	switch msg.part.(type) {
	case AmfNumber:
		marker = numberMarker
	case AmfString:
		marker = stringMarker
	}
	return append([]byte{marker}, msg.part.buffer()...)
}

func NewAmfMessage(part AmfType) AmfMessage {
	return AmfMessage{
		part: part,
	}
}
