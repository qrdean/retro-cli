package client

import (
	"bufio"
	// "bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"pkg/shared"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	// tea "github.com/charmbracelet/bubbletea"
)

func Hey() string {
	return "hey from client"
}

// type Client struct {
// 	Connection net.Conn
// }

// type Model struct {
// }

func RefactorConnectAndRead(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		weBreak, err := handleMessage(reader)
		if err != nil {
			log.Printf("error: %v\n", err)
		}

		if weBreak {
			break
		}
	}
}

type ErrMsg struct{ err error }
type Break bool
type Init bool
type TopicsDone bool
type TopicLength uint32

func initMessageHandler(conn net.Conn) tea.Cmd {
	return func() tea.Msg {
		newReader := bufio.NewReader(conn)
		return refactorHandleMessage(newReader)
	}
}

// func refactorHandleMessage(conn net.Conn) interface{} {
func refactorHandleMessage(newReader io.Reader) interface{} {
	// return func() tea.Msg {
	// newReader := bufio.NewReader(conn)
	// log.Println("handleing")
	// msg := buf[:n]
	// newReader := bytes.NewReader(msg)
	var version byte
	err := binary.Read(newReader, binary.BigEndian, &version)
	if err != nil {
		// log.Printf("%v\n", err)
		return ErrMsg{err}
	}
	// log.Printf("version is %v\n", version)
	var typ byte
	err = binary.Read(newReader, binary.BigEndian, &typ)
	if err != nil {
		// log.Printf("%v\n", err)
		return ErrMsg{err}
	}
	// log.Printf("typ is %v\n", typ)

	switch typ {
	case 0:
		// log.Println("connection closed")
		return Break(true)
	case 6:
		// log.Println("Received shutdown signal")
		return Break(true)

	case 43:
		return TopicsDone(true)

	case 49:
		var packet shared.Packet
		_, err := packet.Byte.ReadFrom(newReader)
		if err != nil {
			log.Println(err)
			return ErrMsg{err}
		}
		length := shared.UnmarshalPointerTopicLength(packet.Byte)

		log.Printf("topic length: %v", length)

		return TopicLength(length)

	case shared.PointerType:
		// var pointerBytes shared.PointerBytes = msg
		var pointerBytes shared.PointerBytes
		_, err := pointerBytes.ReadFrom(newReader)
		if err != nil {
			// log.Printf("%v\n", err)
			return ErrMsg{err}
		}

		// log.Printf("ns len %v\n", ns)
		// log.Printf("output to pointer bytes %v\n", pointerBytes)
		pointer := pointerBytes.UnmarshalPointer()
		// log.Printf("Id: %v\n", pointer.PointerId)
		return pointer
	case shared.TopicType:
		// var topicBytes shared.TopicBytes = msg
		var topicBytes shared.TopicBytes
		_, err := topicBytes.ReadFrom(newReader)
		if err != nil {
			// log.Printf("%v\n", err)
			return ErrMsg{err}
		}

		// log.Printf("ns len %v\n", ns)
		// log.Printf("output to topic bytes %v\n", topicBytes)
		topic := topicBytes.UnmarshalTopic()
		// log.Printf("Id: %v, topic message: %v\n", topic.Id, string(topic.Header[:]))
		return topic
	case shared.StickyType:
		// var stickyBytes shared.StickyBytes = msg
		var stickyBytes shared.StickyBytes
		_, err := stickyBytes.ReadFrom(newReader)
		if err != nil {
			// log.Printf("%v\n", err)
			return ErrMsg{err}
		}

		// log.Printf("ns len %v\n", ns)
		// log.Printf("output to sticky bytes %v\n", stickyBytes)
		sticky := stickyBytes.UnmarshalBinaryStick()
		// log.Printf("Id: %v, topic id: %v, poster id:%v, votes: %v, sticky message: %v\n",
		// 	sticky.Id,
		// 	sticky.TopicId,
		// 	sticky.PosterId,
		// 	sticky.Votes,
		// 	string(sticky.StickyMessage[:]),
		// )
		return sticky
	default:
		log.Printf("not a valid action %v\n", typ)
	}

	// log.Printf("output from: %v\n", msg)
	return nil //Init(true)
	// }
}

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
				sticky, err := shared.NewAddSticky(1, 1, msg)
				if err != nil {
					log.Println(err)
					continue
				}
				var stickyBytes shared.AddStickyBytes
				stickyBytes = sticky.MarshalBinary()
				n, err := stickyBytes.WriteTo(conn)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Printf("successfully wrote %v bytes %v\n", n, stickyBytes[:n-8])
			case shared.VoteStickyType:
				println(msg)
				int, err := strconv.Atoi(msg)
				if err != nil {
					log.Println(err)
					continue
				}
				voteSticky := shared.NewVoteSticky(uint32(int))
				var voteBytes shared.VoteBytes = voteSticky.MarshalBinary()
				n, err := voteBytes.WriteTo(conn)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Printf("wrote %v bytes %v\n", n, voteBytes[:])
			case shared.QuitType:
				println(msg)
				quit := shared.NewQuit(1)
				var quitBytes shared.QuitBytes = quit.MarshalBinary()
				n, err := quitBytes.WriteTo(conn)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println("wrote %v bytes %v\n", n, quitBytes[:])
			case shared.PointToType:
				println(msg)
				int, err := strconv.Atoi(msg)
				if err != nil {
					log.Println(err)
					continue
				}
				pointTo := shared.NewPointToSticky(uint32(int))
				var pointToBytes shared.PointToStickyBytes = pointTo.MarshalBinary()
				n, err := pointToBytes.WriteTo(conn)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Printf("wrote %v bytes %v\n", n, pointToBytes[:])
			}
		}
	}()

	// buf := make([]byte, 1024*4)
	reader := bufio.NewReader(conn)
	for {
		weBreak, err := handleMessage(reader)
		if err != nil {
			log.Printf("error: %v\n", err)
		}

		if weBreak {
			break
		}
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

	if strings.Contains(strings.ToLower(text), "point") {
		fmt.Println("point")
		_, after, _ := strings.Cut(text, "point")
		return strings.TrimSpace(after), shared.PointToType
	}

	var nothing byte = 0
	return "", nothing
}

// func handleMessage(buf []byte, n int) (bool, error) {
func handleMessage(newReader io.Reader) (bool, error) {
	log.Println("handleing")
	// msg := buf[:n]
	// newReader := bytes.NewReader(msg)
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
	case 0:
		log.Println("connection closed")
		return true, nil
	case 6:
		log.Println("Received shutdown signal")
		return true, nil
	case shared.PointerType:
		// var pointerBytes shared.PointerBytes = msg
		var pointerBytes shared.PointerBytes
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
		// var topicBytes shared.TopicBytes = msg
		var topicBytes shared.TopicBytes
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
		// var stickyBytes shared.StickyBytes = msg
		var stickyBytes shared.StickyBytes
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
			sticky.PosterId,
			sticky.Votes,
			string(sticky.StickyMessage[:]),
		)
	default:
		log.Printf("not a valid action %v\n", typ)
	}

	// log.Printf("output from: %v\n", msg)
	return false, nil
}
