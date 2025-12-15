package server

import (
	"fmt"
	"log"
	"net"
	"rtmp/conn"
	"rtmp/handshake"
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
		panic(fmt.Errorf("failed to start rtmp server: %s", err))
	}

	return &Server{
		DefaultMaxChunkSize:   128,
		DefaultNetworkTimeout: time.Second * 10,
		Listener:              listener,
		Connections:           make(chan *conn.Conn, 1),
	}
}

func (server *Server) Accept() {
	log.Println("rtmp server started")
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			panic(fmt.Errorf("error closing listener: %s", err))
		}
	}(server.Listener)
	for {
		netConnection, _ := server.Listener.Accept()
		connection := conn.NewConn(netConnection, server.DefaultMaxChunkSize, server.DefaultNetworkTimeout)
		server.Connections <- connection
		go func() {
			err := handleConnection(connection)
			if err != nil {
				log.Println("Error handling connection", err)
				connection.Errors <- err
			}
		}()
	}
}

func handleConnection(connection *conn.Conn) error {
	defer func(conn *conn.Conn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing connection", err)
		}
	}(connection)
	err := handshake.Accept(connection)
	if err != nil {
		log.Println("Handshake failed", err)
		return err
	}
	for {
		_, err = message.Accept(connection)
		if err != nil {
			log.Println("Chunk reading failed", err)
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
