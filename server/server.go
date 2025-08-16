package server

import (
	"fmt"
	"log"
	"miistream/handshake"
	"net"
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
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		err := handshake.Accept(conn)
		if err != nil {
			log.Println("Handshake failed", err)
			return
		}
	}
}
