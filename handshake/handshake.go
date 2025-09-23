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
