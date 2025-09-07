package game

import (
	"github.com/OpenBachelor/OpenBachelorSS/internal/session"
	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
)

type SessionGameStatus struct {
	EnemyDuel *EnemyDuelGame
}

type EnemyDuelGameState interface {
	OnEnter()
	OnExit()
	Update()
}

type EnemyDuelGameBaseState struct {
	EnemyDuel *EnemyDuelGame
}

type EnemyDuelGameEntryState struct {
	Base EnemyDuelGameBaseState
}

type EnemyDuelGameBetState struct {
	Base EnemyDuelGameBaseState
}

type EnemyDuelGameBattleState struct {
	Base EnemyDuelGameBaseState
}

type EnemyDuelGameSettleState struct {
	Base EnemyDuelGameBaseState
}

type EnemyDuelGameFinishState struct {
	Base EnemyDuelGameBaseState
}

type EnemyDuelGame struct {
	state EnemyDuelGameState
}

func HandleSessionMessage(s *session.Session, g *SessionGameStatus, c contract.Content) {

}
