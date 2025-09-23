package message

import (
	"rtmp/server"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageReceived(t *testing.T) {
	rtmpServer, clientConn := server.StartTestingServerWithHandshake(t)
	message := NewMessage(uint8(1), uint32(123456), generateRandomBytes(1024))
	err := SendMessage(t, *message, clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	select {
	case messageReceived := <-serverConn.Messages:
		assert.Equal(t, message.MessageTypeId, messageReceived.MessageTypeId)
		assert.Equal(t, message.MessageStreamId, messageReceived.MessageStreamId)
		assert.Equal(t, message.Data, messageReceived.Data)
	case err = <-serverConn.Errors:
		assert.Nil(t, err)
	}
}
