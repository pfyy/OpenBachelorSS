package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/OpenBachelor/OpenBachelorSS/internal/session"
	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
)

type SessionGameStatus struct {
	EnemyDuel *EnemyDuelGame
}

var (
	enemyDuelGamesMu     sync.Mutex
	enemyDuelGames       = make(map[string]*EnemyDuelGame)
	noNewEnemyDuelGame   bool
	enemyDuelGamesCtx    context.Context
	enemyDuelGamesCancel context.CancelFunc
)

func SetEnemyDuelGameCtx(parentCtx context.Context) {
	enemyDuelGamesCtx, enemyDuelGamesCancel = context.WithCancel(parentCtx)
}

func StopEnemyDuelGame() {
	enemyDuelGamesCancel()

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

func unregisterGame(game *EnemyDuelGame) error {
	enemyDuelGamesMu.Lock()
	defer enemyDuelGamesMu.Unlock()

	delete(enemyDuelGames, game.GameID)

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
	GameID   string
	state    EnemyDuelGameState
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	stopOnce sync.Once
}

func NewEnemyDuelGame(gameID string) *EnemyDuelGame {
	ctx, cancel := context.WithCancel(enemyDuelGamesCtx)

	gm := &EnemyDuelGame{
		GameID: gameID,
		ctx:    ctx,
		cancel: cancel,
	}

	return gm
}

func (gm *EnemyDuelGame) Run() {
	gm.wg.Add(1)

	go func() {
		defer gm.wg.Done()

		ticker := time.NewTicker(100 * time.Millisecond)

		for {
			select {
			case <-ticker.C:
				if gm.state != nil {
					gm.state.Update()
				}
			case <-gm.ctx.Done():
				return
			}
		}
	}()
}

func (gm *EnemyDuelGame) Stop() {
	gm.stopOnce.Do(func() {
		gm.cancel()

		gm.wg.Wait()

		unregisterGame(gm)
	})
}

func HandleSessionMessage(s *session.Session, g *SessionGameStatus, c contract.Content) {

}
