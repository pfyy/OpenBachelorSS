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

func getNickName(isSelf bool, externalPlayerID string) string {
	var pureNickName string
	if isSelf {
		pureNickName = "Bachelor"
	} else {
		pureNickName = "Undergraduate"
	}

	return fmt.Sprintf("%s#%04s", pureNickName, externalPlayerID)
}

func NewS2CEnemyDuelJoinMessage(stageID string, playerID string, externalPlayerID string, otherPlayerIDSlice []string, seed uint32) *S2CEnemyDuelJoinMessage {
	players := []*EnemyDuelServicePlayer{
		{
			PlayerID: playerID, AvatarID: "avatar_def_01", NickName: getNickName(true, externalPlayerID), AvatarType: "ICON",
		},
	}

	for _, otherPlayerID := range otherPlayerIDSlice {
		players = append(players, &EnemyDuelServicePlayer{PlayerID: otherPlayerID, AvatarID: "avatar_def_01", NickName: getNickName(false, otherPlayerID), AvatarType: "ICON"})
	}

	currentTime := time.Now()

	return &S2CEnemyDuelJoinMessage{
		StageID:   stageID,
		NowTs:     uint64(currentTime.Unix()),
		Players:   players,
		StageSeed: seed,
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
		Duration: 300,
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

func NewS2CEnemyDuelClientStateMessageForSettle(state uint8, round uint8, forceEndTs time.Time, allPlayerResult map[string]*EnemyDuelBattleStatusRoundLeaderBoard, playerID string, externalPlayerID string) *S2CEnemyDuelClientStateMessage {
	msg := NewS2CEnemyDuelClientStateMessage(state, round, forceEndTs)

	var leaderBoard []*EnemyDuelBattleStatusRoundLeaderBoard

	for k, v := range allPlayerResult {
		if k == externalPlayerID {
			nv := *v
			nv.PlayerID = playerID
			leaderBoard = append(leaderBoard, &nv)
		} else {
			leaderBoard = append(leaderBoard, v)
		}
	}

	msg.LeaderBoard = leaderBoard

	return msg
}

func NewS2CEnemyDuelClientStateMessageForEntry(state uint8, round uint8, forceEndTs time.Time, seed uint32) *S2CEnemyDuelClientStateMessage {
	msg := NewS2CEnemyDuelClientStateMessage(state, round, forceEndTs)
	msg.SrcEntryData = []*EnemyDuelBattleStatusEntryData{{Seed: seed}}
	return msg
}
