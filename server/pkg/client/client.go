package client

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"pkg/shared"
	"strings"
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

	go func() {
		for {
			msg, typ := getText()
			switch typ {
			case 0:
				continue
			case shared.AddStickyType:
				println(msg)
			case shared.VoteStickyType:
				println(msg)
			case shared.QuitType:
				println(msg)
			}
		}
	}()

	for {
		buf := make([]byte, 1024*4)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
		}

		weBreak, err := handleMessage(buf, n)
		if err != nil {
			log.Printf("error: %v\n", err)
		}

		if weBreak {
			break
		}

		n, err = conn.Write([]byte{1, 5})
		if err != nil {
			log.Println(err)
		}
		// log.Printf("wrote x numb of bytes %v\n", n)
	}
}

func getText() (string, byte) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter text:")
	text, _ := reader.ReadString('\n')
	if strings.Contains(strings.ToLower(text), "add") {
		_, after, _ := strings.Cut(text, "add")
		return strings.TrimSpace(after), shared.AddStickyType
	}

	if strings.Contains(strings.ToLower(text), "vote") {
		fmt.Println("vote")
		_, after, _ := strings.Cut(text, "vote")
		return strings.TrimSpace(after), shared.VoteStickyType
	}

	if strings.Contains(strings.ToLower(text), "quit") {
		fmt.Println("quit")
		_, after, _ := strings.Cut(text, "quit")
		return strings.TrimSpace(after), shared.QuitType
	}
	var nothing byte = 0
	return "", nothing
}

func handleMessage(buf []byte, n int) (bool, error) {
	msg := buf[:n]
	newReader := bytes.NewReader(msg)
	var version byte
	err := binary.Read(newReader, binary.BigEndian, &version)
	if err != nil {
		log.Printf("%v\n", err)
	}
	// log.Printf("version is %v\n", version)
	var typ byte
	err = binary.Read(newReader, binary.BigEndian, &typ)
	if err != nil {
		log.Printf("%v\n", err)
	}
	// log.Printf("typ is %v\n", typ)

	switch typ {
	case 6:
		log.Println("Received shutdown signal")
		return true, nil
	case shared.PointerType:
		var pointerBytes shared.PointerBytes = msg
		_, err := pointerBytes.ReadFrom(newReader)
		if err != nil {
			log.Printf("%v\n", err)
			return true, err
		}

		// log.Printf("ns len %v\n", ns)
		// log.Printf("output to pointer bytes %v\n", pointerBytes)
		pointer := pointerBytes.UnmarshalPointer()
		log.Printf("Id: %v\n",
			pointer.PointerId,
		)
	case shared.TopicType:
		var topicBytes shared.TopicBytes = msg
		_, err := topicBytes.ReadFrom(newReader)
		if err != nil {
			log.Printf("%v\n", err)
			return true, err
		}

		// log.Printf("ns len %v\n", ns)
		// log.Printf("output to topic bytes %v\n", topicBytes)
		topic := topicBytes.UnmarshalTopic()
		log.Printf("Id: %v, topic message: %v\n",
			topic.Id,
			string(topic.Header[:]),
		)
	case shared.StickyType:
		var stickyBytes shared.StickyBytes = msg
		_, err := stickyBytes.ReadFrom(newReader)
		if err != nil {
			log.Printf("%v\n", err)
			return true, err
		}

		// log.Printf("ns len %v\n", ns)
		// log.Printf("output to sticky bytes %v\n", stickyBytes)
		sticky := stickyBytes.UnmarshalBinaryStick()
		log.Printf("Id: %v, topic id: %v, poster id:%v, votes: %v, sticky message: %v\n",
			sticky.Id,
			sticky.TopicId,
			sticky.Votes,
			sticky.PosterId,
			string(sticky.StickyMessage[:]),
		)
	default:
		log.Printf("not a valid action %v\n", typ)
	}

	// log.Printf("output from: %v\n", msg)
	return false, nil
}
