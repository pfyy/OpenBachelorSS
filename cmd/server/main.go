package main

import (
	"context"
	"log"
	"net"

	"github.com/OpenBachelor/OpenBachelorSS/internal/session"
)

func handleConnection(ctx context.Context, conn session.Conn) {
	defer conn.Close()

	log.Printf("conn: %+v", conn)

	session := session.NewSession(ctx, conn)
	defer session.Close()
}

func main() {
	listenAddress := "127.0.0.1:8453"

	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	defer listener.Close()

	for {
		netConn, err := listener.Accept()

		if err != nil {
			log.Printf("failed to accept: %v", err)
			continue
		}

		conn, ok := netConn.(session.Conn)
		if !ok {
			log.Printf("not a valid conn type: %T", netConn)
			continue
		}

		go handleConnection(context.Background(), conn)
	}
}
