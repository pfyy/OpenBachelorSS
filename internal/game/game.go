package game

import (
	"context"
	"sync"

	"github.com/OpenBachelor/OpenBachelorSS/internal/session"
	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
)

type SessionGameStatus struct {
	EnemyDuel *EnemyDuelGame
}

var (
	enemyDuelGamesMu sync.Mutex
	enemyDuelGames   = make(map[string]*EnemyDuelGame)
	ctx              context.Context
	cancel           context.CancelFunc
)

func SetEnemyDuelGameCtx(parentCtx context.Context) {
	ctx, cancel = context.WithCancel(parentCtx)
}

func StopEnemyDuelGame() {
	cancel()
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
