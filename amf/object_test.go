package amf

import (
	"encoding/binary"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestObjectEncoding(t *testing.T) {
	object, _ := generateTestAmfObject()
	bytes := object.Encode()
	assert.NotNil(t, bytes)
	_, decodedObject, _ := decodeNextObject(bytes)
	assert.Equal(t, decodedObject, object)
}

func TestObjectDecodingFailWrongValueMarker(t *testing.T) {
	bytes := make([]byte, 0)
	bytes = append(bytes, 0x01)
	bytes = append(bytes, []byte{0x00, 0x00}...)
	bytes = append(bytes, 0x10)
	_, _, err := decodeNextObject(bytes)
	assert.Error(t, err)
}

func TestObjectDecodingFailWrongMarker(t *testing.T) {
	bytes := make([]byte, 0)
	bytes = append(bytes, 0x01)
	bytes = append(bytes, []byte{0x00, 0x00}...)
	bytes = append(bytes, 0x09)
	_, _, err := decodeNextObject(bytes)
	assert.Error(t, err)
}

func TestObjectDecodingFailNotEnoughBytes(t *testing.T) {
	bytes := make([]byte, 0)
	bytes = append(bytes, 0x00)
	_, _, err := decodeNextObject(bytes)
	assert.Error(t, err)
}

func TestObjectDecoding(t *testing.T) {
	amfObject, bytes := generateTestAmfObject()

	_, object, err := decodeNextObject(bytes)

	assert.NoError(t, err)

	assert.Equal(t, object, NewObject(amfObject...))
}

func generateTestAmfObject() (Object, []byte) {
	amfObject := make([]ObjectProperty, 3)
	amfObject[0] = ObjectProperty{"propertyOne", NewString("propertyOneValue")}
	amfObject[1] = ObjectProperty{"propertyTwo", NewString("propertyTwoValue")}
	amfObject[2] = ObjectProperty{"propertyThree", NewBoolean(0)}
	bytes := make([]byte, 0)
	bytes = append(bytes, 0x03)
	// property one
	bytes = binary.BigEndian.AppendUint16(bytes, uint16(utf8.RuneCountInString("propertyOne")))
	bytes = append(bytes, []byte("propertyOne")...)
	bytes = append(bytes, amfObject[0].Value.Encode()...)

	// property two
	bytes = binary.BigEndian.AppendUint16(bytes, uint16(utf8.RuneCountInString("propertyTwo")))
	bytes = append(bytes, []byte("propertyTwo")...)
	bytes = append(bytes, amfObject[1].Value.Encode()...)

	//property three
	bytes = binary.BigEndian.AppendUint16(bytes, uint16(utf8.RuneCountInString("propertyThree")))
	bytes = append(bytes, []byte("propertyThree")...)
	bytes = append(bytes, amfObject[2].Value.Encode()...)

	//end
	bytes = append(bytes, []byte{0x00, 0x00}...)
	bytes = append(bytes, 0x09)
	return amfObject, bytes
}
