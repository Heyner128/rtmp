package testutil

import (
	"math/rand"
	"rtmp/amf"
	"rtmp/message"
)

func GenerateTestConnectCommand() message.Message {
	connectCommand := amf.NewCommand(
		amf.NewString("connect"),
		amf.NewNumber(1),
		amf.NewObject(
			amf.ObjectProperty{Name: "app", Value: amf.NewString("testApp")},
			amf.ObjectProperty{Name: "objectEncoding", Value: amf.NewNumber(0)},
		),
	)
	return *message.NewMessage(message.TypeCommandMessageAmf0, rand.Uint32(), connectCommand.Encode())
}

func GenerateTestUnknownCommand() message.Message {
	connectCommand := amf.NewCommand(
		amf.NewString("notacommand"),
		amf.NewNumber(9),
		amf.NewObject(
			amf.ObjectProperty{Name: "prop1", Value: amf.NewString("testApp")},
			amf.ObjectProperty{Name: "prop2", Value: amf.NewNumber(0)},
		),
	)
	return *message.NewMessage(message.TypeCommandMessageAmf0, rand.Uint32(), connectCommand.Encode())
}

func GenerateTestRandomMessage(size int) message.Message {
	return *message.NewMessage(message.TypeVideo, rand.Uint32(), generateRandomBytes(size))
}

func generateRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(rand.Intn(255))
	}
	return bytes
}
