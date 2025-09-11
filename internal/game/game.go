package game

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/OpenBachelor/OpenBachelorSS/internal/session"
	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
)

type EnemyDuelGamePlayerStatus struct {
	PlayerID         string
	internalPlayerID int
	Money            uint32
	ShieldState      uint8
	Streak           uint8
	Side             uint8
	AllIn            uint8
	ReportSide       uint8
}

func getExternalPlayerID(internalPlayerID int) string {
	return strconv.Itoa(100 + internalPlayerID)
}

func (s *EnemyDuelGamePlayerStatus) getExternalPlayerID() string {
	return getExternalPlayerID(s.internalPlayerID)
}

type SessionGameStatus struct {
	LastActiveTime            time.Time
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

	log.Printf("num of active game: %d", len(enemyDuelGames))

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
	GetForceExitTime() time.Time
}

type EnemyDuelGameStateBase struct {
	EnemyDuel     *EnemyDuelGame
	EnterTime     time.Time
	ForceExitTime time.Time
}

func (b *EnemyDuelGameStateBase) SetEnterTime() {
	b.EnterTime = time.Now()
}

func (b *EnemyDuelGameStateBase) SetForceExitTime(maxDuration time.Duration) {
	b.ForceExitTime = b.EnterTime.Add(maxDuration)
}

func (b *EnemyDuelGameStateBase) GetForceExitTime() time.Time {
	return b.ForceExitTime
}

type EnemyDuelGameWaitingState struct {
	EnemyDuelGameStateBase
}

func (s *EnemyDuelGameWaitingState) OnEnter() {
	s.SetEnterTime()
	s.SetForceExitTime(30 * time.Second)
}

func (s *EnemyDuelGameWaitingState) OnExit() {

}

func (s *EnemyDuelGameWaitingState) Update() {
	currentTime := time.Now()

	if currentTime.After(s.ForceExitTime) {
		s.EnemyDuel.setNoNewSession()
		s.EnemyDuel.SetState(&EnemyDuelGameEntryState{EnemyDuelGameStateBase: EnemyDuelGameStateBase{EnemyDuel: s.EnemyDuel}})
		return
	}
}

type EnemyDuelGameEntryState struct {
	EnemyDuelGameStateBase
}

func (s *EnemyDuelGameEntryState) OnEnter() {
	s.SetEnterTime()
	s.SetForceExitTime(3 * time.Second)

	s.EnemyDuel.clearStep()
	s.EnemyDuel.clearState()

	sessions := s.EnemyDuel.getSessions()

	for session := range sessions {
		session.SendMessage(contract.NewS2CEnemyDuelClientStateMessage(1, s.EnemyDuel.round, s.ForceExitTime))
	}
}

func (s *EnemyDuelGameEntryState) OnExit() {

}

func (s *EnemyDuelGameEntryState) Update() {
	currentTime := time.Now()

	if currentTime.After(s.ForceExitTime) {
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
	s.SetForceExitTime(20 * time.Second)

	sessions := s.EnemyDuel.getSessions()

	for session := range sessions {
		session.SendMessage(contract.NewS2CEnemyDuelClientStateMessage(2, s.EnemyDuel.round, s.ForceExitTime))
	}
}

func (s *EnemyDuelGameBetState) OnExit() {

}

func (s *EnemyDuelGameBetState) Update() {
	currentTime := time.Now()

	if currentTime.After(s.ForceExitTime) {
		s.EnemyDuel.SetState(&EnemyDuelGameBattleState{EnemyDuelGameStateBase: EnemyDuelGameStateBase{EnemyDuel: s.EnemyDuel}})
		return
	}
}

type EnemyDuelGameBattleState struct {
	EnemyDuelGameStateBase
}

func (s *EnemyDuelGameBattleState) OnEnter() {
	s.SetEnterTime()
	s.SetForceExitTime(150 * time.Second)

	sessions := s.EnemyDuel.getSessions()

	for session := range sessions {
		session.SendMessage(contract.NewS2CEnemyDuelClientStateMessage(3, s.EnemyDuel.round, s.ForceExitTime))
	}
}

func (s *EnemyDuelGameBattleState) OnExit() {

}

func (s *EnemyDuelGameBattleState) Update() {
	reportSide := s.EnemyDuel.getReportSide()

	currentTime := time.Now()

	if reportSide != 0 || currentTime.After(s.ForceExitTime) {
		s.EnemyDuel.reportSide = reportSide
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
	s.SetForceExitTime(10 * time.Second)

	sessions := s.EnemyDuel.getSessions()

	for session, gameStatus := range sessions {
		allPlayerResult := s.EnemyDuel.getAllPlayerResult()

		session.SendMessage(contract.NewS2CEnemyDuelClientStateMessageForSettle(4, s.EnemyDuel.round, s.ForceExitTime, allPlayerResult, gameStatus.EnemyDuelGamePlayerStatus.PlayerID, gameStatus.EnemyDuelGamePlayerStatus.getExternalPlayerID()))
	}
}

func (s *EnemyDuelGameSettleState) OnExit() {

}

func (s *EnemyDuelGameSettleState) Update() {
	currentTime := time.Now()

	if currentTime.After(s.ForceExitTime) {
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
	s.SetForceExitTime(10 * time.Second)

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
	GameID               string
	ModeID               string
	StageID              string
	state                EnemyDuelGameState
	wg                   sync.WaitGroup
	ctx                  context.Context
	cancel               context.CancelFunc
	stopOnce             sync.Once
	sessionsMu           sync.Mutex
	sessions             map[*session.Session]*SessionGameStatus
	nextInternalPlayerID int
	noNewSession         bool
	round                uint8
	step                 uint32
	reportSide           uint8
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
		defer gm.cancel()

		ticker := time.NewTicker(100 * time.Millisecond)

		for {
			select {
			case <-ticker.C:
				if !gm.hasAliveSession() {
					return
				}

				if gm.state != nil {
					gm.state.Update()
				} else {
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

func (gm *EnemyDuelGame) AddSession(s *session.Session, g *SessionGameStatus) error {
	gm.sessionsMu.Lock()
	defer gm.sessionsMu.Unlock()

	if gm.noNewSession {
		return fmt.Errorf("game no new session")
	}

	if gm.nextInternalPlayerID >= gm.getMaxNumPlayer() {
		return fmt.Errorf("game full")
	}

	gm.sessions[s] = g

	g.EnemyDuel = gm
	g.EnemyDuelGamePlayerStatus.internalPlayerID = gm.nextInternalPlayerID
	gm.nextInternalPlayerID++

	return nil
}

func (gm *EnemyDuelGame) setNoNewSession() {
	gm.sessionsMu.Lock()
	defer gm.sessionsMu.Unlock()

	gm.noNewSession = true
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

func (gm *EnemyDuelGame) getMaxNumPlayer() int {
	switch gm.ModeID {
	case "multiOperationMatch":
		return 8

	}

	return 30
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

func (gm *EnemyDuelGame) clearState() {
	sessions := gm.getSessions()

	for _, g := range sessions {
		g.EnemyDuelGamePlayerStatus.Side = 0
		g.EnemyDuelGamePlayerStatus.AllIn = 0
		g.EnemyDuelGamePlayerStatus.ReportSide = 0
	}
}

func (gm *EnemyDuelGame) getReportSide() uint8 {
	sessions := gm.getSessions()

	reportSideCntMap := make(map[uint8]int)

	for s, g := range sessions {
		if !s.IsClosed() {
			reportSideCntMap[g.EnemyDuelGamePlayerStatus.ReportSide]++
		}
	}

	reportSide := uint8(0)
	reportSideCnt := 0

	for k, v := range reportSideCntMap {
		if v > reportSideCnt {
			reportSide = k
			reportSideCnt = v
		}
	}

	return reportSide
}

func (gm *EnemyDuelGame) hasAliveSession() bool {
	sessions := gm.getSessions()

	for s := range sessions {
		if !s.IsClosed() {
			return true
		}
	}

	return false
}

func (gm *EnemyDuelGame) handleEmojiMessage(s *session.Session, g *SessionGameStatus, msg *contract.C2SEnemyDuelEmojiMessage) {
	sessions := gm.getSessions()

	for session := range sessions {
		var playerID string
		if session == s {
			playerID = g.EnemyDuelGamePlayerStatus.PlayerID
		} else {
			playerID = g.EnemyDuelGamePlayerStatus.getExternalPlayerID()
		}
		session.SendMessage(contract.NewS2CEnemyDuelEmojiMessage(msg.EmojiGroup, msg.EmojiID, playerID))
	}
}

func (gm *EnemyDuelGame) handleBetMessage(s *session.Session, g *SessionGameStatus, msg *contract.C2SEnemyDuelBetMessage) {
	state := gm.state

	if _, ok := state.(*EnemyDuelGameBetState); !ok {
		return
	}

	g.EnemyDuelGamePlayerStatus.Side = msg.Side
	g.EnemyDuelGamePlayerStatus.AllIn = msg.AllIn

	sessions := gm.getSessions()

	for session := range sessions {
		var playerID string
		if session == s {
			playerID = g.EnemyDuelGamePlayerStatus.PlayerID
		} else {
			playerID = g.EnemyDuelGamePlayerStatus.getExternalPlayerID()
		}
		session.SendMessage(
			contract.NewS2CEnemyDuelClientStateMessageForBet(
				2, gm.round, state.GetForceExitTime(),
				playerID,
				g.EnemyDuelGamePlayerStatus.Side,
				g.EnemyDuelGamePlayerStatus.AllIn,
				g.EnemyDuelGamePlayerStatus.Streak,
			),
		)
	}
}

func (gm *EnemyDuelGame) getOtherPlayerIDSlice(internalPlayerID int) []string {
	otherPlayerIDSlice := []string(nil)

	maxNumPlayer := gm.getMaxNumPlayer()

	for i := 0; i < maxNumPlayer; i++ {
		if i == internalPlayerID {
			continue
		}

		otherPlayerIDSlice = append(otherPlayerIDSlice, getExternalPlayerID(i))
	}

	return otherPlayerIDSlice
}

func (gm *EnemyDuelGame) initPlayerStatus(g *SessionGameStatus, playerID string) {
	g.EnemyDuelGamePlayerStatus.PlayerID = playerID

	switch gm.ModeID {
	case "multiOperationMatch":
		g.EnemyDuelGamePlayerStatus.Money = 10000
		return
	}
	g.EnemyDuelGamePlayerStatus.Money = 1
	g.EnemyDuelGamePlayerStatus.ShieldState = 2
}

func newEmptyEnemyDuelBattleStatusRoundLeaderBoard(playerID string) *contract.EnemyDuelBattleStatusRoundLeaderBoard {
	return &contract.EnemyDuelBattleStatusRoundLeaderBoard{
		PlayerID: playerID,
	}
}

var roundMoneyMap = map[int]uint32{
	0: 2000,
	1: 2400,
	2: 3000,
	3: 4000,
	4: 5500,
	5: 8000,
	6: 12000,
	7: 18000,
	8: 30000,
	9: 50000,
}

func (gm *EnemyDuelGame) getRoundMoney() uint32 {
	if roundMoney, ok := roundMoneyMap[int(gm.round)]; ok {
		return roundMoney
	}

	return 50000
}

func (gm *EnemyDuelGame) getPlayerResult(s *session.Session, g *SessionGameStatus) *contract.EnemyDuelBattleStatusRoundLeaderBoard {
	if g.EnemyDuelGamePlayerStatus.ShieldState == 1 {
		g.EnemyDuelGamePlayerStatus.ShieldState = 0
	}

	if g.EnemyDuelGamePlayerStatus.Side == 3 {
		g.EnemyDuelGamePlayerStatus.Side = 0
	}

	won := (g.EnemyDuelGamePlayerStatus.Money > 0) && ((gm.reportSide & g.EnemyDuelGamePlayerStatus.Side) != 0)
	skip := g.EnemyDuelGamePlayerStatus.Side == 0

	leaderBoard := contract.EnemyDuelBattleStatusRoundLeaderBoard{
		PlayerID: g.EnemyDuelGamePlayerStatus.getExternalPlayerID(),
		OldMoney: g.EnemyDuelGamePlayerStatus.Money,
		Result:   gm.reportSide,
		Bet:      g.EnemyDuelGamePlayerStatus.Side,
	}

	switch gm.ModeID {
	case "multiOperationMatch":
		actualRoundMoney := min(gm.getRoundMoney(), g.EnemyDuelGamePlayerStatus.Money)
		if won {
			g.EnemyDuelGamePlayerStatus.Streak++

			if g.EnemyDuelGamePlayerStatus.AllIn != 0 {
				g.EnemyDuelGamePlayerStatus.Money += 2 * actualRoundMoney
			} else {
				g.EnemyDuelGamePlayerStatus.Money += actualRoundMoney
			}
		} else {
			g.EnemyDuelGamePlayerStatus.Streak = 0

			if !won && !skip {
				if g.EnemyDuelGamePlayerStatus.AllIn != 0 {
					g.EnemyDuelGamePlayerStatus.Money = 0
				} else {
					g.EnemyDuelGamePlayerStatus.Money -= actualRoundMoney
				}
			}
		}

		leaderBoard.NewMoney = g.EnemyDuelGamePlayerStatus.Money
		leaderBoard.Streak = g.EnemyDuelGamePlayerStatus.Streak
		if g.EnemyDuelGamePlayerStatus.Money > 0 {
			leaderBoard.MaxRound = gm.round + 1
		}

		return &leaderBoard
	}

	if gm.round < 5 {
		if !won && g.EnemyDuelGamePlayerStatus.ShieldState == 2 {
			won = true
			g.EnemyDuelGamePlayerStatus.ShieldState = 1
		}
	} else {
		g.EnemyDuelGamePlayerStatus.ShieldState = 0
	}

	if !won {
		g.EnemyDuelGamePlayerStatus.Money = 0
	}

	leaderBoard.NewMoney = g.EnemyDuelGamePlayerStatus.Money
	leaderBoard.ShieldState = g.EnemyDuelGamePlayerStatus.ShieldState
	if g.EnemyDuelGamePlayerStatus.Money > 0 {
		leaderBoard.MaxRound = gm.round + 1
	}

	return &leaderBoard
}

func (gm *EnemyDuelGame) getAllPlayerResult() map[string]*contract.EnemyDuelBattleStatusRoundLeaderBoard {
	sessions := gm.getSessions()

	allPlayerResult := make(map[string]*contract.EnemyDuelBattleStatusRoundLeaderBoard)

	for s, g := range sessions {
		playerResult := gm.getPlayerResult(s, g)
		allPlayerResult[playerResult.PlayerID] = playerResult
	}

	maxNumPlayer := gm.getMaxNumPlayer()

	for i := 0; i < maxNumPlayer; i++ {
		playerID := getExternalPlayerID(i)
		if _, ok := allPlayerResult[playerID]; !ok {
			allPlayerResult[playerID] = newEmptyEnemyDuelBattleStatusRoundLeaderBoard(playerID)
		}
	}

	return allPlayerResult
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

func CloseInactiveSession(sessions map[*session.Session]*SessionGameStatus) {
	currentTime := time.Now()
	tenSecAgo := currentTime.Add(-10 * time.Second)

	for s, g := range sessions {
		if g.LastActiveTime.Before(tenSecAgo) {
			s.Close()
		}
	}
}

func HandleSessionMessage(s *session.Session, g *SessionGameStatus, c contract.Content) {
	g.LastActiveTime = time.Now()

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
			defer s.SendMessage(contract.NewS2CEnemyDuelQuitMessage())
			log.Printf("failed to get modeID, stageID: %v", err)
			return
		}

		gameID := strings.Join([]string{msg.SceneID, modeID, stageID}, "|")

		game := getGame(gameID)
		if game == nil {
			defer s.SendMessage(contract.NewS2CEnemyDuelQuitMessage())
			log.Printf("game not found")
			return
		}

		err = game.AddSession(s, g)
		if err != nil {
			defer s.SendMessage(contract.NewS2CEnemyDuelQuitMessage())
			log.Printf("failed to add session to game: %v", err)
			return
		}

		game.initPlayerStatus(g, msg.PlayerID)

		otherPlayerIDSlice := game.getOtherPlayerIDSlice(g.EnemyDuelGamePlayerStatus.internalPlayerID)

		s.SendMessage(contract.NewS2CEnemyDuelJoinMessage(stageID, msg.PlayerID, otherPlayerIDSlice))

		return
	}

	if msg, ok := c.(*contract.C2SEnemyDuelRoundSettleMessage); ok {
		g.EnemyDuelGamePlayerStatus.ReportSide = msg.Side

		return
	}

	if msg, ok := c.(*contract.C2SEnemyDuelEmojiMessage); ok {
		g.EnemyDuel.handleEmojiMessage(s, g, msg)

		return
	}

	if msg, ok := c.(*contract.C2SEnemyDuelBetMessage); ok {
		g.EnemyDuel.handleBetMessage(s, g, msg)

		return
	}
}
