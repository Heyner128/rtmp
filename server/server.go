package server

import (
	"fmt"
	"log"
	"net"
	"rtmp/handshake"
	"rtmp/message"
	"rtmp/rtmpconn"
	"time"
)

type RtmpServer struct {
	DefaultMaxChunkSize   uint32
	DefaultNetworkTimeout time.Duration
	Connections           chan *rtmpconn.RtmpConn
	Listener              net.Listener
}

func NewRtmpServer(address string) *RtmpServer {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(fmt.Errorf("failed to start rtmp server: %s", err))
	}

	return &RtmpServer{
		DefaultMaxChunkSize:   128,
		DefaultNetworkTimeout: time.Second * 10,
		Listener:              listener,
		Connections:           make(chan *rtmpconn.RtmpConn, 1),
	}
}

func (server *RtmpServer) Accept() {
	log.Println("rtmp server started")
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			panic(fmt.Errorf("error closing listener: %s", err))
		}
	}(server.Listener)
	for {
		conn, _ := server.Listener.Accept()
		rtmpConn := rtmpconn.NewRtmpConn(conn, server.DefaultMaxChunkSize, server.DefaultNetworkTimeout)
		server.Connections <- rtmpConn
		go func() {
			err := handleConnection(rtmpConn)
			if err != nil {
				log.Println("Error handling connection", err)
				rtmpConn.Errors <- err
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
		_, err = message.Accept(conn)
		if err != nil {
			log.Println("Chunk reading failed", err)
			return err
		}
		if conn.CurrentMessage == nil {
			err = conn.Close()
			if err != nil {
				return err
			}
		}
	}
}
