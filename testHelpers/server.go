package testHelpers

import (
	"net"
	"rtmp/message"
	"rtmp/rtmpconn"
	"rtmp/server"
	"testing"
	"time"
)

func StartTestingServer(t *testing.T) *server.RtmpServer {
	t.Helper()
	address := "127.0.0.1:0"
	rtmpServer := server.NewRtmpServer(address)
	rtmpServer.DefaultNetworkTimeout = 3 * time.Second
	go func() {
		rtmpServer.Accept()
	}()
	return rtmpServer
}

func StartTestingServerWithHandshake(t *testing.T) (*server.RtmpServer, rtmpconn.RtmpConn) {
	t.Helper()
	rtmpServer := StartTestingServer(t)
	conn, _ := net.Dial("tcp", rtmpServer.Listener.Addr().String())
	err := conn.SetDeadline(time.Now().Add(3 * time.Second))
	clientConn := rtmpconn.NewRtmpConn(conn, rtmpServer.DefaultMaxChunkSize, rtmpServer.DefaultNetworkTimeout)
	if err != nil {
		t.Error(err)
	}
	_, err = RequestTestHandshake(t, clientConn)
	go func() {
		for {
			_, err = message.Accept(clientConn)
			if err != nil {
				t.Error(err)
			}
		}
	}()
	if err != nil {
		t.Error(err)
	}
	return rtmpServer, *clientConn
}
