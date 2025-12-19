package message_test

import (
	"encoding/binary"
	"rtmp/message"
	"rtmp/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageReceived(t *testing.T) {
	rtmpServer, clientConn := testutil.StartTestingServerWithHandshake(t)
	testMessage := testutil.GenerateTestRandomMessage(1024)
	_, err := testMessage.Send(clientConn)
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
	rtmpServer, clientConn := testutil.StartTestingServerWithHandshake(t)
	newSize := uint32(100)
	setSizeMessage := message.NewMessage(uint8(1), binary.BigEndian.AppendUint32(make([]byte, 0), newSize<<1))
	randomDataMessage := testutil.GenerateTestRandomMessage(120)
	_, err := setSizeMessage.Send(clientConn)
	clientConn.MaxChunkSize = newSize
	assert.Nil(t, err)
	_, err = randomDataMessage.Send(clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	<-serverConn.Messages
	select {
	case <-serverConn.Messages:
		assert.Equal(t, newSize, serverConn.MaxChunkSize)
	case <-serverConn.Errors:
		t.FailNow()
	}
}

func TestSetChunkSizeMessageReceived(t *testing.T) {
	rtmpServer, clientConn := testutil.StartTestingServerWithHandshake(t)
	newSize := uint32(1024)
	testMessage := message.NewMessage(message.TypeSetChunkSize, binary.BigEndian.AppendUint32(make([]byte, 0), newSize<<1))
	_, err := testMessage.Send(clientConn)
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
	rtmpServer, clientConn := testutil.StartTestingServerWithHandshake(t)
	testMessage := message.NewMessage(message.TypeAbortMessage, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(2)))
	_, err := testMessage.Send(clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	select {
	case receivedMessage := <-serverConn.Messages:
		assert.NotNil(t, receivedMessage)
		assert.Equal(t, 0, len(serverConn.CurrentMessage.Data))
	case <-serverConn.Errors:
		t.FailNow()
	}
}

func TestWindowAcknowledgementSizeMessageReceived(t *testing.T) {
	rtmpServer, clientConn := testutil.StartTestingServerWithHandshake(t)
	windowAcknowledgementSize := 1024
	windowAcknowledgementSizeMessage := message.NewWindowAcknowledgementSizeMessage(windowAcknowledgementSize)
	_, err := windowAcknowledgementSizeMessage.Send(clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	select {
	case <-serverConn.Messages:
		assert.Equal(t, uint32(windowAcknowledgementSize), serverConn.WindowAcknowledgementSize)
	case <-serverConn.Errors:
		t.FailNow()
	}
	randomMessage := testutil.GenerateTestRandomMessage(windowAcknowledgementSize + 20)
	_, err = randomMessage.Send(clientConn)
	if err != nil {
		return
	}
	select {
	case acknowledgement := <-clientConn.Messages:
		assert.Equal(t, message.TypeAcknowledgement, acknowledgement.TypeId)
		assert.GreaterOrEqual(t, binary.BigEndian.Uint32(acknowledgement.Data[:4]), uint32(windowAcknowledgementSize))
		assert.LessOrEqual(t, clientConn.UnacknowledgedBytesReceived, uint32(windowAcknowledgementSize))
	case <-clientConn.Errors:
		t.FailNow()
	}

}

func TestSetPeerBandwidthMessageTypeHardDifferentFromPreviousSizeReceived(t *testing.T) {
	rtmpServer, clientConn := testutil.StartTestingServerWithHandshake(t)
	// client sends initial window acknowledgement size message
	windowAcknowledgementSize := 1024
	serverConn := <-rtmpServer.Connections
	testWindowAcknowledgementSizeMessage := message.NewWindowAcknowledgementSizeMessage(windowAcknowledgementSize)
	_, err := testWindowAcknowledgementSizeMessage.Send(clientConn)
	assert.Nil(t, err)
	<-serverConn.Messages
	// client sends set peer bandwidth message with different size
	bandwidthSize := 2048
	testMessage := message.NewSetPeerBandwidthMessage(bandwidthSize, message.SetPeerBandwidthLimitTypeHard)
	_, err = testMessage.Send(clientConn)
	assert.Nil(t, err)
	// server should send window acknowledgement size message
	select {
	case windowAcknowledgementSizeMessage := <-clientConn.Messages:
		assert.Equal(t, message.TypeWindowAcknowledgementSize, windowAcknowledgementSizeMessage.TypeId)
		assert.Equal(t, binary.BigEndian.AppendUint32(make([]byte, 0), uint32(bandwidthSize)), windowAcknowledgementSizeMessage.Data)
		assert.Equal(t, uint32(bandwidthSize), serverConn.PeerWindowAcknowledgementSize)
		assert.Equal(t, uint32(bandwidthSize), clientConn.WindowAcknowledgementSize)
	case <-clientConn.Errors:
		t.FailNow()
	}
	// Server sends a message with the new size
	randomTestMessage := testutil.GenerateTestRandomMessage(bandwidthSize)
	_, err = randomTestMessage.Send(serverConn)
	assert.Nil(t, err)
	// client should ack
	select {
	case acknowledgementMessage := <-serverConn.Messages:
		assert.Equal(t, message.TypeAcknowledgement, acknowledgementMessage.TypeId)
	case <-serverConn.Errors:
		t.FailNow()

	}
}

func TestUnknownCommandMessageNoAnswer(t *testing.T) {
	rtmpServer, clientConn := testutil.StartTestingServerWithHandshake(t)
	testMessage := testutil.GenerateTestUnknownCommand()
	_, err := testMessage.Send(clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	select {
	case <-serverConn.Messages:
	case <-serverConn.Errors:
		t.FailNow()
	}
	select {
	case <-clientConn.Messages:
		t.FailNow()
	case err = <-serverConn.Errors:
		assert.NotNil(t, err)
	}

}

func TestConnectMessageFlow(t *testing.T) {
	rtmpServer, clientConn := testutil.StartTestingServerWithHandshake(t)
	testMessage := testutil.GenerateTestConnectCommand()
	_, err := testMessage.Send(clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	// server receives connect message
	select {
	case <-serverConn.Messages:
		// TODO assert contents
	case <-serverConn.Errors:
		t.FailNow()
	}
	// client receives window acknowledgement size message
	select {
	case windowAcknowledgementSizeMessage := <-clientConn.Messages:
		assert.Equal(t, message.TypeWindowAcknowledgementSize, windowAcknowledgementSizeMessage.TypeId)
		assert.Equal(t, binary.BigEndian.AppendUint32(make([]byte, 0), serverConn.PeerWindowAcknowledgementSize), windowAcknowledgementSizeMessage.Data)
	case <-clientConn.Errors:
		t.FailNow()
	}
	// client receives peer bandwidth message
	select {
	case setPeerBandwidthMessage := <-clientConn.Messages:
		assert.Equal(t, message.TypeSetPeerBandwidth, setPeerBandwidthMessage.TypeId)
		assert.Equal(t, 5, len(setPeerBandwidthMessage.Data))
		assert.Equal(t, binary.BigEndian.AppendUint32(make([]byte, 0), serverConn.PeerWindowAcknowledgementSize), setPeerBandwidthMessage.Data[:4])
	case <-clientConn.Errors:
		t.FailNow()
	}
	// TODO client sends acknowledgement message
	// TODO assert server receives window acknowledgement size message
	// TODO assert client receives StreamBegin message
	// TODO assert client receives _result command message
}
