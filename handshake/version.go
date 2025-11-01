package handshake

import (
	"net"
)

type Version struct {
	Version uint8
}

func newVersion(version uint8) *Version {
	return &Version{Version: version}
}

func ReadVersion(conn net.Conn) (*Version, error) {
	buffer := make([]byte, 1)
	_, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}
	return &Version{
		Version: buffer[0],
	}, nil
}

func (version Version) Encode() []byte {
	return []byte{version.Version}
}

func (version Version) Send(conn net.Conn) error {
	_, err := conn.Write(version.Encode())
	return err
}
