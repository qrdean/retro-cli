package main

import (
	"log"
	"pkg/server"
	"pkg/client"
)

var id = 0

func main() {
	log.Println("hello world")

	log.Println(server.Hey())
	log.Println(client.Hey())
}
