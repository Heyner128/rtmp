package message_test

import (
	"encoding/binary"
	"math/rand"
	"rtmp/amf"
	"rtmp/message"
	"rtmp/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageReceived(t *testing.T) {
	rtmpServer, clientConn := testutil.StartTestingServerWithHandshake(t)
	testMessage := testutil.GenerateTestRandomMessage(256)
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
	setSizeMessage := message.NewMessage(uint8(1), rand.Uint32(), binary.BigEndian.AppendUint32(make([]byte, 0), newSize&0x7FFFFFFF))
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
	testMessage := message.NewMessage(message.TypeSetChunkSize, rand.Uint32(), binary.BigEndian.AppendUint32(make([]byte, 0), newSize<<1))
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
	testMessage := message.NewMessage(message.TypeAbortMessage, rand.Uint32(), binary.BigEndian.AppendUint32(make([]byte, 0), uint32(2)))
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
	windowAcknowledgementSizeMessage := message.NewWindowAcknowledgementSizeMessage(rand.Uint32(), windowAcknowledgementSize)
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
	messageStreamId := rand.Uint32()
	// client sends initial window acknowledgement size message
	windowAcknowledgementSize := 1024
	serverConn := <-rtmpServer.Connections
	testWindowAcknowledgementSizeMessage := message.NewWindowAcknowledgementSizeMessage(messageStreamId, windowAcknowledgementSize)
	_, err := testWindowAcknowledgementSizeMessage.Send(clientConn)
	assert.Nil(t, err)
	<-serverConn.Messages
	// client sends set peer bandwidth message with different size
	bandwidthSize := 2048
	testMessage := message.NewSetPeerBandwidthMessage(messageStreamId, bandwidthSize, message.SetPeerBandwidthLimitTypeHard)
	_, err = testMessage.Send(clientConn)
	assert.Nil(t, err)
	<-serverConn.Messages
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
	case connectMessage := <-serverConn.Messages:
		decodedConnectMessage, err := amf.DecodeCommand(connectMessage.Data)
		assert.Nil(t, err)
		assert.Equal(t, message.TypeCommandMessageAmf0, connectMessage.TypeId)
		assert.Equal(t, amf.NewString("connect"), decodedConnectMessage.Parts[0])
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
	// client sends window acknowledgement size message
	windowAcknowledgementSizeMessage := message.NewWindowAcknowledgementSizeMessage(rand.Uint32(), int(serverConn.PeerWindowAcknowledgementSize))
	_, err = windowAcknowledgementSizeMessage.Send(clientConn)
	assert.Nil(t, err)
	select {
	case receivedMessage := <-serverConn.Messages:
		assert.Equal(t, message.TypeWindowAcknowledgementSize, receivedMessage.TypeId)
		assert.Equal(t, binary.BigEndian.AppendUint32(make([]byte, 0), serverConn.PeerWindowAcknowledgementSize), receivedMessage.Data)
	case <-serverConn.Errors:
		t.FailNow()
	}
	// server sends stream begin message
	select {
	case streamBeginMessage := <-clientConn.Messages:
		assert.Equal(t, message.TypeUserControl, streamBeginMessage.TypeId)
		assert.Equal(t, uint32(0), streamBeginMessage.StreamId)
		assert.Equal(t, uint16(0), binary.BigEndian.Uint16(streamBeginMessage.Data[:2]))
		assert.GreaterOrEqual(t, binary.BigEndian.Uint16(streamBeginMessage.Data[2:6]), uint16(0))
	case <-clientConn.Errors:
		t.FailNow()
	case <-serverConn.Errors:
		t.FailNow()
	}
	//server sends result command message
	select {
	case resultCommandMessage := <-clientConn.Messages:
		decodedResultCommandMessage, err := amf.DecodeCommand(resultCommandMessage.Data)
		assert.Nil(t, err)
		assert.Equal(t, amf.NewString("_result"), decodedResultCommandMessage.Parts[0])
		assert.Equal(t, amf.NewNumber(1), decodedResultCommandMessage.Parts[1])
	case <-clientConn.Errors:
		t.FailNow()
	}
}
