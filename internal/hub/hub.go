package hub

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

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

func (h *Hub) copySessions() map[*session.Session]*game.SessionGameStatus {
	h.sessionsMu.Lock()
	defer h.sessionsMu.Unlock()

	sessions := make(map[*session.Session]*game.SessionGameStatus)

	for s, g := range h.sessions {
		sessions[s] = g
	}

	return sessions
}

func (h *Hub) Start() {
	h.wg.Add(1)

	go func() {
		defer h.wg.Done()

		ticker := time.NewTicker(3 * time.Second)

		for {
			select {
			case <-ticker.C:
				sessions := h.copySessions()
				game.CloseInactiveSession(sessions)
			case <-h.ctx.Done():
				return
			}
		}
	}()
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

func (h *Hub) addSession(s *session.Session) (*game.SessionGameStatus, error) {
	g := &game.SessionGameStatus{}

	h.sessionsMu.Lock()
	defer h.sessionsMu.Unlock()
	if h.noNewSession {
		return nil, fmt.Errorf("no new session")
	}
	h.sessions[s] = g
	return g, nil
}

func (h *Hub) removeSession(s *session.Session) {
	defer h.wg.Done()

	<-s.Done()

	h.sessionsMu.Lock()
	defer h.sessionsMu.Unlock()

	delete(h.sessions, s)

	log.Printf("num of active session: %d", len(h.sessions))
}

func (h *Hub) readSession(s *session.Session, g *game.SessionGameStatus) {
	defer h.wg.Done()

	for c := range s.Recv() {
		game.HandleSessionMessage(s, g, c)
	}
}

func (h *Hub) AddSession(s *session.Session) error {
	g, err := h.addSession(s)
	if err != nil {
		s.Close()
		return err
	}

	h.wg.Add(2)
	go h.readSession(s, g)
	go h.removeSession(s)

	return nil
}
