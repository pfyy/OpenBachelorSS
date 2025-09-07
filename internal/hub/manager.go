package hub

import (
	"context"
	"sync"
)

type Hub struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	closeOnce sync.Once
	done      chan struct{}
}

func NewHub(parentCtx context.Context) *Hub {
	ctx, cancel := context.WithCancel(parentCtx)

	return &Hub{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (h *Hub) Start() {
	go func() {
		<-h.ctx.Done()
		h.Close()
	}()
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
