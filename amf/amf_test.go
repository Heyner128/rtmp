package amf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringAndNumberCommandEncode(t *testing.T) {
	testString := "test string"
	testNumber := 1234.5678
	amfString := NewAmfString(testString)
	amfNumber := NewAmfNumber(testNumber)
	amfMessage := NewAmfCommand(amfString, amfNumber)
	assert.NotNil(t, amfMessage)
	assert.Equal(t, amfMessage.Encode(), append(amfString.Encode(), amfNumber.Encode()...))
}

func TestStringAndNumberCommandDecode(t *testing.T) {
	testString := "test string"
	testNumber := 1234.5678
	bytes := make([]byte, 0)
	bytes = append(bytes, NewAmfString(testString).Encode()...)
	bytes = append(bytes, NewAmfNumber(testNumber).Encode()...)
	amfCommand, _ := DecodeAmfCommand(bytes)
	assert.NotNil(t, amfCommand)
	assert.Equal(t, amfCommand.parts[0], NewAmfString(testString))
	assert.Equal(t, amfCommand.parts[1], NewAmfNumber(testNumber))
}

func TestMultipleTypesDecode(t *testing.T) {
	testString := "test string"
	testNumber := 1234.5678
	testObject, _ := generateTestAmfObject()
	bytes := make([]byte, 0)
	bytes = append(bytes, NewAmfString(testString).Encode()...)
	bytes = append(bytes, NewAmfNumber(testNumber).Encode()...)
	bytes = append(bytes, NewAmfObject(testObject).Encode()...)
	amfCommand, _ := DecodeAmfCommand(bytes)
	assert.NotNil(t, amfCommand)
	assert.Equal(t, amfCommand.parts[0], NewAmfString(testString))
	assert.Equal(t, amfCommand.parts[1], NewAmfNumber(testNumber))
	assert.Equal(t, amfCommand.parts[2], NewAmfObject(testObject))
}
