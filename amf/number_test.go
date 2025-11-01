package amf

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumberDecoding(t *testing.T) {
	testNumber := 1234.5678
	bytes := make([]byte, 0)
	bytes = append(bytes, 0x00)
	bytes = binary.BigEndian.AppendUint64(bytes, math.Float64bits(testNumber))
	assert.Equal(t, decodeAmfNumber(bytes), newAmfNumber(testNumber))
}

func TestNumberEncoding(t *testing.T) {
	testNumber := 1234.5678
	amfNumber := newAmfNumber(testNumber)
	amfMessage := newAmfMessage(amfNumber)
	assert.NotNil(t, amfMessage)
	assert.Equal(t, decodeAmfNumber(amfNumber.encode()), amfNumber)
	assert.Equal(t, amfMessage.encode(), amfNumber.encode())
}
