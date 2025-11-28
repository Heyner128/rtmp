package amf

import (
	"encoding/binary"
	"errors"
	"iter"
	"math"
)

var objectMarker = byte(0x03)
var objectEndMarker = byte(0x09)

type AmfObject []AmfObjectProperty

type AmfObjectProperty struct {
	Name  string
	Value AmfValueType
}

func (object AmfObject) Encode() []byte {
	bytes := make([]byte, 0)
	bytes = append(bytes, objectMarker)
	for _, property := range object {
		bytes = binary.BigEndian.AppendUint16(bytes, uint16(len(property.Name)))
		bytes = append(bytes, []byte(property.Name)...)
		bytes = append(bytes, property.Value.Encode()...)
	}
	bytes = append(bytes, []byte{0x00, 0x00}...)
	bytes = append(bytes, objectEndMarker)
	return bytes
}

func NewAmfObject(object []AmfObjectProperty) AmfObject {
	return object
}

func decodeNextAmfObject(bytes []byte) (int, AmfObject, error) {
	if len(bytes) < 4 {
		return 0, nil, errors.New("Can't decode object, not enough bytes")
	}
	if bytes[0] != objectMarker {
		return 0, nil, errors.New("Can't decode object, object marker is not 0x03")
	}
	object := make(AmfObject, 0)
	totalLength := 0
	for length, property := range decodeObjectProperties(bytes) {
		object = append(object, property)
		totalLength += length
	}
	return totalLength, object, nil
}

func decodeObjectProperties(bytes []byte) iter.Seq2[int, AmfObjectProperty] {
	return func(yield func(int, AmfObjectProperty) bool) {
		pointer := 1
		maxObjectLength := math.MaxUint16
		for {
			propertyNameLength := binary.BigEndian.Uint16(bytes[pointer : pointer+2])
			pointer += 2
			propertyName := string(bytes[pointer : pointer+int(propertyNameLength)])
			pointer += int(propertyNameLength)
			if (bytes[pointer] == objectEndMarker && propertyNameLength == uint16(0)) || pointer >= maxObjectLength {
				pointer++
				break
			}
			propertyValueLength, propertyValue := decodeNextAmfType(bytes[pointer:])
			pointer += propertyValueLength
			property := AmfObjectProperty{propertyName, propertyValue}
			if !yield(pointer, property) {
				return
			}
		}
	}
}
