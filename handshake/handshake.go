package handshake

import (
	"net"
	"rtmp/logger"
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
	logger.Get().Debug("Version sent")
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
	logger.Get().Debug("ACK sent")
	_, err = ReadEcho(conn, serverTimestamp)
	if err != nil {
		return err
	}
	logger.Get().Debug("Handshake successful")
	return nil
}
