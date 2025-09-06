package session

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
)

type Session struct {
	conn    net.Conn
	wg      sync.WaitGroup
	send    chan contract.Content
	recv    chan contract.Content
	ctx     context.Context
	cancel  context.CancelFunc
	errOnce sync.Once
	err     error
}

func NewSession(parentCtx context.Context, conn net.Conn) *Session {
	ctx, cancel := context.WithCancel(parentCtx)

	return &Session{
		conn:   conn,
		send:   make(chan contract.Content),
		recv:   make(chan contract.Content),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Session) setErr(err error) {
	s.errOnce.Do(func() {
		s.err = err
		s.cancel()
	})
}

func (s *Session) readLoop() {
	defer s.wg.Done()
	defer close(s.recv)

	for {
		content, err := contract.ReadContent(s.conn)
		if err == io.EOF {
			return
		}

		if err != nil {
			s.setErr(err)
			return
		}

		select {
		case s.recv <- content:
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Session) writeLoop() {
	defer s.wg.Done()

	for {
		select {
		case content, ok := <-s.send:
			if !ok {
				return
			}

			err := contract.WriteContent(s.conn, content)
			if err != nil {
				s.setErr(err)
				return
			}

		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Session) Start() {
	s.wg.Add(2)
	go s.readLoop()
	go s.writeLoop()
}

func (s *Session) Close() {
	s.cancel()
	s.wg.Wait()

	s.conn.Close()
}
