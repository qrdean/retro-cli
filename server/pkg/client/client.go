package client

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
)

func Hey() string {
	return "hey from client"
}

// type Client struct {
// 	Connection net.Conn
// }

// type Model struct {
// }

func ConnectAndRead(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
		}

		msg := buf[:n]
		newReader := bytes.NewReader(msg)
		var typ uint8
		err = binary.Read(newReader, binary.BigEndian, &typ)
		if err != nil {
			log.Println("%v\n", err)
		}
		log.Println("typ is %v\n", typ)

		if typ == 6 {
			log.Println("Received shutdown signal")
			break
		}

		log.Printf("output from: %v\n", msg)

		n, err = conn.Write([]byte{5})
		if err != nil {
			log.Println(err)
		}
		log.Printf("wrote x numb of bytes %v\n", n)
	}
}
