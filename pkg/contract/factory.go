package contract

func NewS2CEnemyDuelHeartBeatMessage(seq uint32, time uint64) *S2CEnemyDuelHeartBeatMessage {
	return &S2CEnemyDuelHeartBeatMessage{
		Seq:  seq,
		Time: time,
	}
}
