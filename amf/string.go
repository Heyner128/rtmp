package amf

import (
	"encoding/binary"
	"unicode/utf8"
)

var stringMarker = byte(0x02)

type AmfString string

func NewAmfString(str string) AmfString {
	return AmfString(str)
}

func decodeNextAmfString(bytes []byte) (int, AmfString) {
	length := 3 + int(binary.BigEndian.Uint16(bytes[1:3]))
	return length, AmfString(bytes[3:length])
}

func (str AmfString) Encode() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, stringMarker)
	bytes = binary.BigEndian.AppendUint16(bytes, uint16(utf8.RuneCountInString(string(str))))
	bytes = append(bytes, []byte(str)...)
	return bytes
}
