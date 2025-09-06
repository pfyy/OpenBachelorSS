package session

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
)

type Conn interface {
	net.Conn
	CloseRead() error
}

type Session struct {
	conn    Conn
	wg      sync.WaitGroup
	send    chan contract.Content
	recv    chan contract.Content
	ctx     context.Context
	cancel  context.CancelFunc
	errOnce sync.Once
	err     error
}

func NewSession(parentCtx context.Context, conn Conn) *Session {
	ctx, cancel := context.WithCancel(parentCtx)

	return &Session{
		conn:   conn,
		send:   make(chan contract.Content),
		recv:   make(chan contract.Content),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Session) close() {
	s.conn.CloseRead()

	s.cancel()
}

func (s *Session) setErr(err error) {
	s.errOnce.Do(func() {
		s.err = err
		s.close()
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
			for content := range s.send {
				err := contract.WriteContent(s.conn, content)
				if err != nil {
					s.setErr(err)
					return
				}
			}
		}
	}
}

func (s *Session) Start() {
	s.wg.Add(2)
	go s.readLoop()
	go s.writeLoop()
}

func (s *Session) Close() {
	s.close()

	s.wg.Wait()
	s.conn.Close()
}
