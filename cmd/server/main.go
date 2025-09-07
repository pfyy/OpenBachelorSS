package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/OpenBachelor/OpenBachelorSS/internal/config"
	"github.com/OpenBachelor/OpenBachelorSS/internal/session"
)

func handleConnection(ctx context.Context, conn net.Conn) {
	cfg := config.Get()

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		log.Printf("unsupported connection type: %T", conn)
		return
	}

	if cfg.Server.Debug {
		log.Printf("conn: %+v", tcpConn.RemoteAddr())
	}

	s := session.NewSession(ctx, tcpConn)
	s.Start()
	defer s.Close()
}

func mainLoop(ctx context.Context) error {
	cfg := config.Get()

	listener, err := net.Listen("tcp", cfg.Server.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
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

	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err := mainLoop(ctx)

	if err != nil {
		log.Fatalf("failed to start main loop: %v", err)
	}
}
