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
		amf.NewObject(
			amf.ObjectProperty{Name: "app", Value: amf.NewString("testApp")},
			amf.ObjectProperty{Name: "objectEncoding", Value: amf.NewNumber(0)},
		),
	)
	return *message.NewMessage(message.TypeCommandMessageAmf0, connectCommand.Encode())
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
	return *message.NewMessage(message.TypeCommandMessageAmf0, connectCommand.Encode())
}

func GenerateTestRandomMessage(size int) message.Message {
	return *message.NewMessage(message.TypeVideo, generateRandomBytes(size))
}

func GenerateTestWindowAcknowledgementSize(acknowledgementSize int) message.Message {
	return *message.NewMessage(message.TypeWindowAcknowledgementSize, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(acknowledgementSize)))
}

func GenerateTestSetPeerBandwidth(size int, limitType uint8) message.Message {
	contents := binary.BigEndian.AppendUint32(make([]byte, 0), uint32(size))
	contents = append(contents, limitType)
	return *message.NewMessage(message.TypeSetPeerBandwidth, contents)
}

func generateRandomBytes(length int) []byte {
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(rand.Intn(255))
	}
	return bytes
}
