package timeoutconn

import (
	"net"
	"time"
)

const maxWaitTimeInSeconds = 10

func Read(conn net.Conn, buffer []byte) error {
	err := conn.SetReadDeadline(time.Now().Add(time.Second * maxWaitTimeInSeconds))
	if err != nil {
		return err
	}
	_, rerr := conn.Read(buffer)
	return rerr
}

func Write(conn net.Conn, buffer []byte) error {
	err := conn.SetWriteDeadline(time.Now().Add(time.Second * maxWaitTimeInSeconds))
	if err != nil {
		return err
	}
	_, werr := conn.Write(buffer)
	return werr
}
