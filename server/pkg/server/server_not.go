package server

import (
	"net"
	"sync"
	"testing"
)

func TestServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Logf("%v\n", err)
		return
	}

	t.Logf("hello test\n")
	tcp := TCP{
		Connections: []Connection{},
		Listener:    listener,
		mutex:       sync.RWMutex{},
	}

	// go tcp.Start()
	connections := []net.Conn{}
	numberOf := 0
	for numberOf < 100 {
		conn, err := net.Dial("tcp", listener.Addr().String())
		if err != nil {
			t.Logf("err %v\n", err)
		}
		connections = append(connections, conn)
		numberOf++
	}

	for _, connection := range connections {
		buf := make([]byte, 1024)
		n, err := connection.Read(buf)
		if err != nil {
			t.Logf("err %v\n", err)
			continue
		}

		t.Logf("output %v\n", buf[:n])
	}

	tcp.Send([]byte{100})

	// n, err = connections[0].Read(buf)
	// if err != nil {
	// 	t.Logf("err %v\n", err)
	// 	return
	// }
	// t.Logf("output %v\n", buf[:n])

	for _, connection := range connections {
		buf := make([]byte, 1024)
		n, err := connection.Read(buf)
		t.Logf("numbbytes: %v\n", n)
		if err != nil {
			t.Logf("err %v\n", err)
			return
		}
		t.Logf("output %v\n", buf[:n])
		connection.Close()
	}
}
