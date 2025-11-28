package amf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBooleanEncoding(t *testing.T) {
	testBoolean := uint8(1)
	amfBoolean := NewAmfBoolean(testBoolean)
	amfMessage := NewAmfCommand(amfBoolean)
	assert.NotNil(t, amfMessage)
	encodedBoolean := amfBoolean.Encode()
	encodedMessage := amfMessage.Encode()
	assert.NotNil(t, encodedBoolean)
	assert.Equal(t, 2, len(encodedBoolean))
	assert.Equal(t, booleanMarker, encodedBoolean[0])
	assert.Equal(t, testBoolean, encodedBoolean[1])
	_, decodedBoolean := decodeNextAmfBoolean(encodedBoolean)
	assert.Equal(t, amfBoolean, decodedBoolean)
	assert.Equal(t, encodedMessage, encodedBoolean)
}

func TestBooleanDecoding(t *testing.T) {
	testBoolean := uint8(0)
	bytes := make([]byte, 0)
	bytes = append(bytes, 0x01)
	bytes = append(bytes, testBoolean)
	_, decodedBoolean := decodeNextAmfBoolean(bytes)
	assert.Equal(t, decodedBoolean, NewAmfBoolean(testBoolean))
}
