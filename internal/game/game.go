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
	Money       uint32
	ShieldState uint8
	Streak      uint8
	Side        uint8
	AllIn       uint8
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

type EnemyDuelGameStateBase struct {
	EnemyDuel     *EnemyDuelGame
	EnterTime     int64
	ForceExitTime int64
}

func (b *EnemyDuelGameStateBase) SetEnterTime() {
	b.EnterTime = time.Now().Unix()
}

func (b *EnemyDuelGameStateBase) SetForceExitTime(maxDuration int64) {
	b.ForceExitTime = b.EnterTime + maxDuration
}

type EnemyDuelGameWaitingState struct {
	EnemyDuelGameStateBase
}

func (s *EnemyDuelGameWaitingState) OnEnter() {
	s.SetEnterTime()
	s.SetForceExitTime(30)
}

func (s *EnemyDuelGameWaitingState) OnExit() {

}

func (s *EnemyDuelGameWaitingState) Update() {
	currentTime := time.Now().Unix()

	if currentTime > s.ForceExitTime {
		s.EnemyDuel.SetState(&EnemyDuelGameEntryState{EnemyDuelGameStateBase: EnemyDuelGameStateBase{EnemyDuel: s.EnemyDuel}})
		return
	}
}

type EnemyDuelGameEntryState struct {
	EnemyDuelGameStateBase
}

func (s *EnemyDuelGameEntryState) OnEnter() {
	s.SetEnterTime()
	s.SetForceExitTime(3)

	s.EnemyDuel.clearStep()

	sessions := s.EnemyDuel.getSessions()

	for session := range sessions {
		session.SendMessage(contract.NewS2CEnemyDuelClientStateMessage(1, s.EnemyDuel.round, s.ForceExitTime))
	}
}

func (s *EnemyDuelGameEntryState) OnExit() {

}

func (s *EnemyDuelGameEntryState) Update() {
	currentTime := time.Now().Unix()

	if currentTime > s.ForceExitTime {
		s.EnemyDuel.SetState(&EnemyDuelGameBetState{EnemyDuelGameStateBase: EnemyDuelGameStateBase{EnemyDuel: s.EnemyDuel}})
		return
	}

	s.EnemyDuel.doStep()
}

type EnemyDuelGameBetState struct {
	EnemyDuelGameStateBase
}

func (s *EnemyDuelGameBetState) OnEnter() {
	s.SetEnterTime()
	s.SetForceExitTime(20)

	sessions := s.EnemyDuel.getSessions()

	for session := range sessions {
		session.SendMessage(contract.NewS2CEnemyDuelClientStateMessage(2, s.EnemyDuel.round, s.ForceExitTime))
	}
}

func (s *EnemyDuelGameBetState) OnExit() {

}

func (s *EnemyDuelGameBetState) Update() {
	currentTime := time.Now().Unix()

	if currentTime > s.ForceExitTime {
		s.EnemyDuel.SetState(&EnemyDuelGameBattleState{EnemyDuelGameStateBase: EnemyDuelGameStateBase{EnemyDuel: s.EnemyDuel}})
		return
	}
}

type EnemyDuelGameBattleState struct {
	EnemyDuelGameStateBase
}

func (s *EnemyDuelGameBattleState) OnEnter() {
	s.SetEnterTime()
	s.SetForceExitTime(150)

	sessions := s.EnemyDuel.getSessions()

	for session := range sessions {
		session.SendMessage(contract.NewS2CEnemyDuelClientStateMessage(3, s.EnemyDuel.round, s.ForceExitTime))
	}
}

func (s *EnemyDuelGameBattleState) OnExit() {

}

func (s *EnemyDuelGameBattleState) Update() {
	currentTime := time.Now().Unix()

	if currentTime > s.ForceExitTime {
		s.EnemyDuel.SetState(&EnemyDuelGameSettleState{EnemyDuelGameStateBase: EnemyDuelGameStateBase{EnemyDuel: s.EnemyDuel}})
		return
	}

	s.EnemyDuel.doStep()
}

type EnemyDuelGameSettleState struct {
	EnemyDuelGameStateBase
}

func (s *EnemyDuelGameSettleState) OnEnter() {
	s.SetEnterTime()
	s.SetForceExitTime(10)

	sessions := s.EnemyDuel.getSessions()

	for session := range sessions {
		session.SendMessage(contract.NewS2CEnemyDuelClientStateMessage(4, s.EnemyDuel.round, s.ForceExitTime))
	}
}

func (s *EnemyDuelGameSettleState) OnExit() {

}

func (s *EnemyDuelGameSettleState) Update() {
	currentTime := time.Now().Unix()

	if currentTime > s.ForceExitTime {
		maxRound := s.EnemyDuel.getMaxRound()
		if s.EnemyDuel.round+1 >= maxRound {
			s.EnemyDuel.SetState(&EnemyDuelGameFinishState{EnemyDuelGameStateBase: EnemyDuelGameStateBase{EnemyDuel: s.EnemyDuel}})
		} else {
			s.EnemyDuel.round++
			s.EnemyDuel.SetState(&EnemyDuelGameEntryState{EnemyDuelGameStateBase: EnemyDuelGameStateBase{EnemyDuel: s.EnemyDuel}})
		}

		return
	}
}

type EnemyDuelGameFinishState struct {
	EnemyDuelGameStateBase
}

func (s *EnemyDuelGameFinishState) OnEnter() {
	s.SetEnterTime()
	s.SetForceExitTime(10)

	sessions := s.EnemyDuel.getSessions()

	for session := range sessions {
		session.SendMessage(contract.NewS2CEnemyDuelClientStateMessage(5, s.EnemyDuel.round, s.ForceExitTime))
	}
}

func (s *EnemyDuelGameFinishState) OnExit() {
	sessions := s.EnemyDuel.getSessions()

	for session := range sessions {
		session.SendMessage(contract.NewC2SEnemyDuelFinalSettleMessage())
		session.SendMessage(contract.NewS2CEnemyDuelEndMessage())
	}
}

func (s *EnemyDuelGameFinishState) Update() {
	s.EnemyDuel.SetState(nil)
}

type EnemyDuelGame struct {
	GameID     string
	ModeID     string
	StageID    string
	state      EnemyDuelGameState
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	stopOnce   sync.Once
	sessionsMu sync.Mutex
	sessions   map[*session.Session]*SessionGameStatus
	round      uint8
	step       uint32
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

	gm.SetState(&EnemyDuelGameWaitingState{EnemyDuelGameStateBase: EnemyDuelGameStateBase{EnemyDuel: gm}})

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

func (gm *EnemyDuelGame) getSessions() map[*session.Session]*SessionGameStatus {
	gm.sessionsMu.Lock()
	defer gm.sessionsMu.Unlock()

	sessions := make(map[*session.Session]*SessionGameStatus)
	for s, g := range gm.sessions {
		sessions[s] = g
	}

	return sessions
}

func (gm *EnemyDuelGame) getMaxRound() uint8 {
	switch gm.ModeID {
	case "multiOperationMatch":
		return 10

	}

	return 0xff
}

func (gm *EnemyDuelGame) doStep() {
	sessions := gm.getSessions()

	for session := range sessions {
		session.SendMessage(contract.NewS2CEnemyDuelStepMessage(gm.step, gm.round))

	}

	gm.step++
}

func (gm *EnemyDuelGame) clearStep() {
	gm.step = 0
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

	if _, ok := c.(*contract.C2SEnemyDuelQuitMessage); ok {
		s.SendMessage(contract.NewS2CEnemyDuelQuitMessage())
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
			log.Printf("game not found")
			return
		}

		game.AddSession(s, g)
		g.EnemyDuel = game

		s.SendMessage(contract.NewS2CEnemyDuelJoinMessage(stageID))

		return
	}
}
