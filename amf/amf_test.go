package amf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringAndNumberMessageEncode(t *testing.T) {
	testString := "test string"
	testNumber := 1234.5678
	amfString := newAmfString(testString)
	amfNumber := newAmfNumber(testNumber)
	amfMessage := newAmfMessage(amfString, amfNumber)
	assert.NotNil(t, amfMessage)
	assert.Equal(t, amfMessage.encode(), append(amfString.encode(), amfNumber.encode()...))
}

// TODO test decode message with multi part
