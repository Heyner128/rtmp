package handshake

import (
	"encoding/binary"
	"errors"
	"net"
)

type Echo struct {
	Timestamp  uint32
	TimeStamp2 uint32
	Random     [1528]byte
}

func ReadEcho(conn net.Conn, sentTimestampChunk Timestamp) (*Echo, error) {
	buffer := make([]byte, 1536)
	_, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}
	echo := new(Echo)
	echo.Timestamp = binary.BigEndian.Uint32(buffer[0:4])
	echo.TimeStamp2 = binary.BigEndian.Uint32(buffer[4:8])
	echo.Random = [1528]byte(buffer[8:])
	if echo.Timestamp != sentTimestampChunk.Timestamp || echo.Random != sentTimestampChunk.Random {
		return nil, errors.New("peer timestamp echo does not match")
	}
	return echo, nil
}

func (echo Echo) Encode() []byte {
	buffer := make([]byte, 0)
	buffer = binary.BigEndian.AppendUint32(buffer, echo.Timestamp)
	buffer = binary.BigEndian.AppendUint32(buffer, echo.TimeStamp2)
	buffer = append(buffer, echo.Random[:]...)
	return buffer
}

func (echo Echo) Send(conn net.Conn) error {

	_, err := conn.Write(echo.Encode())
	return err
}
