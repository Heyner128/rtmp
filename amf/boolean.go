package amf

var booleanMarker = byte(0x01)

type AmfBoolean uint8

func NewAmfBoolean(bool uint8) AmfBoolean {
	return AmfBoolean(bool)
}

func (bool AmfBoolean) Encode() []byte {
	return []byte{booleanMarker, uint8(bool)}
}

func decodeNextAmfBoolean(bytes []byte) (int, AmfBoolean) {
	length := 2
	return length, AmfBoolean(bytes[length-1])
}
