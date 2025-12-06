package testHelpers

import (
	"encoding/binary"
	"math/rand"
	"rtmp/amf"
	"rtmp/message"
)

func GenerateTestConnectCommand() message.Message {
	connectCommand := amf.NewAmfCommand(
		amf.NewAmfString("connect"),
		amf.NewAmfNumber(1),
		amf.NewAmfObject(amf.AmfObject{
			amf.AmfObjectProperty{Name: "app", Value: amf.NewAmfString("testApp")},
			amf.AmfObjectProperty{Name: "objectEncoding", Value: amf.NewAmfNumber(0)},
		}),
	)
	return *message.NewMessage(uint8(20), connectCommand.Encode())
}

func GenerateTestWindowAcknowledgementSize(acknowledgementSize int) message.Message {
	return *message.NewMessage(uint8(5), binary.BigEndian.AppendUint32(make([]byte, 0), uint32(acknowledgementSize)))
}

func GenerateRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(rand.Intn(255))
	}
	return bytes
}
