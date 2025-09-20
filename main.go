package main

import (
	"rtmp/server"
)

func main() {
	server.NewRtmpServer("127.0.0.1:9999").Accept()
}
