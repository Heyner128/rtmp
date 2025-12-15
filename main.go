package main

import (
	"rtmp/server"
)

func main() {
	server.NewServer("127.0.0.1:9999").Accept()
}
