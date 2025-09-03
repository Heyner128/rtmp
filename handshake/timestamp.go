package handshake

import (
	"encoding/binary"
	"math/rand"
	"net"
)

type Timestamp struct {
	Timestamp uint32
	Zero      uint32
	Random    [1528]byte
}

func GenerateTimestamp() Timestamp {
	timestamp := new(Timestamp)
	timestamp.Timestamp = uint32(0)
	timestamp.Zero = uint32(0)
	rands := make([]byte, 1528)
	for n := range rands {
		rands[n] = byte(rand.Intn(255))
	}
	timestamp.Random = [1528]byte(rands)
	return *timestamp
}

func ReadTimestamp(conn net.Conn) (*Timestamp, error) {
	buffer := make([]byte, 1536)
	_, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}
	timestamp := new(Timestamp)
	timestamp.Timestamp = binary.BigEndian.Uint32(buffer[0:4])
	timestamp.Zero = binary.BigEndian.Uint32(buffer[4:8])
	timestamp.Random = [1528]byte(buffer[8:])
	return timestamp, nil
}

func (timestamp Timestamp) Buffer() []byte {
	buffer := make([]byte, 0)
	buffer = binary.BigEndian.AppendUint32(buffer, timestamp.Timestamp)
	buffer = binary.BigEndian.AppendUint32(buffer, timestamp.Zero)
	buffer = append(buffer, timestamp.Random[:]...)
	return buffer
}

func (timestamp Timestamp) Send(conn net.Conn) error {
	_, err := conn.Write(timestamp.Buffer())
	return err
}
