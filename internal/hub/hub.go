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
		sessions: make(map[*session.Session]*game.SessionGameStatus),
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
	}
}

func (h *Hub) Start() {

}

func (h *Hub) Close() {
	h.closeOnce.Do(func() {
		h.cancel()

		h.setNoNewSession()

		h.closeSessions()

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

func (h *Hub) setNoNewSession() {
	h.sessionsMu.Lock()
	defer h.sessionsMu.Unlock()

	h.noNewSession = true
}

func (h *Hub) getSessions() []*session.Session {
	h.sessionsMu.Lock()
	defer h.sessionsMu.Unlock()

	sessions := make([]*session.Session, 0, len(h.sessions))

	for s := range h.sessions {
		sessions = append(sessions, s)
	}

	return sessions
}

func (h *Hub) closeSessions() {
	sessions := h.getSessions()

	for _, s := range sessions {
		s.Close()
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

func (h *Hub) removeSession(s *session.Session) {
	defer h.wg.Done()

	<-s.Done()

	h.sessionsMu.Lock()
	defer h.sessionsMu.Unlock()

	delete(h.sessions, s)
}

func (h *Hub) AddSession(s *session.Session) error {
	err := h.addSession(s)
	if err != nil {
		return err
	}

	h.wg.Add(1)
	go h.removeSession(s)

	return nil
}
