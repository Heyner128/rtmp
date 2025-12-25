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
	testMessage := message.NewMessage(message.TypeSetChunkSize, rand.Uint32(), binary.BigEndian.AppendUint32(make([]byte, 0), newSize&0x7FFFFFFF))
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
	windowAcknowledgementSizeMessage := message.NewWindowAcknowledgementSizeMessage(int(serverConn.PeerWindowAcknowledgementSize))
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

func TestCreateStreamMessageFlow(t *testing.T) {
	rtmpServer, clientConn := testutil.StartTestingServerWithHandshake(t)
	transactionId := 2.0
	createStreamCommand := amf.NewCommand(
		amf.NewString("createStream"),
		amf.NewNumber(transactionId),
		amf.NewNull(),
	)
	testMessage := message.NewMessage(message.TypeCommandMessageAmf0, 0, createStreamCommand.Encode())
	_, err := testMessage.Send(clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	// server receives createStream message
	select {
	case createStreamMessage := <-serverConn.Messages:
		decodedCreateStreamMessage, err := amf.DecodeCommand(createStreamMessage.Data)
		assert.Nil(t, err)
		assert.Equal(t, message.TypeCommandMessageAmf0, createStreamMessage.TypeId)
		assert.Equal(t, amf.NewString("createStream"), decodedCreateStreamMessage.Parts[0])
		assert.Equal(t, amf.NewNumber(transactionId), decodedCreateStreamMessage.Parts[1])
	case <-serverConn.Errors:
		t.FailNow()
	}
	// client receives result command message
	select {
	case resultCommandMessage := <-clientConn.Messages:
		decodedResultCommandMessage, err := amf.DecodeCommand(resultCommandMessage.Data)
		assert.Nil(t, err)
		assert.Equal(t, message.TypeCommandMessageAmf0, resultCommandMessage.TypeId)
		assert.Equal(t, amf.NewString("_result"), decodedResultCommandMessage.Parts[0])
		assert.Equal(t, amf.NewNumber(transactionId), decodedResultCommandMessage.Parts[1])
		assert.Equal(t, amf.NewNull(), decodedResultCommandMessage.Parts[2])
		// the stream id should be a number
		streamId, ok := decodedResultCommandMessage.Parts[3].(amf.Number)
		assert.True(t, ok)
		assert.Greater(t, float64(streamId), float64(0))
	case <-clientConn.Errors:
		t.FailNow()
	}
}

func TestPublishMessageFlow(t *testing.T) {
	rtmpServer, clientConn := testutil.StartTestingServerWithHandshake(t)
	transactionId := 0.0
	streamName := "testStream"
	publishCommand := amf.NewCommand(
		amf.NewString("publish"),
		amf.NewNumber(transactionId),
		amf.NewNull(),
		amf.NewString(streamName),
	)
	testMessage := message.NewMessage(message.TypeCommandMessageAmf0, 0, publishCommand.Encode())
	_, err := testMessage.Send(clientConn)
	assert.Nil(t, err)
	serverConn := <-rtmpServer.Connections
	// client receives onStatus, stream begin, and result messages
	messagesReceived := 0
	var receivedOnStatus, receivedStreamBegin, receivedResult bool
	for messagesReceived < 3 {
		select {
		case msg := <-clientConn.Messages:
			messagesReceived++
			if msg.TypeId == message.TypeCommandMessageAmf0 {
				decodedCommand, err := amf.DecodeCommand(msg.Data)
				assert.Nil(t, err)
				commandName, ok := decodedCommand.Parts[0].(amf.String)
				assert.True(t, ok)
				if commandName == amf.NewString("onStatus") {
					receivedOnStatus = true
					assert.Equal(t, amf.NewNumber(0), decodedCommand.Parts[1])
					assert.Equal(t, amf.NewNull(), decodedCommand.Parts[2])
					// check the info object
					infoObj, ok := decodedCommand.Parts[3].(amf.Object)
					assert.True(t, ok)
					// find the properties in the object
					var level, code amf.ValueType
					for _, prop := range infoObj {
						if prop.Name == "level" {
							level = prop.Value
						}
						if prop.Name == "code" {
							code = prop.Value
						}
					}
					assert.Equal(t, amf.NewString("status"), level)
					assert.Equal(t, amf.NewString("NetConnection.Publish.Start"), code)
				} else if commandName == amf.NewString("_result") {
					receivedResult = true
					assert.Equal(t, amf.NewNumber(transactionId), decodedCommand.Parts[1])
					assert.Equal(t, amf.NewNull(), decodedCommand.Parts[2])
					// the stream id should be a number
					streamId, ok := decodedCommand.Parts[3].(amf.Number)
					assert.True(t, ok)
					assert.Greater(t, float64(streamId), float64(0))
				}
			} else if msg.TypeId == message.TypeUserControl {
				receivedStreamBegin = true
				assert.Equal(t, uint32(0), msg.StreamId)
				assert.Equal(t, uint16(0), binary.BigEndian.Uint16(msg.Data[:2]))
				messageStreamId := binary.BigEndian.Uint32(msg.Data[2:6])
				assert.Greater(t, messageStreamId, uint32(0))
			}
		case <-clientConn.Errors:
			t.FailNow()
		case <-serverConn.Errors:
			t.FailNow()
		}
	}
	assert.True(t, receivedOnStatus, "did not receive onStatus message")
	assert.True(t, receivedStreamBegin, "did not receive stream begin message")
	assert.True(t, receivedResult, "did not receive result message")
}
