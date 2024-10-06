package server

import (
	"context"
	"log"
	"net"
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

			n, err := newConnection.Conn.Write([]byte{1, 0})
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
		buf := make([]byte, 1024)
		n, err := connection.Conn.Read(buf)
		if err != nil {
			log.Println(err)
		}
		log.Println(buf[:n])
		// Echo
		t.Send(buf[:n])
		t.Incoming <- buf[:n]
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
