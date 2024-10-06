package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"pkg/server"
	"sync"
	"syscall"
)

func main() {
	log.Println("Starting Server")
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	tcpServer := server.NewTcpServer("127.0.0.1:3000")
	log.Println("tcp server %v\n", tcpServer.Connections)
	wg.Add(1)
	
	go tcpServer.Start(ctx, &wg)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	<-signalCh

	cancel()
	wg.Wait()
	log.Println("Stopping Server")
}
