package testutil

import (
	"net"
	"rtmp/handshake"
	"testing"
	"time"
)

func AcceptTestHandshake(t *testing.T) string {
	t.Helper()
	address := "127.0.0.1:0"
	listener, _ := net.Listen("tcp", address)
	go func() {
		conn, _ := listener.Accept()
		err := handshake.Accept(conn)
		if err != nil {
			panic(err)
		}
	}()
	return listener.Addr().String()
}

func RequestTestHandshake(t *testing.T, conn net.Conn) (*handshake.Handshake, error) {
	t.Helper()
	err := conn.SetDeadline(time.Now().Add(3 * time.Second))
	if err != nil {
		return nil, err
	}
	// sends C0 and C1
	clientVersion := &handshake.Version{Version: 1}
	err = clientVersion.Send(conn)
	if err != nil {
		return nil, err
	}
	clientTimestamp := handshake.GenerateTimestamp()
	err = clientTimestamp.Send(conn)
	if err != nil {
		return nil, err
	}
	// receives S0 and S1
	serverVersion, err := handshake.ReadVersion(conn)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	serverTimestamp, err := handshake.ReadTimestamp(conn)
	if err != nil {
		return nil, err
	}
	serverTimestampReadingTimeInMs := time.Since(start).Milliseconds()
	// sends C2
	clientEcho := &handshake.Echo{
		Timestamp:  serverTimestamp.Timestamp,
		TimeStamp2: uint32(uint64(serverTimestamp.Timestamp) + uint64(serverTimestampReadingTimeInMs)),
		Random:     serverTimestamp.Random,
	}
	err = clientEcho.Send(conn)
	if err != nil {
		return nil, err
	}
	// receives S2
	serverEcho, err := handshake.ReadEcho(conn, clientTimestamp)
	if err != nil {
		return nil, err
	}
	return &handshake.Handshake{
		ClientVersion:   clientVersion,
		ServerVersion:   serverVersion,
		ClientTimestamp: &clientTimestamp,
		ServerTimestamp: serverTimestamp,
		ClientEcho:      clientEcho,
		ServerEcho:      serverEcho,
	}, nil
}
