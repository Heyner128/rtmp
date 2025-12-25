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
	// buffers the channels to avoid blocking
	rtmpServer.Connections = make(chan *conn.Conn, 100)
	go func() {
		rtmpServer.Accept()
	}()
	return rtmpServer
}

func StartTestingServerWithHandshake(t *testing.T) (*server.Server, *conn.Conn) {
	t.Helper()
	rtmpServer := StartTestingServer(t)
	// buffers the channels to avoid blocking
	rtmpServer.Connections = make(chan *conn.Conn, 100)
	netConnection, _ := net.Dial("tcp", rtmpServer.Listener.Addr().String())
	err := netConnection.SetDeadline(time.Now().Add(3 * time.Second))
	clientConn, _ := conn.NewConn(netConnection, rtmpServer.DefaultMaxChunkSize, rtmpServer.DefaultNetworkTimeout)
	// buffers the channels to avoid blocking
	clientConn.Messages = make(chan *conn.Message, 100)
	clientConn.Errors = make(chan error, 1)
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
	return rtmpServer, clientConn
}
