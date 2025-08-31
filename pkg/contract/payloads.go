package contract

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

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

func readPrefixedString(r io.Reader) (string, error) {
	var length uint16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", err
	}

	stringBytes := make([]byte, length)
	if _, err := io.ReadFull(r, stringBytes); err != nil {
		return "", err
	}
	return string(stringBytes), nil
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

type C2SEnemyDuelEmojiMessage struct {
	EmojiGroup string
	EmojiID    string
}

func (m *C2SEnemyDuelEmojiMessage) ContentType() uint32 {
	return C2SEnemyDuelEmojiMessageType
}

func (m *C2SEnemyDuelEmojiMessage) Marshal() ([]byte, error) {
	panic("TODO")
}

func (m *C2SEnemyDuelEmojiMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)

	var err error

	m.EmojiGroup, err = readPrefixedString(reader)
	if err != nil {
		return err
	}

	m.EmojiID, err = readPrefixedString(reader)
	if err != nil {
		return err
	}

	return nil
}
