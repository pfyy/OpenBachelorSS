package session

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
)

type Conn interface {
	net.Conn
	CloseRead() error
}

type Session struct {
	conn     Conn
	wg       sync.WaitGroup
	send     chan contract.Content
	recv     chan contract.Content
	ctx      context.Context
	cancel   context.CancelFunc
	readErr  error
	writeErr error
}

const sessionChanSize = 1024

func NewSession(parentCtx context.Context, conn Conn) *Session {
	ctx, cancel := context.WithCancel(parentCtx)

	return &Session{
		conn:   conn,
		send:   make(chan contract.Content, sessionChanSize),
		recv:   make(chan contract.Content, sessionChanSize),
		ctx:    ctx,
		cancel: cancel,
	}
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
			if s.ctx.Err() == nil {
				s.readErr = err
			}
			return
		}

		select {
		case s.recv <- content:
		case <-s.ctx.Done():
			return
		}
	}
}

const writeTimeout = 10 * time.Second

func (s *Session) writeContent(content contract.Content, ok bool) bool {
	if !ok {
		return true
	}

	err := s.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err != nil {
		s.writeErr = err
		return true
	}

	err = contract.WriteContent(s.conn, content)
	if err != nil {
		s.writeErr = err
		return true
	}

	return false
}

func (s *Session) writeLoop() {
	defer s.wg.Done()

	for {
		select {
		case content, ok := <-s.send:
			if s.writeContent(content, ok) {
				return
			}

		case <-s.ctx.Done():
			for {
				select {
				case content, ok := <-s.send:
					if s.writeContent(content, ok) {
						return
					}

				default:
					return
				}
			}
		}
	}
}

func (s *Session) Send() chan<- contract.Content {
	return s.send
}

func (s *Session) Recv() <-chan contract.Content {
	return s.recv
}

func (s *Session) Start() {
	s.wg.Add(2)
	go s.readLoop()
	go s.writeLoop()

	go func() {
		<-s.ctx.Done()
		s.Close()
	}()
}

func (s *Session) Close() {
	s.cancel()

	s.conn.CloseRead()

	s.wg.Wait()

	s.conn.Close()
}
