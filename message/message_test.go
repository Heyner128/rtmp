package message_test

import (
	"encoding/binary"
	"rtmp/message"
	"rtmp/testHelpers"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageReceived(t *testing.T) {
	rtmpServer, clientConn := testHelpers.StartTestingServerWithHandshake(t)
	testMessage := message.NewMessage(uint8(9), testHelpers.GenerateRandomBytes(1024))
	err := testMessage.Send(clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	select {
	case messageReceived := <-serverConn.Messages:
		assert.Equal(t, testMessage.MessageTypeId, messageReceived.TypeId)
		assert.Equal(t, testMessage.MessageStreamId, messageReceived.StreamId)
		assert.Equal(t, testMessage.Data, messageReceived.Data)
	case <-serverConn.Errors:
		t.FailNow()
	}
}

func TestSetChunkSizeAndMultiChunkRandomMessageReceived(t *testing.T) {
	rtmpServer, clientConn := testHelpers.StartTestingServerWithHandshake(t)
	newSize := uint32(100)
	setSizeMessage := message.NewMessage(uint8(1), binary.BigEndian.AppendUint32(make([]byte, 0), newSize<<1))
	randomDataMessage := message.NewMessage(uint8(9), testHelpers.GenerateRandomBytes(120))
	err := setSizeMessage.Send(clientConn)
	clientConn.MaxChunkSize = newSize
	assert.Nil(t, err)
	err = randomDataMessage.Send(clientConn)
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
	rtmpServer, clientConn := testHelpers.StartTestingServerWithHandshake(t)
	newSize := uint32(1024)
	testMessage := message.NewMessage(uint8(1), binary.BigEndian.AppendUint32(make([]byte, 0), newSize<<1))
	err := testMessage.Send(clientConn)
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
	rtmpServer, clientConn := testHelpers.StartTestingServerWithHandshake(t)
	testMessage := message.NewMessage(uint8(2), binary.BigEndian.AppendUint32(make([]byte, 0), uint32(2)))
	err := testMessage.Send(clientConn)
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
	rtmpServer, clientConn := testHelpers.StartTestingServerWithHandshake(t)
	windowAcknowledgementSize := 1024
	testMessage := testHelpers.GenerateTestWindowAcknowledgementSize(windowAcknowledgementSize)
	err := testMessage.Send(clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	select {
	case <-serverConn.Messages:
		assert.Equal(t, uint32(windowAcknowledgementSize), serverConn.WindowAcknowledgementSize)
	case <-serverConn.Errors:
		t.FailNow()
	}
}

func TestCommandMessageReceived(t *testing.T) {
	rtmpServer, clientConn := testHelpers.StartTestingServerWithHandshake(t)
	testMessage := testHelpers.GenerateTestConnectCommand()
	err := testMessage.Send(clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	select {
	case receivedMessage := <-serverConn.Messages:
		assert.Equal(t, testMessage.Data, receivedMessage.Data)
	case <-serverConn.Errors:
		t.FailNow()
	}
}

func TestConnectMessageFlow(t *testing.T) {
	rtmpServer, clientConn := testHelpers.StartTestingServerWithHandshake(t)
	testMessage := testHelpers.GenerateTestConnectCommand()
	err := testMessage.Send(clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	select {
	case <-serverConn.Messages:
	case <-serverConn.Errors:
		t.FailNow()
	}
	select {
	case sentTestMessage := <-clientConn.Messages:
		assert.Equal(t, binary.BigEndian.AppendUint32(make([]byte, 0), serverConn.SendWindowAcknowledgementSize), sentTestMessage.Data)
	case <-clientConn.Errors:
		t.FailNow()
	}
}
