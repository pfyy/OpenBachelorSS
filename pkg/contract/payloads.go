package contract

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/OpenBachelor/OpenBachelorSS/pkg/protocol"
)

var messageRegistry = make(map[uint32]func() Content)

func init() {
	RegisterMessage(S2CEnemyDuelEmojiMessageType, func() Content { return &S2CEnemyDuelEmojiMessage{} })
	RegisterMessage(S2CEnemyDuelHeartBeatMessageType, func() Content { return &S2CEnemyDuelHeartBeatMessage{} })
	RegisterMessage(S2CEnemyDuelQuitMessageType, func() Content { return &S2CEnemyDuelQuitMessage{} })
	RegisterMessage(S2CEnemyDuelStepMessageType, func() Content { return &S2CEnemyDuelStepMessage{} })
	RegisterMessage(S2CEnemyDuelClientStateMessageType, func() Content { return &S2CEnemyDuelClientStateMessage{} })
	RegisterMessage(S2CEnemyDuelJoinMessageType, func() Content { return &S2CEnemyDuelJoinMessage{} })

	RegisterMessage(C2SEnemyDuelEmojiMessageType, func() Content { return &C2SEnemyDuelEmojiMessage{} })
	RegisterMessage(C2SEnemyDuelReadyMessageType, func() Content { return &C2SEnemyDuelReadyMessage{} })
	RegisterMessage(C2SEnemyDuelJoinMessageType, func() Content { return &C2SEnemyDuelJoinMessage{} })
	RegisterMessage(C2SEnemyDuelRoundSettleMessageType, func() Content { return &C2SEnemyDuelRoundSettleMessage{} })
	RegisterMessage(C2SEnemyDuelBetMessageType, func() Content { return &C2SEnemyDuelBetMessage{} })
	RegisterMessage(C2SEnemyDuelHistoryMessageType, func() Content { return &C2SEnemyDuelHistoryMessage{} })
	RegisterMessage(C2SEnemyDuelQuitMessageType, func() Content { return &C2SEnemyDuelQuitMessage{} })
	RegisterMessage(C2SEnemyDuelHeartBeatMessageType, func() Content { return &C2SEnemyDuelHeartBeatMessage{} })
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
		return fmt.Errorf("string too long (%d bytes)", len(s))
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

func SerializeSlice[T encoding.BinaryMarshaler](items []T) ([]byte, error) {
	if len(items) > 0xffff {
		return nil, fmt.Errorf("slice too long (%d items)", len(items))
	}

	var buf bytes.Buffer

	err := binary.Write(&buf, binary.BigEndian, uint16(len(items)))
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		itemBytes, err := item.MarshalBinary()
		if err != nil {
			return nil, err
		}

		_, err = buf.Write(itemBytes)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil

}

type FixedSize interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~int8 | ~int16 | ~int32 | ~int64 |
		~float32 | ~float64
}

func SerializePrimitiveSlice[T FixedSize](items []T) ([]byte, error) {
	if len(items) > 0xffff {
		return nil, fmt.Errorf("slice too long (%d items)", len(items))
	}

	var buf bytes.Buffer

	err := binary.Write(&buf, binary.BigEndian, uint16(len(items)))
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, items)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil

}

// ------------------------------

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

type S2CEnemyDuelHeartBeatMessage struct {
	Seq  uint32
	Time uint64
}

func (m *S2CEnemyDuelHeartBeatMessage) ContentType() uint32 {
	return S2CEnemyDuelHeartBeatMessageType
}

func (m *S2CEnemyDuelHeartBeatMessage) Marshal() ([]byte, error) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, m.Seq); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, m.Time); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *S2CEnemyDuelHeartBeatMessage) Unmarshal(payload []byte) error {
	panic("TODO")
}

type S2CEnemyDuelQuitMessage struct {
}

func (m *S2CEnemyDuelQuitMessage) ContentType() uint32 {
	return S2CEnemyDuelQuitMessageType
}

func (m *S2CEnemyDuelQuitMessage) Marshal() ([]byte, error) {

	return nil, nil
}

func (m *S2CEnemyDuelQuitMessage) Unmarshal(payload []byte) error {
	panic("TODO")
}

type S2CEnemyDuelStepMessage struct {
	Index    uint32
	Duration uint32
	CheckSeq int32
	Round    uint8
}

func (m *S2CEnemyDuelStepMessage) ContentType() uint32 {
	return S2CEnemyDuelStepMessageType
}

func (m *S2CEnemyDuelStepMessage) Marshal() ([]byte, error) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, m.Index); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, m.Duration); err != nil {
		return nil, err
	}

	if _, err := buf.Write([]byte{0, 0}); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, m.CheckSeq); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, m.Round); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *S2CEnemyDuelStepMessage) Unmarshal(payload []byte) error {
	panic("TODO")
}

type EnemyDuelBattleStatusEntryData struct {
	Seed        uint32
	SeedHistory []uint32
	SideHistory []uint8
}

func (t *EnemyDuelBattleStatusEntryData) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, t.Seed); err != nil {
		return nil, err
	}

	seedHistory, err := SerializePrimitiveSlice(t.SeedHistory)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(seedHistory)
	if err != nil {
		return nil, err
	}

	sideHistory, err := SerializePrimitiveSlice(t.SideHistory)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(sideHistory)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type EnemyDuelBattleStatusBetItem struct {
	PlayerID string
	Side     uint8
	AllIn    uint8
	Streak   uint8
	UpdateTs uint64
}

func (t *EnemyDuelBattleStatusBetItem) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	if err := writePrefixedString(&buf, t.PlayerID); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, t.Side); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, t.AllIn); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, t.Streak); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, t.UpdateTs); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type EnemyDuelBattleStatusRoundLeaderBoard struct {
	PlayerID    string
	OldMoney    uint32
	NewMoney    uint32
	MaxRound    uint8
	Streak      uint8
	Result      uint8
	Bet         uint8
	ShieldState uint8
}

func (t *EnemyDuelBattleStatusRoundLeaderBoard) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	if err := writePrefixedString(&buf, t.PlayerID); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, t.OldMoney); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, t.NewMoney); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, t.MaxRound); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, t.Streak); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, t.Result); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, t.Bet); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, t.ShieldState); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type S2CEnemyDuelClientStateMessage struct {
	State        uint8
	Round        uint8
	ForceEndTs   uint64
	SrcEntryData []*EnemyDuelBattleStatusEntryData
	BetList      []*EnemyDuelBattleStatusBetItem
	LeaderBoard  []*EnemyDuelBattleStatusRoundLeaderBoard
}

func (m *S2CEnemyDuelClientStateMessage) ContentType() uint32 {
	return S2CEnemyDuelClientStateMessageType
}

func (m *S2CEnemyDuelClientStateMessage) Marshal() ([]byte, error) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, m.State); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, m.Round); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, m.ForceEndTs); err != nil {
		return nil, err
	}

	srcEntryData, err := SerializeSlice(m.SrcEntryData)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(srcEntryData)
	if err != nil {
		return nil, err
	}

	betList, err := SerializeSlice(m.BetList)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(betList)
	if err != nil {
		return nil, err
	}

	leaderBoard, err := SerializeSlice(m.LeaderBoard)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(leaderBoard)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *S2CEnemyDuelClientStateMessage) Unmarshal(payload []byte) error {
	panic("TODO")
}

type S2CEnemyDuelJoinMessage struct {
}

func (m *S2CEnemyDuelJoinMessage) ContentType() uint32 {
	return S2CEnemyDuelJoinMessageType
}

func (m *S2CEnemyDuelJoinMessage) Marshal() ([]byte, error) {
	var buf bytes.Buffer

	return buf.Bytes(), nil
}

func (m *S2CEnemyDuelJoinMessage) Unmarshal(payload []byte) error {
	panic("TODO")
}

// ------------------------------

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

type C2SEnemyDuelReadyMessage struct{}

func (m *C2SEnemyDuelReadyMessage) ContentType() uint32 {
	return C2SEnemyDuelReadyMessageType
}

func (m *C2SEnemyDuelReadyMessage) Marshal() ([]byte, error) {
	panic("TODO")
}

func (m *C2SEnemyDuelReadyMessage) Unmarshal(payload []byte) error {
	return nil
}

type C2SEnemyDuelJoinMessage struct {
	PlayerID string
	SceneID  string
	Token    string
}

func (m *C2SEnemyDuelJoinMessage) ContentType() uint32 {
	return C2SEnemyDuelJoinMessageType
}

func (m *C2SEnemyDuelJoinMessage) Marshal() ([]byte, error) {
	panic("TODO")
}

func (m *C2SEnemyDuelJoinMessage) Unmarshal(payload []byte) error {
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

type C2SEnemyDuelHistoryMessage struct {
	Seq uint32
}

func (m *C2SEnemyDuelHistoryMessage) ContentType() uint32 {
	return C2SEnemyDuelHistoryMessageType
}

func (m *C2SEnemyDuelHistoryMessage) Marshal() ([]byte, error) {
	panic("TODO")
}

func (m *C2SEnemyDuelHistoryMessage) Unmarshal(payload []byte) error {
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

type C2SEnemyDuelHeartBeatMessage struct {
	Seq  uint32
	Time uint64
}

func (m *C2SEnemyDuelHeartBeatMessage) ContentType() uint32 {
	return C2SEnemyDuelHeartBeatMessageType
}

func (m *C2SEnemyDuelHeartBeatMessage) Marshal() ([]byte, error) {
	panic("TODO")
}

func (m *C2SEnemyDuelHeartBeatMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)

	var err error

	err = binary.Read(reader, binary.BigEndian, &m.Seq)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.Time)
	if err != nil {
		return err
	}

	return nil
}
