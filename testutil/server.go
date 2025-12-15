package testutil

import (
	"net"
	"rtmp/conn"
	"rtmp/message"
	"rtmp/server"
	"testing"
	"time"
)

func StartTestingServer(t *testing.T) *server.Server {
	t.Helper()
	address := "127.0.0.1:0"
	rtmpServer := server.NewServer(address)
	rtmpServer.DefaultNetworkTimeout = 3 * time.Second
	go func() {
		rtmpServer.Accept()
	}()
	return rtmpServer
}

func StartTestingServerWithHandshake(t *testing.T) (*server.Server, conn.Conn) {
	t.Helper()
	rtmpServer := StartTestingServer(t)
	netConnection, _ := net.Dial("tcp", rtmpServer.Listener.Addr().String())
	err := netConnection.SetDeadline(time.Now().Add(3 * time.Second))
	clientConn := conn.NewConn(netConnection, rtmpServer.DefaultMaxChunkSize, rtmpServer.DefaultNetworkTimeout)
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
