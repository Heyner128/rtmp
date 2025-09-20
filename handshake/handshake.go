package handshake

import (
	"log"
	"net"
	"time"
)

type Handshake struct {
	ClientVersion   *Version
	ServerVersion   *Version
	ClientTimestamp *Timestamp
	ServerTimestamp *Timestamp
	ClientEcho      *Echo
	ServerEcho      *Echo
}

func NewHandshake(clientVersion *Version, serverVersion *Version, clientTimestamp *Timestamp, serverTimestamp *Timestamp, clientEcho *Echo, serverEcho *Echo) *Handshake {
	return &Handshake{
		ClientVersion:   clientVersion,
		ServerVersion:   serverVersion,
		ClientTimestamp: clientTimestamp,
		ServerTimestamp: serverTimestamp,
		ClientEcho:      clientEcho,
		ServerEcho:      serverEcho,
	}
}

func Accept(conn net.Conn) error {
	// receives C0 and C1
	_, err := ReadVersion(conn)
	if err != nil {
		return err
	}
	start := time.Now()
	clientTimestamp, err := ReadTimestamp(conn)
	clientTimestampReadingTimeInMs := time.Since(start).Milliseconds()
	if err != nil {
		return err
	}
	// sends S0 and S1
	serverVersion := Version{Version: 3}
	err = serverVersion.Send(conn)
	if err != nil {
		return err
	}
	serverTimestamp := GenerateTimestamp()
	err = serverTimestamp.Send(conn)
	if err != nil {
		return err
	}
	log.Println("Version sent")
	//sends S2
	serverEcho := Echo{
		Timestamp:  clientTimestamp.Timestamp,
		TimeStamp2: uint32(uint64(clientTimestamp.Timestamp) + uint64(clientTimestampReadingTimeInMs)),
		Random:     clientTimestamp.Random,
	}
	err = serverEcho.Send(conn)
	if err != nil {
		return err
	}
	log.Println("ACK sent")
	_, err = ReadEcho(conn, serverTimestamp)
	if err != nil {
		return err
	}
	log.Println("Handshake successful")
	return nil
}

func Request(conn net.Conn) (*Handshake, error) {
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
