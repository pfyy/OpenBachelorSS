package main

import (
	"context"
	"log"
	"net"

	"github.com/OpenBachelor/OpenBachelorSS/internal/session"
)

func handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		log.Printf("unsupported connection type: %T", conn)
		return
	}

	log.Printf("conn: %+v", tcpConn)

	s := session.NewSession(ctx, tcpConn)
	defer s.Close()
}

func main() {
	listenAddress := "127.0.0.1:8453"

	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Printf("failed to accept: %v", err)
			continue
		}

		go handleConnection(context.Background(), conn)
	}
}
