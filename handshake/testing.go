package handshake

import (
	"net"
	"testing"
	"time"
)

func acceptTestHandshake(t *testing.T) string {
	t.Helper()
	address := "127.0.0.1:0"
	listener, _ := net.Listen("tcp", address)
	go func() {
		conn, _ := listener.Accept()
		err := Accept(conn)
		if err != nil {
			panic(err)
		}
	}()
	return listener.Addr().String()
}

func RequestTestHandshake(t *testing.T, conn net.Conn) (*Handshake, error) {
	t.Helper()
	// sends C0 and C1
	clientVersion := &Version{Version: 1}
	err := clientVersion.Send(conn)
	if err != nil {
		return nil, err
	}
	clientTimestamp := GenerateTimestamp()
	err = clientTimestamp.Send(conn)
	if err != nil {
		return nil, err
	}
	// receives S0 and S1
	serverVersion, err := ReadVersion(conn)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	serverTimestamp, err := ReadTimestamp(conn)
	if err != nil {
		return nil, err
	}
	serverTimestampReadingTimeInMs := time.Since(start).Milliseconds()
	// sends C2
	clientEcho := &Echo{
		Timestamp:  serverTimestamp.Timestamp,
		TimeStamp2: uint32(uint64(serverTimestamp.Timestamp) + uint64(serverTimestampReadingTimeInMs)),
		Random:     serverTimestamp.Random,
	}
	err = clientEcho.Send(conn)
	if err != nil {
		return nil, err
	}
	// receives S2
	serverEcho, err := ReadEcho(conn, clientTimestamp)
	if err != nil {
		return nil, err
	}
	return &Handshake{
		ClientVersion:   clientVersion,
		ServerVersion:   serverVersion,
		ClientTimestamp: &clientTimestamp,
		ServerTimestamp: serverTimestamp,
		ClientEcho:      clientEcho,
		ServerEcho:      serverEcho,
	}, nil
}
