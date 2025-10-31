package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os/signal"
	"sync"
	"syscall"

	"github.com/OpenBachelor/OpenBachelorSS/internal/config"
	"github.com/OpenBachelor/OpenBachelorSS/internal/game"
	"github.com/OpenBachelor/OpenBachelorSS/internal/hub"
	"github.com/OpenBachelor/OpenBachelorSS/internal/session"
	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
)

func handleConnection(ctx context.Context, conn net.Conn, h *hub.Hub, wg *sync.WaitGroup, msgDomain contract.MessageDomain) {
	defer wg.Done()

	cfg := config.Get()

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		log.Printf("unsupported connection type: %T", conn)
		return
	}

	if cfg.Server.Debug {
		log.Printf("conn: %+v", tcpConn.RemoteAddr())
	}

	s := session.NewSession(msgDomain, ctx, tcpConn)
	s.Start()

	err := h.AddSession(s)
	if err != nil {
		log.Printf("failed to add session: %v", err)
	}
}

func mainLoop(ctx context.Context, h *hub.Hub, addr string, msgDomain contract.MessageDomain) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

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

		wg.Add(1)
		go handleConnection(ctx, conn, h, &wg, msgDomain)
	}

	wg.Wait()

	return nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	game.SetEnemyDuelGameCtx(ctx)
	defer game.StopEnemyDuelGames()

	h := hub.NewHub(ctx)
	h.Start()
	defer h.Close()

	cfg := config.Get()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := mainLoop(ctx, h, cfg.Server.Addr, contract.EnemyDuelMessageDomain)

		if err != nil {
			log.Fatalf("failed to start main loop: %v", err)
		}

	}()

	wg.Wait()

	log.Printf("closing server")
}
