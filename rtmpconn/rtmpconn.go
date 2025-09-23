package rtmpconn

import (
	"net"
	"time"
)

type PartialMessage struct {
	MessageLength   uint32
	MessageTypeId   uint8
	MessageStreamId uint32
	Data            []byte
}

type RtmpConn struct {
	Conn           net.Conn
	MaxChunkSize   uint32
	NetworkTimeout time.Duration
	CurrentMessage PartialMessage
	Messages       chan PartialMessage
	Errors         chan error
}

func NewRtmpConn(conn net.Conn, maxChunkSize uint32, networkTimeout time.Duration) *RtmpConn {
	return &RtmpConn{
		Conn:           conn,
		MaxChunkSize:   maxChunkSize,
		NetworkTimeout: networkTimeout,
		Messages:       make(chan PartialMessage, 1),
		Errors:         make(chan error),
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
		return 0, err
	}
	return rtmpConn.Conn.Read(buffer)
}

func (rtmpConn RtmpConn) Write(buffer []byte) (int, error) {
	err := rtmpConn.Conn.SetWriteDeadline(time.Now().Add(rtmpConn.NetworkTimeout))
	if err != nil {
		return 0, err
	}
	return rtmpConn.Conn.Write(buffer)
}

func (rtmpConn RtmpConn) Close() error {
	return rtmpConn.Conn.Close()
}
