package conn

import (
	"net"
	"time"
)

type Message struct {
	Length   uint32
	TypeId   uint8
	StreamId uint32
	Data     []byte
}

func (message *Message) DataSize() uint32 {
	return uint32(len(message.Data))
}

type Conn struct {
	Conn                          net.Conn
	MaxChunkSize                  uint32
	NetworkTimeout                time.Duration
	CurrentMessage                *Message
	Messages                      chan *Message
	UnacknowledgedBytesReceived   uint32
	UnacknowledgedBytesSent       uint32
	WindowAcknowledgementSize     uint32
	PeerWindowAcknowledgementSize uint32
	Errors                        chan error
}

func NewConn(conn net.Conn, maxChunkSize uint32, networkTimeout time.Duration) (*Conn, error) {
	newConn := &Conn{
		Conn:                          conn,
		MaxChunkSize:                  maxChunkSize,
		NetworkTimeout:                networkTimeout,
		CurrentMessage:                &Message{},
		PeerWindowAcknowledgementSize: 2 * 1024,
		Messages:                      make(chan *Message, 1),
		Errors:                        make(chan error),
	}
	if newConn.Conn != nil {
		err := newConn.Conn.SetReadDeadline(time.Now().Add(newConn.NetworkTimeout))
		if err != nil {
			newConn.Errors <- err
			return nil, err
		}
		err = newConn.Conn.SetWriteDeadline(time.Now().Add(newConn.NetworkTimeout))
		if err != nil {
			newConn.Errors <- err
			return nil, err
		}
	}

	return newConn, nil
}

func (rtmpConn *Conn) LocalAddr() net.Addr {
	return rtmpConn.Conn.LocalAddr()
}

func (rtmpConn *Conn) RemoteAddr() net.Addr {
	return rtmpConn.Conn.RemoteAddr()
}

func (rtmpConn *Conn) SetDeadline(t time.Time) error {
	return rtmpConn.Conn.SetDeadline(t)
}

func (rtmpConn *Conn) SetReadDeadline(t time.Time) error {
	return rtmpConn.Conn.SetReadDeadline(t)
}

func (rtmpConn *Conn) SetWriteDeadline(t time.Time) error {
	return rtmpConn.Conn.SetWriteDeadline(t)
}

func (rtmpConn *Conn) Read(buffer []byte) (int, error) {
	n, err := rtmpConn.Conn.Read(buffer)
	if err != nil {
		rtmpConn.Errors <- err
		return 0, err
	}
	return n, err
}

func (rtmpConn *Conn) Write(buffer []byte) (int, error) {
	n, err := rtmpConn.Conn.Write(buffer)
	if err != nil {
		rtmpConn.Errors <- err
		return 0, err
	}
	return n, err
}

func (rtmpConn *Conn) Close() error {
	return rtmpConn.Conn.Close()
}
