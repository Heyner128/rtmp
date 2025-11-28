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
	_, decodedNumber := decodeNextAmfNumber(bytes)
	assert.Equal(t, decodedNumber, NewAmfNumber(testNumber))
}

func TestNumberEncoding(t *testing.T) {
	testNumber := 1234.5678
	amfNumber := NewAmfNumber(testNumber)
	amfMessage := NewAmfCommand(amfNumber)
	assert.NotNil(t, amfMessage)
	_, decodedNumber := decodeNextAmfNumber(amfNumber.Encode())
	assert.Equal(t, decodedNumber, amfNumber)
	assert.Equal(t, amfMessage.Encode(), amfNumber.Encode())
}
