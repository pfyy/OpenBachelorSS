package contract

import (
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

func NewS2CEnemyDuelJoinMessage() *S2CEnemyDuelJoinMessage {
	return &S2CEnemyDuelJoinMessage{}
}

func NewS2CEnemyDuelClientStateMessage(state uint8, round uint8) *S2CEnemyDuelClientStateMessage {
	return &S2CEnemyDuelClientStateMessage{State: state, Round: round}
}

func NewC2SEnemyDuelFinalSettleMessage() *C2SEnemyDuelFinalSettleMessage {
	return &C2SEnemyDuelFinalSettleMessage{}
}
