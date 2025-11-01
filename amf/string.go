package amf

import (
	"encoding/binary"
	"unicode/utf8"
)

var stringMarker = byte(0x02)

type AmfString string

func newAmfString(str string) AmfString {
	return AmfString(str)
}

func decodeAmfString(bytes []byte) AmfString {
	length := binary.BigEndian.Uint16(bytes[1:3])
	return AmfString(bytes[3 : 3+length])
}

func (str AmfString) encode() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, stringMarker)
	bytes = binary.BigEndian.AppendUint16(bytes, uint16(utf8.RuneCountInString(string(str))))
	bytes = append(bytes, []byte(str)...)
	return bytes
}
