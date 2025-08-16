package main

import (
	"log"
	"miistream/server"
)

func main() {
	err := server.Listen("127.0.0.1:9999")
	if err != nil {
		log.Fatal(err)
	}
}
