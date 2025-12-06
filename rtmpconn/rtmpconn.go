package rtmpconn

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

type RtmpConn struct {
	Conn                          net.Conn
	MaxChunkSize                  uint32
	NetworkTimeout                time.Duration
	CurrentMessage                *Message
	Messages                      chan *Message
	WindowAcknowledgementSize     uint32
	SendWindowAcknowledgementSize uint32
	Errors                        chan error
}

func NewRtmpConn(conn net.Conn, maxChunkSize uint32, networkTimeout time.Duration) *RtmpConn {
	return &RtmpConn{
		Conn:                          conn,
		MaxChunkSize:                  maxChunkSize,
		NetworkTimeout:                networkTimeout,
		CurrentMessage:                &Message{},
		SendWindowAcknowledgementSize: 256 * 1024,
		Messages:                      make(chan *Message, 1),
		Errors:                        make(chan error),
	}
}

func (rtmpConn RtmpConn) LocalAddr() net.Addr {
	return rtmpConn.Conn.LocalAddr()
}

func (rtmpConn RtmpConn) RemoteAddr() net.Addr {
	return rtmpConn.Conn.RemoteAddr()
}

func (rtmpConn RtmpConn) SetDeadline(t time.Time) error {
	return rtmpConn.Conn.SetDeadline(t)
}

func (rtmpConn RtmpConn) SetReadDeadline(t time.Time) error {
	return rtmpConn.Conn.SetReadDeadline(t)
}

func (rtmpConn RtmpConn) SetWriteDeadline(t time.Time) error {
	return rtmpConn.Conn.SetWriteDeadline(t)
}

func (rtmpConn RtmpConn) Read(buffer []byte) (int, error) {
	err := rtmpConn.Conn.SetReadDeadline(time.Now().Add(rtmpConn.NetworkTimeout))
	if err != nil {
		rtmpConn.Errors <- err
		return 0, err
	}
	n, err := rtmpConn.Conn.Read(buffer)
	if err != nil {
		rtmpConn.Errors <- err
		return 0, err
	}
	return n, err
}

func (rtmpConn RtmpConn) Write(buffer []byte) (int, error) {
	err := rtmpConn.Conn.SetWriteDeadline(time.Now().Add(rtmpConn.NetworkTimeout))
	if err != nil {
		rtmpConn.Errors <- err
		return 0, err
	}
	n, err := rtmpConn.Conn.Write(buffer)
	if err != nil {
		rtmpConn.Errors <- err
		return 0, err
	}
	return n, err
}

func (rtmpConn RtmpConn) Close() error {
	return rtmpConn.Conn.Close()
}
