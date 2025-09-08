package game

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/OpenBachelor/OpenBachelorSS/internal/session"
	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
)

type EnemyDuelGamePlayerStatus struct {
	Money  uint32
	Side   uint8
	AllIn  uint8
	Streak uint8
}

type SessionGameStatus struct {
	EnemyDuel                 *EnemyDuelGame
	EnemyDuelGamePlayerStatus EnemyDuelGamePlayerStatus
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

func StopEnemyDuelGames() {
	enemyDuelGamesCancel()

	setNoNewGame()

	stopEnemyDuelGames()
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

	if _, ok := enemyDuelGames[game.GameID]; ok {
		return fmt.Errorf("game already exist")
	}

	enemyDuelGames[game.GameID] = game

	game.wg.Add(1)
	go func(g *EnemyDuelGame) {
		defer g.wg.Done()

		<-g.ctx.Done()

		unregisterGame(g)
	}(game)

	return nil
}

func unregisterGame(game *EnemyDuelGame) error {
	enemyDuelGamesMu.Lock()
	defer enemyDuelGamesMu.Unlock()

	delete(enemyDuelGames, game.GameID)

	return nil
}

func getGame(gameID string) *EnemyDuelGame {
	enemyDuelGamesMu.Lock()
	defer enemyDuelGamesMu.Unlock()

	if game, ok := enemyDuelGames[gameID]; ok {
		return game
	}

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

func stopEnemyDuelGames() {
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

func (s *EnemyDuelGameEntryState) OnEnter() {

}

func (s *EnemyDuelGameEntryState) OnExit() {

}

func (s *EnemyDuelGameEntryState) Update() {

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
	GameID         string
	ModeID         string
	StageID        string
	state          EnemyDuelGameState
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	stopOnce       sync.Once
	sessionsMu     sync.Mutex
	sessions       map[*session.Session]*SessionGameStatus
	playerStatusMu sync.Mutex
}

func NewEnemyDuelGame(gameID string, modeID string, stageID string) *EnemyDuelGame {
	ctx, cancel := context.WithCancel(enemyDuelGamesCtx)

	gm := &EnemyDuelGame{
		GameID:   gameID,
		ModeID:   modeID,
		StageID:  stageID,
		ctx:      ctx,
		cancel:   cancel,
		sessions: make(map[*session.Session]*SessionGameStatus),
	}

	gm.SetState(&EnemyDuelGameEntryState{Base: EnemyDuelGameBaseState{EnemyDuel: gm}})

	return gm
}

func (gm *EnemyDuelGame) SetState(newState EnemyDuelGameState) {
	if gm.state != nil {
		gm.state.OnExit()
	}
	gm.state = newState
	if gm.state != nil {
		gm.state.OnEnter()
	}
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
				} else {
					gm.cancel()
					return
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
	})
}

func (gm *EnemyDuelGame) AddSession(s *session.Session, g *SessionGameStatus) {
	gm.sessionsMu.Lock()
	defer gm.sessionsMu.Unlock()

	gm.sessions[s] = g
}

func getModeIDStageID(teamToken string) (string, string, error) {
	parts := strings.Split(teamToken, "|")

	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid teamToken")
	}

	modeID, stageID := parts[0], parts[1]

	return modeID, stageID, nil
}

func getOrCreateGame(gameID string, modeID string, stageID string) *EnemyDuelGame {
	game := getGame(gameID)
	if game == nil {
		game = NewEnemyDuelGame(gameID, modeID, stageID)
		err := registerGame(game)
		if err == nil {
			game.Run()
		} else {
			log.Printf("failed to register game: %v", err)
			game = getGame(gameID)
		}
	}

	return game
}

func HandleSessionMessage(s *session.Session, g *SessionGameStatus, c contract.Content) {
	if msg, ok := c.(*contract.C2SEnemyDuelHeartBeatMessage); ok {
		s.SendMessage(contract.NewS2CEnemyDuelHeartBeatMessage(msg.Seq, msg.Time))
		return
	}

	if msg, ok := c.(*contract.C2SEnemyDuelTeamJoinMessage); ok {
		defer s.SendMessage(contract.NewS2CEnemyDuelKickMessage())

		modeID, stageID, err := getModeIDStageID(msg.TeamToken)
		if err != nil {
			log.Printf("failed to get modeID, stageID: %v", err)
			return
		}

		gameID := strings.Join([]string{msg.TeamID, modeID, stageID}, "|")

		game := getOrCreateGame(gameID, modeID, stageID)

		if game == nil {
			log.Printf("failed to get or create game")
			return
		}

		s.SendMessage(contract.NewS2CEnemyDuelTeamJoinMessage())

		s.SendMessage(contract.NewS2CEnemyDuelTeamStatusMessage(msg.TeamID, msg.TeamToken))

		return
	}

	if msg, ok := c.(*contract.C2SEnemyDuelJoinMessage); ok {
		modeID, stageID, err := getModeIDStageID(msg.Token)
		if err != nil {
			log.Printf("failed to get modeID, stageID: %v", err)
			return
		}

		gameID := strings.Join([]string{msg.SceneID, modeID, stageID}, "|")

		game := getGame(gameID)
		if game == nil {
			defer s.SendMessage(contract.NewS2CEnemyDuelEndMessage())

			log.Printf("game not found")

			return
		}

		game.AddSession(s, g)
		g.EnemyDuel = game

		return
	}
}
