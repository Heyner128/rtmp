package amf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringAndNumberCommandEncode(t *testing.T) {
	testString := "test string"
	testNumber := 1234.5678
	amfString := NewString(testString)
	amfNumber := NewNumber(testNumber)
	amfMessage := NewCommand(amfString, amfNumber)
	assert.NotNil(t, amfMessage)
	assert.Equal(t, amfMessage.Encode(), append(amfString.Encode(), amfNumber.Encode()...))
}

func TestStringAndNumberCommandDecode(t *testing.T) {
	testString := "test string"
	testNumber := 1234.5678
	bytes := make([]byte, 0)
	bytes = append(bytes, NewString(testString).Encode()...)
	bytes = append(bytes, NewNumber(testNumber).Encode()...)
	amfCommand, _ := DecodeCommand(bytes)
	assert.NotNil(t, amfCommand)
	assert.Equal(t, amfCommand.Parts[0], NewString(testString))
	assert.Equal(t, amfCommand.Parts[1], NewNumber(testNumber))
}

func TestMultipleTypesDecode(t *testing.T) {
	testString := "test string"
	testNumber := 1234.5678
	testObject, _ := generateTestAmfObject()
	bytes := make([]byte, 0)
	bytes = append(bytes, NewString(testString).Encode()...)
	bytes = append(bytes, NewNumber(testNumber).Encode()...)
	bytes = append(bytes, NewObject(testObject).Encode()...)
	amfCommand, _ := DecodeCommand(bytes)
	assert.NotNil(t, amfCommand)
	assert.Equal(t, amfCommand.Parts[0], NewString(testString))
	assert.Equal(t, amfCommand.Parts[1], NewNumber(testNumber))
	assert.Equal(t, amfCommand.Parts[2], NewObject(testObject))
}
