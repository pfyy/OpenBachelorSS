package game

import (
	"github.com/OpenBachelor/OpenBachelorSS/internal/session"
	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
)

type SessionGameStatus struct{}

type EnemyDuelGame struct{}

func HandleSessionMessage(s *session.Session, g *SessionGameStatus, c contract.Content) {

}
