package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"pkg/shared"
	"sync"
)

func Hey() string {
	return "hey"
}

type Connection struct {
	Id   int
	Conn net.Conn
}

type TCP struct {
	Connections []Connection
	Listener    net.Listener
	mutex       sync.RWMutex
	Incoming    chan []byte
}

func NewTcpServer(addr string) TCP {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	return TCP{
		Connections: []Connection{},
		Listener:    listener,
		mutex:       sync.RWMutex{},
		Incoming:    make(chan []byte),
	}
}

func NewConnection(id int, conn net.Conn) Connection {
	id = id + 1
	return Connection{
		Id:   id,
		Conn: conn,
	}
}

func (t *TCP) acceptConnection(ctx context.Context) {
	id := 0
	for {
		conn, err := t.Listener.Accept()
		// select {
		// case <-ctx.Done():
		// 	log.Println("shutting down tcp server gracefully")
		// 	break
		// }
		// log.Println(conn)
		if err != nil {
			log.Printf("error in accpet %v\n", err)
			continue
		}

		go func(conn net.Conn) {
			// Probably send a msg to the connection welcoming and attaching
			// an id
			// Send msg
			t.mutex.Lock()
			newConnection := NewConnection(id, conn)

			// need to make a welcome message
			n, err := newConnection.Conn.Write([]byte{1, 69})
			if err != nil {
				log.Printf("error occurred when trying to write to new conneciton: %v\n", err)
				return
			}

			log.Println(n)

			t.Connections = append(t.Connections, newConnection)
			t.mutex.Unlock()

			go t.readConnection(newConnection)
		}(conn)
	}
}

func (t *TCP) readConnection(connection Connection) {
	for {
		log.Println("inside read connection")
		buf := make([]byte, 1024*4)
		n, err := connection.Conn.Read(buf)
		if err != nil {
			log.Printf("error reading buffer: %v\n", err)
		}
		// log.Println(buf[:n])
		msg := buf[:n]
		log.Println(msg)
		newReader := bytes.NewReader(msg)
		var version byte
		err = binary.Read(newReader, binary.BigEndian, &version)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("socket received EOF %v\n", err)
			} else {
				log.Printf("error reading binary: %v\n", err)
			}
		}
		log.Printf("got version: %v\n", version)
		if version != shared.VERSION {
			log.Printf("version mismatch we should break the connection")
			break
		}
		// log.Printf("version is %v\n", version)
		var typ byte
		err = binary.Read(newReader, binary.BigEndian, &typ)
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("socket received EOF %v\n", err)
			} else {
				log.Printf("error reading binary: %v\n", err)
			}
		}

		log.Printf("got typ: %v\n", typ)
		foundAndBreak := false
		switch typ {
		case shared.AddStickyType:
			var stickyBytes shared.AddStickyBytes = msg
			_, err := stickyBytes.ReadFrom(newReader)
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Printf("socket received EOF %v\n", err)
				} else {
					log.Printf("error reading bytes: %v\n", err)
				}
				continue
			}

			addSticky := stickyBytes.UnmarshalBinary()
			log.Printf("posterid %v\n", addSticky.PosterId)
			log.Printf("topicId %v\n", addSticky.TopicId)
			log.Println(string(addSticky.StickyMessage[:]))
			log.Println("got sticky type")
			log.Println(msg)
		case shared.VoteStickyType:
			var voteBytes shared.VoteBytes = msg
			_, err := voteBytes.ReadFrom(newReader)
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Printf("socket received EOF %v\n", err)
				} else {
					log.Printf("error reading bytes: %v\n", err)
				}
				continue
			}

			voteSticky := voteBytes.UnmarshalBinary()
			log.Printf("vote topic id %v\n", voteSticky.TopicId)
			log.Println("got vote type")
			log.Println(msg)
		case shared.QuitType:
			var quitBytes shared.QuitBytes = msg
			_, err := quitBytes.ReadFrom(newReader)
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Printf("socket received EOF %v\n", err)
				} else {
					log.Printf("error reading bytes: %v\n", err)
				}
				continue
			}

			quit := quitBytes.UnmarshalBinary()
			log.Printf("quit id %v\n", quit.ConnectionId)
			log.Println("got quit type")
			for idx, iterConn := range t.Connections {
				// Close and remove
				if iterConn.Id == connection.Id {
					err := connection.Conn.Close()
					if err != nil {
						log.Println("error closing connection")
					}
					t.Connections[idx] = t.Connections[len(t.Connections)-1]
					t.Connections = t.Connections[:len(t.Connections)-1]
					log.Println("removed connection from list")
					foundAndBreak = true
				}
			}
			log.Println(msg)
		case shared.PointToType:
			var pointToBytes shared.PointToStickyBytes = msg
			_, err := pointToBytes.ReadFrom(newReader)
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Printf("socket received EOF %v\n", err)
				} else {
					log.Printf("error reading bytes: %v\n", err)
				}
				continue
			}

			pointTo := pointToBytes.UnmarshalBinary()
			log.Printf("point to sticky id %v\n", pointTo.StickyId)
			log.Println("got point to type")
			log.Println(msg)
		default:
			log.Printf("got undefined typ: %v\n", typ)
		}
		if foundAndBreak {
			break
		}

		// TODO: We need to trigger off which type we are passing in so we can
		// update the client based on the insert/vote/quit here

		// log.Println(msg)

		// Echo
		// t.Send(buf[:n])
		// t.Incoming <- buf[:n]
		// log.Println(<-t.Incoming)
	}
}

// Change to an actual payload but bytes for now
func (t *TCP) Send(b []byte) {
	t.mutex.RLock()
	for _, connection := range t.Connections {
		n, err := connection.Conn.Write(b)
		if err != nil {
			log.Println(err)
			log.Fatal(err)
			continue
		}
		log.Println(n)
	}
	t.mutex.RUnlock()
}

func (t *TCP) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer t.Listener.Close()
	go t.acceptConnection(ctx)

	select {
	case <-ctx.Done():
		log.Println("shutting down tcp server gracefully")
		for _, connection := range t.Connections {
			_, err := connection.Conn.Write([]byte{1, 6})
			if err != nil {
				log.Printf("error writing to connection: %v\n", connection.Id)
			}
			err = connection.Conn.Close()
			if err != nil {
				log.Printf("error closing connection id: %v\n", connection.Id)
			}
		}
	}

	log.Println("tcp server shutdown")
}
