package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/OpenBachelor/OpenBachelorSS/internal/config"
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

func listen(ctx context.Context) {
	cfg := config.Get()

	listener, err := net.Listen("tcp", cfg.Server.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go func() {
		<-ctx.Done()
		if err := listener.Close(); err != nil {
			log.Printf("failed to close listener: %v", err)
		}
	}()

	for {
		conn, err := listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			break
		}

		if err != nil {
			log.Printf("failed to accept: %v", err)
			continue
		}

		go handleConnection(ctx, conn)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	listen(ctx)
}
