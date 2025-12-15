package amf

import (
	"encoding/binary"
	"unicode/utf8"
)

var stringMarker = byte(0x02)

type String string

func NewString(str string) String {
	return String(str)
}

func decodeNextString(bytes []byte) (int, String) {
	length := 3 + int(binary.BigEndian.Uint16(bytes[1:3]))
	return length, String(bytes[3:length])
}

func (str String) Encode() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, stringMarker)
	bytes = binary.BigEndian.AppendUint16(bytes, uint16(utf8.RuneCountInString(string(str))))
	bytes = append(bytes, []byte(str)...)
	return bytes
}
