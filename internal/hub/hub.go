package hub

import (
	"context"
	"fmt"
	"sync"

	"github.com/OpenBachelor/OpenBachelorSS/internal/game"
	"github.com/OpenBachelor/OpenBachelorSS/internal/session"
)

type Hub struct {
	sessionsMu   sync.Mutex
	sessions     map[*session.Session]*game.SessionGameStatus
	noNewSession bool
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	closeOnce    sync.Once
	done         chan struct{}
}

func NewHub(parentCtx context.Context) *Hub {
	ctx, cancel := context.WithCancel(parentCtx)

	return &Hub{
		ctx:    ctx,
		cancel: cancel,
		done:   make(chan struct{}),
	}
}

func (h *Hub) Start() {

}

func (h *Hub) Close() {
	h.closeOnce.Do(func() {
		h.cancel()

		h.wg.Wait()

		close(h.done)
	})
}

func (h *Hub) IsClosed() bool {
	select {
	case <-h.done:
		return true
	default:
		return false
	}
}

func (h *Hub) addSession(s *session.Session) error {
	gameStatus := &game.SessionGameStatus{}

	h.sessionsMu.Lock()
	defer h.sessionsMu.Unlock()
	if h.noNewSession {
		return fmt.Errorf("no new session")
	}
	h.sessions[s] = gameStatus
	return nil
}

func (h *Hub) AddSession(s *session.Session) error {
	err := h.addSession(s)
	if err != nil {
		return err
	}

	return nil
}
