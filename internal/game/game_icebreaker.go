package game

import (
	"github.com/OpenBachelor/OpenBachelorSS/internal/session"
	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
)

type IceBreakerSessionGameStatus struct {
	BaseSessionGameStatus
}

func (g *IceBreakerSessionGameStatus) GetBaseSessionGameStatus() *BaseSessionGameStatus {
	return &g.BaseSessionGameStatus
}

func handleSessionMessageIceBreaker(s *session.Session, g *IceBreakerSessionGameStatus, c contract.Content) {

}
