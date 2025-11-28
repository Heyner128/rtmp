package message

import (
	"encoding/binary"
	"rtmp/amf"
	"rtmp/server"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageReceived(t *testing.T) {
	rtmpServer, clientConn := server.StartTestingServerWithHandshake(t)
	message := newMessage(uint8(9), uint32(123456), generateRandomBytes(1024))
	err := message.Send(t, clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	select {
	case messageReceived := <-serverConn.Messages:
		assert.Equal(t, message.MessageTypeId, messageReceived.TypeId)
		assert.Equal(t, message.MessageStreamId, messageReceived.StreamId)
		assert.Equal(t, message.Data, messageReceived.Data)
	case <-serverConn.Errors:
		t.FailNow()
	}
}

func TestSetChunkSizeAndMultiChunkRandomMessageReceived(t *testing.T) {
	rtmpServer, clientConn := server.StartTestingServerWithHandshake(t)
	newSize := uint32(100)
	setSizeMessage := newMessage(uint8(1), uint32(123456), binary.BigEndian.AppendUint32(make([]byte, 0), newSize<<1))
	randomDataMessage := newMessage(uint8(9), uint32(123456), generateRandomBytes(120))
	err := setSizeMessage.Send(t, clientConn)
	clientConn.MaxChunkSize = newSize
	assert.Nil(t, err)
	err = randomDataMessage.Send(t, clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	<-serverConn.Messages
	select {
	case messageReceived := <-serverConn.Messages:
		assert.Equal(t, newSize, serverConn.MaxChunkSize)
		assert.Equal(t, randomDataMessage.Data, messageReceived.Data)
		assert.Equal(t, randomDataMessage.MessageTypeId, messageReceived.TypeId)
		assert.Equal(t, randomDataMessage.MessageStreamId, messageReceived.StreamId)
	case <-serverConn.Errors:
		t.FailNow()
	}
}

func TestSetChunkSizeMessageReceived(t *testing.T) {
	rtmpServer, clientConn := server.StartTestingServerWithHandshake(t)
	newSize := uint32(1024)
	message := newMessage(uint8(1), uint32(123456), binary.BigEndian.AppendUint32(make([]byte, 0), newSize<<1))
	err := message.Send(t, clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	select {
	case <-serverConn.Messages:
		assert.Equal(t, newSize, serverConn.MaxChunkSize)
	case <-serverConn.Errors:
		t.FailNow()
	}
}

func TestAbortMessageReceived(t *testing.T) {
	rtmpServer, clientConn := server.StartTestingServerWithHandshake(t)
	message := newMessage(uint8(2), uint32(123456), binary.BigEndian.AppendUint32(make([]byte, 0), uint32(2)))
	err := message.Send(t, clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	select {
	case receivedMessage := <-serverConn.Messages:
		assert.Nil(t, receivedMessage)
	case <-serverConn.Errors:
		t.FailNow()
	}
}

func TestWindowAcknowledgementSizeMessageReceived(t *testing.T) {
	rtmpServer, clientConn := server.StartTestingServerWithHandshake(t)
	windowAcknowledgementSize := uint32(1024)
	message := newMessage(uint8(5), uint32(123456), binary.BigEndian.AppendUint32(make([]byte, 0), windowAcknowledgementSize))
	err := message.Send(t, clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	select {
	case <-serverConn.Messages:
		assert.Equal(t, windowAcknowledgementSize, serverConn.WindowAcknowledgementSize)
	case <-serverConn.Errors:
		t.FailNow()
	}
}

func TestCommandMessageReceived(t *testing.T) {
	rtmpServer, clientConn := server.StartTestingServerWithHandshake(t)
	connectCommand := amf.NewAmfCommand(
		amf.NewAmfString("connect"),
		amf.NewAmfNumber(1),
		amf.NewAmfObject(amf.AmfObject{
			amf.AmfObjectProperty{Name: "app", Value: amf.NewAmfString("testApp")},
			amf.AmfObjectProperty{Name: "objectEncoding", Value: amf.NewAmfNumber(0)},
		}),
	)
	message := newMessage(uint8(20), uint32(123456), connectCommand.Encode())
	err := message.Send(t, clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections

	responseObject := amf.NewAmfCommand(
		amf.NewAmfString("_result"),
		amf.NewAmfNumber(1),
	)
	select {
	case receivedMessage := <-serverConn.Messages:
		assert.Equal(t, message.Data, receivedMessage.Data)
	case <-serverConn.Errors:
		t.FailNow()
	}
	sentMessage := make([]byte, len(responseObject.Encode()))
	_, err = clientConn.Read(sentMessage)
	assert.Nil(t, err)
	assert.Equal(t, responseObject.Encode(), sentMessage)
}
