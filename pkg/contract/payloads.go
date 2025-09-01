package contract

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/OpenBachelor/OpenBachelorSS/pkg/protocol"
)

var messageRegistry = make(map[uint32]func() Content)

func init() {
	RegisterMessage(S2CEnemyDuelEmojiMessageType, func() Content { return &S2CEnemyDuelEmojiMessage{} })
	RegisterMessage(C2SEnemyDuelEmojiMessageType, func() Content { return &C2SEnemyDuelEmojiMessage{} })
	RegisterMessage(C2SEnemyDuelBattleReadyMessageType, func() Content { return &C2SEnemyDuelBattleReadyMessage{} })
	RegisterMessage(C2SEnemyDuelSceneJoinMessageType, func() Content { return &C2SEnemyDuelSceneJoinMessage{} })
	RegisterMessage(C2SEnemyDuelRoundSettleMessageType, func() Content { return &C2SEnemyDuelRoundSettleMessage{} })
	RegisterMessage(C2SEnemyDuelBetMessageType, func() Content { return &C2SEnemyDuelBetMessage{} })
	RegisterMessage(C2SEnemyDuelSceneHistoryMessageType, func() Content { return &C2SEnemyDuelSceneHistoryMessage{} })
	RegisterMessage(C2SEnemyDuelQuitMessageType, func() Content { return &C2SEnemyDuelQuitMessage{} })
}

func RegisterMessage(msgType uint32, constructor func() Content) {
	if _, exists := messageRegistry[msgType]; exists {
		panic(fmt.Sprintf("message type %d is already registered", msgType))
	}
	messageRegistry[msgType] = constructor
}

func FromEnvelop(env *protocol.Envelop) (Content, error) {
	constructor, ok := messageRegistry[env.Type]

	if !ok {
		return nil, fmt.Errorf("unknown envelop")
	}

	c := constructor()

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

func WriteContent(w io.Writer, c Content) error {
	env, err := ToEnvelop(c)
	if err != nil {
		return err
	}
	err = protocol.WriteEnvelop(w, env)
	if err != nil {
		return err
	}
	return nil
}

func ReadContent(r io.Reader) (Content, error) {
	env, err := protocol.ReadEnvelop(r)
	if err != nil {
		return nil, err
	}
	c, err := FromEnvelop(env)
	if err != nil {
		return nil, err
	}
	return c, nil
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

func writePrefixedString(w io.Writer, s string) error {
	if len(s) > 0xffff {
		return fmt.Errorf("string to long (%d bytes)", len(s))
	}

	length := uint16(len(s))

	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return err
	}

	if _, err := w.Write([]byte(s)); err != nil {
		return err
	}

	return nil
}

type S2CEnemyDuelEmojiMessage struct {
	PlayerID   string
	EmojiGroup string
	EmojiID    string
}

func (m *S2CEnemyDuelEmojiMessage) ContentType() uint32 {
	return S2CEnemyDuelEmojiMessageType
}

func (m *S2CEnemyDuelEmojiMessage) Marshal() ([]byte, error) {
	var buf bytes.Buffer

	if err := writePrefixedString(&buf, m.PlayerID); err != nil {
		return nil, err
	}

	if err := writePrefixedString(&buf, m.EmojiGroup); err != nil {
		return nil, err
	}

	if err := writePrefixedString(&buf, m.EmojiID); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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

type C2SEnemyDuelBattleReadyMessage struct{}

func (m *C2SEnemyDuelBattleReadyMessage) ContentType() uint32 {
	return C2SEnemyDuelBattleReadyMessageType
}

func (m *C2SEnemyDuelBattleReadyMessage) Marshal() ([]byte, error) {
	panic("TODO")
}

func (m *C2SEnemyDuelBattleReadyMessage) Unmarshal(payload []byte) error {
	return nil
}

type C2SEnemyDuelSceneJoinMessage struct {
	PlayerID string
	SceneID  string
	Token    string
}

func (m *C2SEnemyDuelSceneJoinMessage) ContentType() uint32 {
	return C2SEnemyDuelSceneJoinMessageType
}

func (m *C2SEnemyDuelSceneJoinMessage) Marshal() ([]byte, error) {
	panic("TODO")
}

func (m *C2SEnemyDuelSceneJoinMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)

	var err error

	m.PlayerID, err = readPrefixedString(reader)
	if err != nil {
		return err
	}

	m.SceneID, err = readPrefixedString(reader)
	if err != nil {
		return err
	}

	m.Token, err = readPrefixedString(reader)
	if err != nil {
		return err
	}

	return nil
}

type C2SEnemyDuelRoundSettleMessage struct {
	Side     uint8
	Info     string
	InfoHash string
}

func (m *C2SEnemyDuelRoundSettleMessage) ContentType() uint32 {
	return C2SEnemyDuelRoundSettleMessageType
}

func (m *C2SEnemyDuelRoundSettleMessage) Marshal() ([]byte, error) {
	panic("TODO")
}

func (m *C2SEnemyDuelRoundSettleMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)

	var err error

	err = binary.Read(reader, binary.BigEndian, &m.Side)
	if err != nil {
		return err
	}

	m.Info, err = readPrefixedString(reader)
	if err != nil {
		return err
	}

	m.InfoHash, err = readPrefixedString(reader)
	if err != nil {
		return err
	}

	return nil
}

type C2SEnemyDuelBetMessage struct {
	PlayerID string
	Side     uint8
	IsPlayer uint8
	AllIn    uint8
}

func (m *C2SEnemyDuelBetMessage) ContentType() uint32 {
	return C2SEnemyDuelBetMessageType
}

func (m *C2SEnemyDuelBetMessage) Marshal() ([]byte, error) {
	panic("TODO")
}

func (m *C2SEnemyDuelBetMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)

	var err error

	m.PlayerID, err = readPrefixedString(reader)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.Side)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.IsPlayer)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.AllIn)
	if err != nil {
		return err
	}

	return nil
}

type C2SEnemyDuelSceneHistoryMessage struct {
	Seq uint32
}

func (m *C2SEnemyDuelSceneHistoryMessage) ContentType() uint32 {
	return C2SEnemyDuelSceneHistoryMessageType
}

func (m *C2SEnemyDuelSceneHistoryMessage) Marshal() ([]byte, error) {
	panic("TODO")
}

func (m *C2SEnemyDuelSceneHistoryMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)

	var err error

	err = binary.Read(reader, binary.BigEndian, &m.Seq)
	if err != nil {
		return err
	}

	return nil
}

type C2SEnemyDuelQuitMessage struct{}

func (m *C2SEnemyDuelQuitMessage) ContentType() uint32 {
	return C2SEnemyDuelQuitMessageType
}

func (m *C2SEnemyDuelQuitMessage) Marshal() ([]byte, error) {
	panic("TODO")
}

func (m *C2SEnemyDuelQuitMessage) Unmarshal(payload []byte) error {
	return nil
}
