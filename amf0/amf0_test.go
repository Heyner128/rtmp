package amf0

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringDecoding(t *testing.T) {
	testString := "test string"
	str := NewAmfString(testString)
	message := NewAmfMessage(str)
	assert.NotNil(t, message)
	assert.Equal(t, AmfString(testString), str)
	assert.Equal(t, message.buffer(), append([]byte{0x02}, str.buffer()...))
}

func TestNumberDecoding(t *testing.T) {
	testNumber := 1234.5678
	number := NewAmfNumber(testNumber)
	message := NewAmfMessage(number)
	assert.NotNil(t, message)
	assert.Equal(t, AmfNumber(testNumber), number)
	assert.Equal(t, message.buffer(), append([]byte{0x00}, number.buffer()...))
}
