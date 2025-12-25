package server

import (
	"net"
	"rtmp/conn"
	"rtmp/handshake"
	"rtmp/logger"
	"rtmp/message"
	"time"
)

type Server struct {
	DefaultMaxChunkSize   uint32
	DefaultNetworkTimeout time.Duration
	Connections           chan *conn.Conn
	Listener              net.Listener
}

func NewServer(address string) *Server {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		logger.Get().Panicf("failed to start rtmp server: %s", err)
	}

	return &Server{
		DefaultMaxChunkSize:   128,
		DefaultNetworkTimeout: time.Second * 10,
		Listener:              listener,
		Connections:           make(chan *conn.Conn),
	}
}

func (server *Server) Accept() {
	logger.Get().Infof("rtmp server started")
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			logger.Get().Panicf("error closing listener: %s", err)
		}
	}(server.Listener)
	for {
		netConnection, _ := server.Listener.Accept()
		connection, _ := conn.NewConn(netConnection, server.DefaultMaxChunkSize, server.DefaultNetworkTimeout)
		select {
		case server.Connections <- connection:
		default:
		}
		go func() {
			err := handleConnection(connection)
			if err != nil {
				logger.Get().Error("Error handling connection ", err)
				connection.Errors <- err
			}
		}()
	}
}

func handleConnection(connection *conn.Conn) error {
	defer func(conn *conn.Conn) {
		err := conn.Close()
		if err != nil {
			logger.Get().Error("Error closing connection ", err)
		}
	}(connection)
	err := handshake.Accept(connection)
	if err != nil {
		logger.Get().Error("Handshake failed ", err)
		return err
	}
	for {
		_, err = message.Accept(connection)
		if err != nil {
			logger.Get().Error("Chunk reading failed ", err)
			return err
		}
		if connection.CurrentMessage == nil {
			err = connection.Close()
			if err != nil {
				return err
			}
		}
	}
}
