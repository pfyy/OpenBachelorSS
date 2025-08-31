package contract

type S2CEnemyDuelEmojiMessage struct{}

func (m *S2CEnemyDuelEmojiMessage) ContentType() uint32 {
	return S2CEnemyDuelEmojiMessageType
}

func (m *S2CEnemyDuelEmojiMessage) Marshal() ([]byte, error) {
	panic("TODO")
}

func (m *S2CEnemyDuelEmojiMessage) Unmarshal(payload []byte) error {
	panic("TODO")
}

type C2SEnemyDuelEmojiMessage struct{}

func (m *C2SEnemyDuelEmojiMessage) ContentType() uint32 {
	return C2SEnemyDuelEmojiMessageType
}

func (m *C2SEnemyDuelEmojiMessage) Marshal() ([]byte, error) {
	panic("TODO")
}

func (m *C2SEnemyDuelEmojiMessage) Unmarshal(payload []byte) error {
	panic("TODO")
}
