package game

import (
	"context"
	"fmt"
	"sync"

	"github.com/OpenBachelor/OpenBachelorSS/internal/session"
	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
)

type SessionGameStatus struct {
	EnemyDuel *EnemyDuelGame
}

var (
	enemyDuelGamesMu   sync.Mutex
	enemyDuelGames     = make(map[string]*EnemyDuelGame)
	noNewEnemyDuelGame bool
	ctx                context.Context
	cancel             context.CancelFunc
)

func SetEnemyDuelGameCtx(parentCtx context.Context) {
	ctx, cancel = context.WithCancel(parentCtx)
}

func StopEnemyDuelGame() {
	cancel()

	setNoNewGame()

	stopEnemyDuelGame()
}

func setNoNewGame() {
	enemyDuelGamesMu.Lock()
	defer enemyDuelGamesMu.Unlock()

	noNewEnemyDuelGame = true
}

func registerGame(game *EnemyDuelGame) error {
	enemyDuelGamesMu.Lock()
	defer enemyDuelGamesMu.Unlock()

	if noNewEnemyDuelGame {
		return fmt.Errorf("no new game")
	}

	enemyDuelGames[game.GameID] = game

	return nil
}

func getEnemyDuelGamesSlice() []*EnemyDuelGame {
	enemyDuelGamesMu.Lock()
	defer enemyDuelGamesMu.Unlock()

	enemyDuelGamesSlice := make([]*EnemyDuelGame, 0, len(enemyDuelGames))

	for _, game := range enemyDuelGames {
		enemyDuelGamesSlice = append(enemyDuelGamesSlice, game)
	}

	return enemyDuelGamesSlice
}

func stopEnemyDuelGame() {
	enemyDuelGamesSlice := getEnemyDuelGamesSlice()

	for _, game := range enemyDuelGamesSlice {
		game.Stop()
	}
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
	GameID string
	state  EnemyDuelGameState
}

func (gm *EnemyDuelGame) Stop() {

}

func HandleSessionMessage(s *session.Session, g *SessionGameStatus, c contract.Content) {

}
