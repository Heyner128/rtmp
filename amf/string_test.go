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
	bytes = binary.BigEndian.AppendUint16(bytes, uint16(utf8.RuneCountInString(testString)))
	bytes = append(bytes, []byte(testString)...)
	assert.Equal(t, decodeAmfString(bytes), newAmfString(testString))
}

func TestStringEncoding(t *testing.T) {
	testString := "test string"
	amfString := newAmfString(testString)
	amfMessage := newAmfMessage(amfString)
	assert.NotNil(t, amfMessage)
	assert.Equal(t, decodeAmfString(amfString.encode()), amfString)
	assert.Equal(t, amfMessage.encode(), amfString.encode())
}
