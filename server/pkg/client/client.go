package client

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"pkg/shared"
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
		var version byte
		err = binary.Read(newReader, binary.BigEndian, &version)
		if err != nil {
			log.Printf("%v\n", err)
		}
		log.Printf("version is %v\n", version)
		var typ byte
		err = binary.Read(newReader, binary.BigEndian, &typ)
		if err != nil {
			log.Printf("%v\n", err)
		}
		log.Printf("typ is %v\n", typ)

		if typ == 6 {
			log.Println("Received shutdown signal")
			break
		}

		if typ == shared.PointerType {
			var pointerBytes shared.PointerBytes = msg
			ns, err := pointerBytes.ReadFrom(newReader)
			if err != nil {
				log.Printf("%v\n", err)
			}

			log.Printf("ns len %v\n", ns)
			log.Printf("output to pointer bytes %v\n", pointerBytes)
			pointer := shared.UnmarshalPointer(pointerBytes)
			log.Printf("Id: %v\n",
				pointer.PointerId,
			)
		}

		if typ == shared.TopicType {
			var topicBytes shared.TopicBytes = msg
			ns, err := topicBytes.ReadFrom(newReader)
			if err != nil {
				log.Printf("%v\n", err)
			}

			log.Printf("ns len %v\n", ns)
			log.Printf("output to topic bytes %v\n", topicBytes)
			topic := shared.UnmarshalTopic(topicBytes)
			log.Printf("Id: %v, topic message: %v\n",
				topic.Id,
				string(topic.Header[:]),
			)
		}

		if typ == shared.StickyType {
			var stickyBytes shared.StickyBytes = msg
			ns, err := stickyBytes.ReadFrom(newReader)
			if err != nil {
				log.Printf("%v\n", err)
			}

			log.Printf("ns len %v\n", ns)
			log.Printf("output to sticky bytes %v\n", stickyBytes)
			sticky := shared.UnmarshalBinaryStick(stickyBytes)
			log.Printf("Id: %v, topic id: %v, poster id:%v, votes: %v, sticky message: %v\n",
				sticky.Id,
				sticky.TopicId,
				sticky.Votes,
				sticky.PosterId,
				string(sticky.StickyMessage[:]),
			)
		}

		log.Printf("output from: %v\n", msg)

		n, err = conn.Write([]byte{1, 5})
		if err != nil {
			log.Println(err)
		}
		log.Printf("wrote x numb of bytes %v\n", n)
	}
}
