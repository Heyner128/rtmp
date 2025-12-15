package testutil

import (
	"encoding/binary"
	"math/rand"
	"rtmp/amf"
	"rtmp/message"
)

func GenerateTestConnectCommand() message.Message {
	connectCommand := amf.NewCommand(
		amf.NewString("connect"),
		amf.NewNumber(1),
		amf.NewObject([]amf.ObjectProperty{
			{Name: "app", Value: amf.NewString("testApp")},
			{Name: "objectEncoding", Value: amf.NewNumber(0)},
		}),
	)
	return *message.NewMessage(message.TypeCommandMessageAmf0, connectCommand.Encode())
}

func GenerateTestUnknownCommand() message.Message {
	connectCommand := amf.NewCommand(
		amf.NewString("notacommand"),
		amf.NewNumber(9),
		amf.NewObject([]amf.ObjectProperty{
			{Name: "prop1", Value: amf.NewString("testApp")},
			{Name: "prop2", Value: amf.NewNumber(0)},
		}),
	)
	return *message.NewMessage(message.TypeCommandMessageAmf0, connectCommand.Encode())
}

func GenerateTestRandomMessage(size int) message.Message {
	return *message.NewMessage(message.TypeVideo, generateRandomBytes(size))
}

func GenerateTestWindowAcknowledgementSize(acknowledgementSize int) message.Message {
	return *message.NewMessage(message.TypeWindowAcknowledgementSize, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(acknowledgementSize)))
}

func generateRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(rand.Intn(255))
	}
	return bytes
}
