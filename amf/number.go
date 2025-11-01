package amf

import (
	"encoding/binary"
	"math"
)

var numberMarker = byte(0x00)

type AmfNumber float64

func newAmfNumber(number float64) AmfNumber {
	return AmfNumber(number)
}

func (number AmfNumber) encode() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, numberMarker)
	bytes = binary.BigEndian.AppendUint64(bytes, math.Float64bits(float64(number)))
	return bytes
}

func decodeAmfNumber(bytes []byte) AmfNumber {
	return AmfNumber(math.Float64frombits(binary.BigEndian.Uint64(bytes[1:])))
}
