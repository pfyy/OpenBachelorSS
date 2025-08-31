package contract

import (
	"fmt"

	"github.com/OpenBachelor/OpenBachelorSS/pkg/protocol"
)

func FromEnvelop(env *protocol.Envelop) (Content, error) {
	var c Content
	switch env.Type {
	case S2CEnemyDuelEmojiMessageType:
		c = &S2CEnemyDuelEmojiMessage{}
	case C2SEnemyDuelEmojiMessageType:
		c = &C2SEnemyDuelEmojiMessage{}
	default:
		return nil, fmt.Errorf("unknown envelop")
	}

	err := c.Unmarshal(env.Payload)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func ToEnvelop(c Content) (*protocol.Envelop, error) {
	payload, err := c.Marshal()
	if err != nil {
		return nil, err
	}
	return &protocol.Envelop{
		Type:    c.ContentType(),
		Payload: payload,
	}, nil
}

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
