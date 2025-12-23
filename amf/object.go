package amf

import (
	"encoding/binary"
	"errors"
)

var objectMarker = byte(0x03)
var objectEndMarker = byte(0x09)

type Object []ObjectProperty

type ObjectProperty struct {
	Name  string
	Value ValueType
}

func (object Object) Encode() []byte {
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

func NewObject(object ...ObjectProperty) Object {
	return object
}

func decodeNextObject(bytes []byte) (int, Object, error) {
	if len(bytes) < 4 {
		return 0, nil, errors.New("Can't decode object, not enough bytes")
	}
	if bytes[0] != objectMarker {
		return 0, nil, errors.New("Can't decode object, object marker is not 0x03")
	}
	object := make(Object, 0)
	pointer := 1

	for {
		propertyNameLength := binary.BigEndian.Uint16(bytes[pointer : pointer+2])
		pointer += 2
		propertyName := string(bytes[pointer : pointer+int(propertyNameLength)])
		pointer += int(propertyNameLength)
		if propertyNameLength == 0 && bytes[pointer] == objectEndMarker {
			pointer++
			break
		}
		propertyValueLength, propertyValue := decodeNextValueType(bytes[pointer:])
		if propertyValue == nil {
			return 0, nil, errors.New("Can't decode object property value")
		}
		pointer += propertyValueLength

		object = append(object, ObjectProperty{propertyName, propertyValue})
	}
	return pointer, object, nil
}
