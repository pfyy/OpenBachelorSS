package hub

import "context"

type Hub struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func NewHub(parentCtx context.Context) *Hub {
	ctx, cancel := context.WithCancel(parentCtx)

	return &Hub{
		ctx:    ctx,
		cancel: cancel,
	}
}
