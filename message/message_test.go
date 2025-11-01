package message

import (
	"encoding/binary"
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
	case err = <-serverConn.Errors:
		assert.Nil(t, err)
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
	case err = <-serverConn.Errors:
		assert.Nil(t, err)
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
	case err = <-serverConn.Errors:
		assert.Nil(t, err)
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
	case err = <-serverConn.Errors:
		assert.Nil(t, err)
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
	case err = <-serverConn.Errors:
		assert.Nil(t, err)
	}
}
