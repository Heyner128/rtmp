package amf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBooleanEncoding(t *testing.T) {
	testBoolean := uint8(1)
	amfBoolean := newAmfBoolean(testBoolean)
	amfMessage := newAmfMessage(amfBoolean)
	assert.NotNil(t, amfMessage)
	assert.Equal(t, decodeAmfBoolean(amfBoolean.encode()), amfBoolean)
	assert.Equal(t, amfMessage.encode(), amfBoolean.encode())
}

func TestBooleanDecoding(t *testing.T) {
	testBoolean := uint8(1)
	bytes := make([]byte, 0)
	bytes = append(bytes, testBoolean)
	assert.Equal(t, decodeAmfBoolean(bytes), newAmfBoolean(testBoolean))
}
