package session

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/OpenBachelor/OpenBachelorSS/internal/config"
	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
)

type Conn interface {
	net.Conn
	CloseRead() error
}

type Session struct {
	conn      Conn
	wg        sync.WaitGroup
	send      chan contract.Content
	recv      chan contract.Content
	ctx       context.Context
	cancel    context.CancelFunc
	errMu     sync.Mutex
	readErr   error
	writeErr  error
	closeOnce sync.Once
	done      chan struct{}
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
		done:   make(chan struct{}),
	}
}

func (s *Session) setReadErr(err error) {
	s.errMu.Lock()
	defer s.errMu.Unlock()
	if s.readErr != nil {
		return
	}
	cfg := config.Get()
	if cfg.Server.Debug {
		log.Printf("%+v readErr: %v", s.conn.RemoteAddr(), err)
	}
	s.readErr = err
}

func (s *Session) readLoop() {
	defer s.wg.Done()
	defer close(s.recv)

	type ReadResult struct {
		content contract.Content
		err     error
	}
	readResultChan := make(chan ReadResult)
	go func() {
		for {
			// this line is still blocking on windows even if CloseRead() is called, fuck windows
			content, err := contract.ReadContent(s.conn)

			readResultChan <- ReadResult{content: content, err: err}

			if err != nil {
				return
			}
		}
	}()

	for {
		var content contract.Content
		var err error

		select {
		case readResult := <-readResultChan:
			content, err = readResult.content, readResult.err
			if err == io.EOF {
				return
			}

			if err != nil {
				// windows does not return EOF after calling CloseRead()
				if s.ctx.Err() == nil {
					s.setReadErr(err)
				}
				return
			}

		case <-s.ctx.Done():
			return
		}

		select {
		case s.recv <- content:
			cfg := config.Get()
			if cfg.Server.Debug {
				log.Printf("%+v -> %#v", s.conn.RemoteAddr(), content)
			}
		case <-s.ctx.Done():
			return
		}
	}
}

const writeTimeout = 10 * time.Second

func (s *Session) setWriteErr(err error) {
	s.errMu.Lock()
	defer s.errMu.Unlock()
	if s.writeErr != nil {
		return
	}
	cfg := config.Get()
	if cfg.Server.Debug {
		log.Printf("%+v writeErr: %v", s.conn.RemoteAddr(), err)
	}
	s.writeErr = err
}

func (s *Session) writeContent(content contract.Content, ok bool) bool {
	if !ok {
		return true
	}

	cfg := config.Get()
	if cfg.Server.Debug {
		log.Printf("%+v <- %#v", s.conn.RemoteAddr(), content)
	}

	err := s.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	if err != nil {
		s.setWriteErr(err)
		return true
	}

	err = contract.WriteContent(s.conn, content)
	if err != nil {
		s.setWriteErr(err)
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

func (s *Session) Err() (error, error) {
	s.errMu.Lock()
	defer s.errMu.Unlock()
	return s.readErr, s.writeErr
}

func (s *Session) Send() chan<- contract.Content {
	return s.send
}

func (s *Session) Recv() <-chan contract.Content {
	return s.recv
}

func (s *Session) SendMessage(content contract.Content) error {
	select {
	case s.send <- content:

	default:
		err := fmt.Errorf("misbehaving client")
		return err
	}

	return nil
}

func (s *Session) Start() {
	s.wg.Add(2)
	go s.readLoop()
	go s.writeLoop()
}

func (s *Session) Close() {
	s.closeOnce.Do(func() {
		s.cancel()

		s.conn.CloseRead()

		s.wg.Wait()

		s.conn.Close()

		close(s.done)
	})
}

func (s *Session) IsClosed() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

func (s *Session) Done() <-chan struct{} {
	return s.done
}
