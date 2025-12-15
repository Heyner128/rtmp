package amf

import (
	"encoding/binary"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestStringDecoding(t *testing.T) {
	testString := "test string"
	bytes := make([]byte, 0)
	bytes = append(bytes, 0x02)
	numberOfRunes := utf8.RuneCountInString(testString)
	bytes = binary.BigEndian.AppendUint16(bytes, uint16(numberOfRunes))
	bytes = append(bytes, []byte(testString)...)
	_, decodedString := decodeNextString(bytes)
	assert.Equal(t, decodedString, NewString(testString))
}

func TestStringEncoding(t *testing.T) {
	testString := "test string"
	amfString := NewString(testString)
	amfMessage := NewCommand(amfString)
	assert.NotNil(t, amfMessage)
	_, decodedString := decodeNextString(amfString.Encode())
	assert.Equal(t, decodedString, amfString)
	assert.Equal(t, amfMessage.Encode(), amfString.Encode())
}
