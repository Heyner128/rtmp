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
	_, decodedString := decodeNextAmfString(bytes)
	assert.Equal(t, decodedString, NewAmfString(testString))
}

func TestStringEncoding(t *testing.T) {
	testString := "test string"
	amfString := NewAmfString(testString)
	amfMessage := NewAmfCommand(amfString)
	assert.NotNil(t, amfMessage)
	_, decodedString := decodeNextAmfString(amfString.Encode())
	assert.Equal(t, decodedString, amfString)
	assert.Equal(t, amfMessage.Encode(), amfString.Encode())
}
