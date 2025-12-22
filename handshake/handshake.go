package handshake

import (
	"errors"
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
		TimeStamp2: uint32(time.Since(start).Milliseconds()),
		Random:     clientTimestamp.Random,
	}
	err = serverEcho.Send(conn)
	if err != nil {
		return err
	}
	log.Println("ACK sent")
	echo, err := ReadEcho(conn, serverTimestamp)
	if echo != nil && (echo.Timestamp != serverTimestamp.Timestamp || echo.Random != serverTimestamp.Random) {
		return errors.New("peer timestamp echo does not match")
	}
	if err != nil {
		return err
	}
	log.Println("Handshake successful")
	return nil
}
