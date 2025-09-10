package contract

import (
	"fmt"
	"time"

	"github.com/OpenBachelor/OpenBachelorSS/internal/config"
)

func NewS2CEnemyDuelHeartBeatMessage(seq uint32, time uint64) *S2CEnemyDuelHeartBeatMessage {
	return &S2CEnemyDuelHeartBeatMessage{
		Seq:  seq,
		Time: time,
	}
}

func NewS2CEnemyDuelKickMessage() *S2CEnemyDuelKickMessage {
	return &S2CEnemyDuelKickMessage{}
}

func NewS2CEnemyDuelTeamJoinMessage() *S2CEnemyDuelTeamJoinMessage {
	return &S2CEnemyDuelTeamJoinMessage{}
}

func NewS2CEnemyDuelTeamStatusMessage(sceneID string, token string) *S2CEnemyDuelTeamStatusMessage {
	cfg := config.Get()
	return &S2CEnemyDuelTeamStatusMessage{State: 3, Address: cfg.Server.Addr, SceneID: sceneID, Token: token}
}

func NewS2CEnemyDuelEndMessage() *S2CEnemyDuelEndMessage {
	return &S2CEnemyDuelEndMessage{}
}

func NewS2CEnemyDuelJoinMessage(stageID string, playerID string, otherPlayerIDSlice []string) *S2CEnemyDuelJoinMessage {
	players := []*EnemyDuelServicePlayer{
		{
			PlayerID: playerID, AvatarID: "avatar_def_01", NickName: "Bachelor#1234", AvatarType: "ICON",
		},
	}

	for i, otherPlayerID := range otherPlayerIDSlice {
		players = append(players, &EnemyDuelServicePlayer{PlayerID: otherPlayerID, AvatarID: "avatar_def_01", NickName: fmt.Sprintf("Undergraduate#%04d", i), AvatarType: "ICON"})
	}

	return &S2CEnemyDuelJoinMessage{
		StageID: stageID,
		Players: players,
	}
}

func NewS2CEnemyDuelClientStateMessage(state uint8, round uint8, forceEndTs time.Time) *S2CEnemyDuelClientStateMessage {
	return &S2CEnemyDuelClientStateMessage{State: state, Round: round, ForceEndTs: uint64(forceEndTs.Unix())}
}

func NewC2SEnemyDuelFinalSettleMessage() *C2SEnemyDuelFinalSettleMessage {
	return &C2SEnemyDuelFinalSettleMessage{}
}

func NewS2CEnemyDuelStepMessage(step uint32, round uint8) *S2CEnemyDuelStepMessage {
	return &S2CEnemyDuelStepMessage{
		Index:    step,
		Duration: 100,
		CheckSeq: -1,
		Round:    round,
	}
}

func NewS2CEnemyDuelQuitMessage() *S2CEnemyDuelQuitMessage {
	return &S2CEnemyDuelQuitMessage{}
}

func NewS2CEnemyDuelEmojiMessage(emojiGroup, emojiID string, playerID string) *S2CEnemyDuelEmojiMessage {
	return &S2CEnemyDuelEmojiMessage{
		PlayerID:   playerID,
		EmojiGroup: emojiGroup,
		EmojiID:    emojiID,
	}
}

func NewS2CEnemyDuelClientStateMessageForBet(state uint8, round uint8, forceEndTs time.Time, playerID string, side uint8, allIn uint8, streak uint8) *S2CEnemyDuelClientStateMessage {
	msg := NewS2CEnemyDuelClientStateMessage(state, round, forceEndTs)
	msg.BetList = []*EnemyDuelBattleStatusBetItem{{PlayerID: playerID, Side: side, AllIn: allIn, Streak: streak}}
	return msg
}
