package amf

import (
	"encoding/binary"
	"math"
)

var numberMarker = byte(0x00)

type Number float64

func NewNumber(number float64) Number {
	return Number(number)
}

func (number Number) Encode() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, numberMarker)
	bytes = binary.BigEndian.AppendUint64(bytes, math.Float64bits(float64(number)))
	return bytes
}

func decodeNextNumber(bytes []byte) (int, Number) {
	length := 9
	return length, Number(math.Float64frombits(binary.BigEndian.Uint64(bytes[1:length])))
}
