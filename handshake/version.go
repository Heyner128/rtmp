package handshake

import (
	"miistream/timeoutconn"
	"net"
)

type Version struct {
	Version uint8
}

func ReadVersion(conn net.Conn) (*Version, error) {
	buffer := make([]byte, 1)
	err := timeoutconn.Read(conn, buffer)
	if err != nil {
		return nil, err
	}
	return &Version{
		Version: buffer[0],
	}, nil
}

func (version Version) Buffer() []byte {
	return []byte{version.Version}
}

func (version Version) Send(conn net.Conn) error {
	return timeoutconn.Write(conn, version.Buffer())
}
