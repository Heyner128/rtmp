package amf

var booleanMarker = byte(0x01)

type Boolean uint8

func NewBoolean(bool uint8) Boolean {
	return Boolean(bool)
}

func (bool Boolean) Encode() []byte {
	return []byte{booleanMarker, uint8(bool)}
}

func decodeNextBoolean(bytes []byte) (int, Boolean) {
	length := 2
	return length, Boolean(bytes[length-1])
}
