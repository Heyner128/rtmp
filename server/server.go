package server

import (
	"fmt"
	"log"
	"net"
	"rtmp/chunk"
	"rtmp/handshake"
	"rtmp/rtmpconn"
	"time"
)

type RtmpServer struct {
	DefaultMaxChunkSize   uint32
	DefaultNetworkTimeout time.Duration
	errors                chan error
	listener              net.Listener
}

func NewRtmpServer(address string) *RtmpServer {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(fmt.Errorf("failed to start rtmp server: %s", err))
	}

	return &RtmpServer{
		DefaultMaxChunkSize:   128,
		DefaultNetworkTimeout: time.Second * 10,
		errors:                make(chan error),
		listener:              listener,
	}
}

func (server *RtmpServer) Accept() {
	log.Println("rtmp server started")
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			panic(fmt.Errorf("error closing listener: %s", err))
		}
	}(server.listener)
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			log.Println("Error accepting connection", err)
			server.errors <- err
			continue
		}
		rtmpConn := rtmpconn.NewRtmpConn(conn, server.DefaultMaxChunkSize, server.DefaultNetworkTimeout)
		go func() {
			err := handleConnection(rtmpConn)
			if err != nil {
				log.Println("Error handling connection", err)
				server.errors <- err
			}
		}()
	}
}

func handleConnection(conn *rtmpconn.RtmpConn) error {
	defer func(conn *rtmpconn.RtmpConn) {
		err := conn.Close()
		if err != nil {
			log.Println("Error closing connection", err)
		}
	}(conn)
	err := handshake.Accept(conn)
	if err != nil {
		log.Println("Handshake failed", err)
		return err
	}
	for {
		err = chunk.Accept(conn)
		if err != nil {
			log.Println("Chunk reading failed", err)
			return err
		}
	}
}
