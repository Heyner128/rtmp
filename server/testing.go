package server

import (
	"net"
	"rtmp/handshake"
	"rtmp/rtmpconn"
	"testing"
	"time"
)

func StartTestingServer(t *testing.T) *RtmpServer {
	t.Helper()
	address := "127.0.0.1:0"
	rtmpServer := NewRtmpServer(address)
	go func() {
		rtmpServer.Accept()
	}()
	return rtmpServer
}

func StartTestingServerWithHandshake(t *testing.T) (*RtmpServer, rtmpconn.RtmpConn) {
	t.Helper()
	rtmpServer := NewRtmpServer("127.0.0.1:0")
	rtmpServer.DefaultNetworkTimeout = 1 * time.Second
	go func() {
		rtmpServer.Accept()
	}()
	conn, _ := net.Dial("tcp", rtmpServer.Listener.Addr().String())
	clientConn := rtmpconn.NewRtmpConn(conn, rtmpServer.DefaultMaxChunkSize, rtmpServer.DefaultNetworkTimeout)
	_, err := handshake.RequestTestHandshake(t, clientConn)
	if err != nil {
		t.Error(err)
	}
	return rtmpServer, *clientConn
}
