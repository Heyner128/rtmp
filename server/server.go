package server

import (
	"fmt"
	"log"
	"miistream/chunk"
	"miistream/handshake"
	"miistream/rtmpconn"
	"net"
	"time"
)

func Listen(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to start rtmp server: %s", err)
	}
	defer listener.Close()
	log.Println("rtmp server started")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection", err)
		}
		rtmpConn := rtmpconn.RtmpConn{
			Conn:           conn,
			MaxChunkSize:   128,
			NetworkTimeout: time.Second * 10,
		}
		go handleConnection(&rtmpConn)
	}
}

func handleConnection(conn *rtmpconn.RtmpConn) {
	defer conn.Close()
	for {
		err := handshake.Accept(conn)
		if err != nil {
			log.Println("Handshake failed", err)
			return
		}
		err = chunk.Accept(conn)
		if err != nil {
			log.Println("Chunk failed", err)
			return
		}
	}
}
