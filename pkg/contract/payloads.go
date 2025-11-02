package contract

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"slices"

	"github.com/OpenBachelor/OpenBachelorSS/pkg/protocol"
)

var enemyDuelMessageRegistry = make(map[uint32]func() Content)

func init() {
	RegisterMessage(EnemyDuelMessageDomain, S2CEnemyDuelEmojiMessageType, func() Content { return &S2CEnemyDuelEmojiMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, S2CEnemyDuelHeartBeatMessageType, func() Content { return &S2CEnemyDuelHeartBeatMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, S2CEnemyDuelQuitMessageType, func() Content { return &S2CEnemyDuelQuitMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, S2CEnemyDuelStepMessageType, func() Content { return &S2CEnemyDuelStepMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, S2CEnemyDuelClientStateMessageType, func() Content { return &S2CEnemyDuelClientStateMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, S2CEnemyDuelJoinMessageType, func() Content { return &S2CEnemyDuelJoinMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, S2CEnemyDuelEndMessageType, func() Content { return &S2CEnemyDuelEndMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, S2CEnemyDuelHistoryMessageType, func() Content { return &S2CEnemyDuelHistoryMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, S2CEnemyDuelTeamJoinMessageType, func() Content { return &S2CEnemyDuelTeamJoinMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, S2CEnemyDuelKickMessageType, func() Content { return &S2CEnemyDuelKickMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, S2CEnemyDuelTeamStatusMessageType, func() Content { return &S2CEnemyDuelTeamStatusMessage{} })

	RegisterMessage(EnemyDuelMessageDomain, C2SEnemyDuelEmojiMessageType, func() Content { return &C2SEnemyDuelEmojiMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, C2SEnemyDuelReadyMessageType, func() Content { return &C2SEnemyDuelReadyMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, C2SEnemyDuelJoinMessageType, func() Content { return &C2SEnemyDuelJoinMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, C2SEnemyDuelRoundSettleMessageType, func() Content { return &C2SEnemyDuelRoundSettleMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, C2SEnemyDuelBetMessageType, func() Content { return &C2SEnemyDuelBetMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, C2SEnemyDuelHistoryMessageType, func() Content { return &C2SEnemyDuelHistoryMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, C2SEnemyDuelQuitMessageType, func() Content { return &C2SEnemyDuelQuitMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, C2SEnemyDuelHeartBeatMessageType, func() Content { return &C2SEnemyDuelHeartBeatMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, C2SEnemyDuelFinalSettleMessageType, func() Content { return &C2SEnemyDuelFinalSettleMessage{} })
	RegisterMessage(EnemyDuelMessageDomain, C2SEnemyDuelTeamJoinMessageType, func() Content { return &C2SEnemyDuelTeamJoinMessage{} })
}

type UnknownMessage struct {
	Envelope protocol.Envelope
}

func (m *UnknownMessage) ContentType() uint32 {
	return m.Envelope.Type
}

func (m *UnknownMessage) Marshal() ([]byte, error) {
	return slices.Clone(m.Envelope.Payload), nil
}

func (m *UnknownMessage) Unmarshal(payload []byte) error {
	m.Envelope.Payload = slices.Clone(payload)
	return nil
}

func NewUnknownMessage(env *protocol.Envelope) *UnknownMessage {
	return &UnknownMessage{
		Envelope: protocol.Envelope{
			Type:    env.Type,
			Payload: slices.Clone(env.Payload),
		},
	}
}

func getMessageRegistry(msgDomain MessageDomain) map[uint32]func() Content {
	switch msgDomain {
	case EnemyDuelMessageDomain:
		return enemyDuelMessageRegistry
	case IceBreakerDomain:
		return iceBreakerMessageRegistry
	}
	return nil
}

func RegisterMessage(msgDomain MessageDomain, msgType uint32, constructor func() Content) {
	messageRegistry := getMessageRegistry(msgDomain)
	if _, exists := messageRegistry[msgType]; exists {
		panic(fmt.Sprintf("message type %d is already registered", msgType))
	}
	messageRegistry[msgType] = constructor
}

func FromEnvelope(msgDomain MessageDomain, env *protocol.Envelope) (Content, error) {
	messageRegistry := getMessageRegistry(msgDomain)

	constructor, ok := messageRegistry[env.Type]

	if !ok {
		return NewUnknownMessage(env), nil
	}

	c := constructor()

	err := c.Unmarshal(env.Payload)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func ToEnvelope(c Content) (*protocol.Envelope, error) {
	payload, err := c.Marshal()
	if err != nil {
		return nil, err
	}
	return &protocol.Envelope{
		Type:    c.ContentType(),
		Payload: payload,
	}, nil
}

func WriteContent(w io.Writer, c Content) error {
	env, err := ToEnvelope(c)
	if err != nil {
		return err
	}
	err = protocol.WriteEnvelope(w, env)
	if err != nil {
		return err
	}
	return nil
}

func ReadContent(msgDomain MessageDomain, r io.Reader) (Content, error) {
	env, err := protocol.ReadEnvelope(r)
	if err != nil {
		return nil, err
	}
	c, err := FromEnvelope(msgDomain, env)
	if err != nil {
		return nil, err
	}
	return c, nil
}

const defaultMaxStrSize = 1 << 10

func readPrefixedString(r io.Reader, maxStrSize uint16) (string, error) {
	var length uint16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", err
	}

	if length > maxStrSize {
		return "", fmt.Errorf("string too long (%d)", length)
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

type BinaryUnmarshalerReader interface {
	UnmarshalBinaryReader(r io.Reader) error
}

const defaultMaxSliceSize = 128

func DeserializeSlice[T any, PT interface {
	*T
	BinaryUnmarshalerReader
}](r io.Reader, maxSliceSize uint16) ([]PT, error) {
	var length uint16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	if length > maxSliceSize {
		return nil, fmt.Errorf("slice too long (%d)", length)
	}

	items := make([]PT, length)

	for i := 0; i < int(length); i++ {
		item := PT(new(T))

		if err := item.UnmarshalBinaryReader(r); err != nil {
			return nil, err
		}

		items[i] = item
	}

	return items, nil
}

func DeserializePrimitiveSlice[T FixedSize](r io.Reader, maxSliceSize uint16) ([]T, error) {
	var length uint16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	if length > maxSliceSize {
		return nil, fmt.Errorf("slice too long (%d)", length)
	}

	items := make([]T, length)

	if err := binary.Read(r, binary.BigEndian, items); err != nil {
		return nil, err
	}

	return items, nil

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
	var err error

	err = writePrefixedString(&buf, m.PlayerID)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.EmojiGroup)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.EmojiID)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *S2CEnemyDuelEmojiMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)
	var err error

	m.PlayerID, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	m.EmojiGroup, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	m.EmojiID, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	return nil
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
	var err error

	err = binary.Write(&buf, binary.BigEndian, m.Seq)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.Time)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *S2CEnemyDuelHeartBeatMessage) Unmarshal(payload []byte) error {
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

type S2CEnemyDuelQuitMessage struct {
}

func (m *S2CEnemyDuelQuitMessage) ContentType() uint32 {
	return S2CEnemyDuelQuitMessageType
}

func (m *S2CEnemyDuelQuitMessage) Marshal() ([]byte, error) {

	return nil, nil
}

func (m *S2CEnemyDuelQuitMessage) Unmarshal(payload []byte) error {
	return nil
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
	var err error

	err = binary.Write(&buf, binary.BigEndian, m.Index)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.Duration)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write([]byte{0, 0})
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.CheckSeq)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.Round)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *S2CEnemyDuelStepMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)
	var err error

	err = binary.Read(reader, binary.BigEndian, &m.Index)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.Duration)
	if err != nil {
		return err
	}

	_, err = io.CopyN(io.Discard, reader, 2)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.CheckSeq)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.Round)
	if err != nil {
		return err
	}

	return nil
}

type EnemyDuelBattleStatusEntryData struct {
	Seed        uint32
	SeedHistory []uint32
	SideHistory []uint8
}

func (t *EnemyDuelBattleStatusEntryData) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	var err error

	err = binary.Write(&buf, binary.BigEndian, t.Seed)
	if err != nil {
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

func (t *EnemyDuelBattleStatusEntryData) UnmarshalBinaryReader(r io.Reader) error {
	var err error

	err = binary.Read(r, binary.BigEndian, &t.Seed)
	if err != nil {
		return err
	}

	t.SeedHistory, err = DeserializePrimitiveSlice[uint32](r, defaultMaxSliceSize)
	if err != nil {
		return err
	}

	t.SideHistory, err = DeserializePrimitiveSlice[uint8](r, defaultMaxSliceSize)
	if err != nil {
		return err
	}

	return nil
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
	var err error

	err = writePrefixedString(&buf, t.PlayerID)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.Side)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.AllIn)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.Streak)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.UpdateTs)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (t *EnemyDuelBattleStatusBetItem) UnmarshalBinaryReader(r io.Reader) error {
	var err error

	t.PlayerID, err = readPrefixedString(r, defaultMaxStrSize)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.Side)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.AllIn)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.Streak)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.UpdateTs)
	if err != nil {
		return err
	}

	return nil
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
	var err error

	err = writePrefixedString(&buf, t.PlayerID)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.OldMoney)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.NewMoney)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.MaxRound)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.Streak)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.Result)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.Bet)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.ShieldState)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (t *EnemyDuelBattleStatusRoundLeaderBoard) UnmarshalBinaryReader(r io.Reader) error {
	var err error

	t.PlayerID, err = readPrefixedString(r, defaultMaxStrSize)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.OldMoney)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.NewMoney)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.MaxRound)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.Streak)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.Result)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.Bet)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.ShieldState)
	if err != nil {
		return err
	}

	return nil
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
	var err error

	err = binary.Write(&buf, binary.BigEndian, m.State)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.Round)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.ForceEndTs)
	if err != nil {
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
	reader := bytes.NewReader(payload)
	var err error

	err = binary.Read(reader, binary.BigEndian, &m.State)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.Round)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.ForceEndTs)
	if err != nil {
		return err
	}

	m.SrcEntryData, err = DeserializeSlice[EnemyDuelBattleStatusEntryData, *EnemyDuelBattleStatusEntryData](reader, defaultMaxSliceSize)
	if err != nil {
		return err
	}

	m.BetList, err = DeserializeSlice[EnemyDuelBattleStatusBetItem, *EnemyDuelBattleStatusBetItem](reader, defaultMaxSliceSize)
	if err != nil {
		return err
	}

	m.LeaderBoard, err = DeserializeSlice[EnemyDuelBattleStatusRoundLeaderBoard, *EnemyDuelBattleStatusRoundLeaderBoard](reader, defaultMaxSliceSize)
	if err != nil {
		return err
	}

	return nil
}

type EnemyDuelServicePlayer struct {
	PlayerID   string
	AvatarID   string
	NickName   string
	AvatarType string
	HaveShield uint8
}

func (t *EnemyDuelServicePlayer) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	var err error

	err = writePrefixedString(&buf, t.PlayerID)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, t.AvatarID)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, t.NickName)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, t.AvatarType)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.HaveShield)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (t *EnemyDuelServicePlayer) UnmarshalBinaryReader(r io.Reader) error {
	var err error

	t.PlayerID, err = readPrefixedString(r, defaultMaxStrSize)
	if err != nil {
		return err
	}

	t.AvatarID, err = readPrefixedString(r, defaultMaxStrSize)
	if err != nil {
		return err
	}

	t.NickName, err = readPrefixedString(r, defaultMaxStrSize)
	if err != nil {
		return err
	}

	t.AvatarType, err = readPrefixedString(r, defaultMaxStrSize)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.HaveShield)
	if err != nil {
		return err
	}

	return nil
}

type S2CEnemyDuelJoinMessage struct {
	RetCode       uint32
	NowTs         uint64
	SceneCreateTs uint64
	NewToken      string
	StageID       string
	StageSeed     uint32
	Players       []*EnemyDuelServicePlayer
}

func (m *S2CEnemyDuelJoinMessage) ContentType() uint32 {
	return S2CEnemyDuelJoinMessageType
}

func (m *S2CEnemyDuelJoinMessage) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	var err error

	err = binary.Write(&buf, binary.BigEndian, m.RetCode)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.NowTs)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.SceneCreateTs)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.NewToken)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.StageID)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.StageSeed)
	if err != nil {
		return nil, err
	}

	players, err := SerializeSlice(m.Players)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(players)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write([]byte{0, 0})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *S2CEnemyDuelJoinMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)
	var err error

	err = binary.Read(reader, binary.BigEndian, &m.RetCode)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.NowTs)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.SceneCreateTs)
	if err != nil {
		return err
	}

	m.NewToken, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	m.StageID, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.StageSeed)
	if err != nil {
		return err
	}

	m.Players, err = DeserializeSlice[EnemyDuelServicePlayer, *EnemyDuelServicePlayer](reader, defaultMaxSliceSize)
	if err != nil {
		return err
	}

	return nil
}

type S2CEnemyDuelEndMessage struct {
	Reason uint8
}

func (m *S2CEnemyDuelEndMessage) ContentType() uint32 {
	return S2CEnemyDuelEndMessageType
}

func (m *S2CEnemyDuelEndMessage) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	var err error

	err = binary.Write(&buf, binary.BigEndian, m.Reason)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *S2CEnemyDuelEndMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)
	var err error

	err = binary.Read(reader, binary.BigEndian, &m.Reason)
	if err != nil {
		return err
	}

	return nil
}

type EnemyDuelServiceStepData struct {
	Index    uint32
	Duration uint32
	CheckSeq int32
	Round    uint8
}

func (t *EnemyDuelServiceStepData) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	var err error

	err = binary.Write(&buf, binary.BigEndian, t.Index)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.Duration)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write([]byte{0, 0})
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.CheckSeq)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.Round)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (t *EnemyDuelServiceStepData) UnmarshalBinaryReader(r io.Reader) error {
	var err error

	err = binary.Read(r, binary.BigEndian, &t.Index)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.Duration)
	if err != nil {
		return err
	}

	_, err = io.CopyN(io.Discard, r, 2)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.CheckSeq)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.Round)
	if err != nil {
		return err
	}

	return nil
}

type S2CEnemyDuelHistoryMessage struct {
	Steps []*EnemyDuelServiceStepData
}

func (m *S2CEnemyDuelHistoryMessage) ContentType() uint32 {
	return S2CEnemyDuelHistoryMessageType
}

func (m *S2CEnemyDuelHistoryMessage) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	var err error

	steps, err := SerializeSlice(m.Steps)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(steps)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *S2CEnemyDuelHistoryMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)
	var err error

	m.Steps, err = DeserializeSlice[EnemyDuelServiceStepData, *EnemyDuelServiceStepData](reader, defaultMaxSliceSize)
	if err != nil {
		return err
	}

	return nil
}

type S2CEnemyDuelTeamJoinMessage struct {
	RetCode uint32
	Reason  string
	SvrTime uint64
}

func (m *S2CEnemyDuelTeamJoinMessage) ContentType() uint32 {
	return S2CEnemyDuelTeamJoinMessageType
}

func (m *S2CEnemyDuelTeamJoinMessage) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	var err error

	err = binary.Write(&buf, binary.BigEndian, m.RetCode)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.Reason)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.SvrTime)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *S2CEnemyDuelTeamJoinMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)
	var err error

	err = binary.Read(reader, binary.BigEndian, &m.RetCode)
	if err != nil {
		return err
	}

	m.Reason, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.SvrTime)
	if err != nil {
		return err
	}

	return nil
}

type S2CEnemyDuelKickMessage struct {
	Reason uint8
}

func (m *S2CEnemyDuelKickMessage) ContentType() uint32 {
	return S2CEnemyDuelKickMessageType
}

func (m *S2CEnemyDuelKickMessage) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	var err error

	err = binary.Write(&buf, binary.BigEndian, m.Reason)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *S2CEnemyDuelKickMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)
	var err error

	err = binary.Read(reader, binary.BigEndian, &m.Reason)
	if err != nil {
		return err
	}

	return nil
}

type EnemyDuelPlayerStatus struct {
	PlayerID   string
	NickName   string
	AvatarType string
	AvatarID   string
	State      uint8
	ConnLeave  uint8
	JoinTs     uint64
}

func (t *EnemyDuelPlayerStatus) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer
	var err error

	err = writePrefixedString(&buf, t.PlayerID)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, t.NickName)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, t.AvatarType)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, t.AvatarID)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.State)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.ConnLeave)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, t.JoinTs)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (t *EnemyDuelPlayerStatus) UnmarshalBinaryReader(r io.Reader) error {
	var err error

	t.PlayerID, err = readPrefixedString(r, defaultMaxStrSize)
	if err != nil {
		return err
	}

	t.NickName, err = readPrefixedString(r, defaultMaxStrSize)
	if err != nil {
		return err
	}

	t.AvatarType, err = readPrefixedString(r, defaultMaxStrSize)
	if err != nil {
		return err
	}

	t.AvatarID, err = readPrefixedString(r, defaultMaxStrSize)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.State)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.ConnLeave)
	if err != nil {
		return err
	}

	err = binary.Read(r, binary.BigEndian, &t.JoinTs)
	if err != nil {
		return err
	}

	return nil
}

type S2CEnemyDuelTeamStatusMessage struct {
	State          uint8
	Owner          string
	TeamStateEndTs uint64
	StageID        string
	ModeID         string
	Players        []*EnemyDuelPlayerStatus
	SceneID        string
	Address        string
	Token          string
	AllowNpc       uint8
}

func (m *S2CEnemyDuelTeamStatusMessage) ContentType() uint32 {
	return S2CEnemyDuelTeamStatusMessageType
}

func (m *S2CEnemyDuelTeamStatusMessage) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	var err error

	err = binary.Write(&buf, binary.BigEndian, m.State)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.Owner)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.TeamStateEndTs)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.StageID)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.ModeID)
	if err != nil {
		return nil, err
	}

	players, err := SerializeSlice(m.Players)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(players)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write([]byte{0, 0})
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.SceneID)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.Address)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.Token)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.AllowNpc)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *S2CEnemyDuelTeamStatusMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)
	var err error

	err = binary.Read(reader, binary.BigEndian, &m.State)
	if err != nil {
		return err
	}

	m.Owner, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.TeamStateEndTs)
	if err != nil {
		return err
	}

	m.StageID, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	m.ModeID, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	m.Players, err = DeserializeSlice[EnemyDuelPlayerStatus, *EnemyDuelPlayerStatus](reader, defaultMaxSliceSize)
	if err != nil {
		return err
	}

	_, err = io.CopyN(io.Discard, reader, 2)
	if err != nil {
		return err
	}

	m.SceneID, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	m.Address, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	m.Token, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	err = binary.Read(reader, binary.BigEndian, &m.AllowNpc)
	if err != nil {
		return err
	}

	return nil
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
	var buf bytes.Buffer
	var err error

	err = writePrefixedString(&buf, m.EmojiGroup)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.EmojiID)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *C2SEnemyDuelEmojiMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)
	var err error

	m.EmojiGroup, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	m.EmojiID, err = readPrefixedString(reader, defaultMaxStrSize)
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
	return nil, nil
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
	var buf bytes.Buffer
	var err error

	err = writePrefixedString(&buf, m.PlayerID)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.SceneID)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.Token)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *C2SEnemyDuelJoinMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)
	var err error

	m.PlayerID, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	m.SceneID, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	m.Token, err = readPrefixedString(reader, defaultMaxStrSize)
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
	var buf bytes.Buffer
	var err error

	err = binary.Write(&buf, binary.BigEndian, m.Side)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.Info)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.InfoHash)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *C2SEnemyDuelRoundSettleMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)
	var err error

	err = binary.Read(reader, binary.BigEndian, &m.Side)
	if err != nil {
		return err
	}

	m.Info, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	m.InfoHash, err = readPrefixedString(reader, defaultMaxStrSize)
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
	var buf bytes.Buffer
	var err error

	err = writePrefixedString(&buf, m.PlayerID)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.Side)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.IsPlayer)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.AllIn)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *C2SEnemyDuelBetMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)
	var err error

	m.PlayerID, err = readPrefixedString(reader, defaultMaxStrSize)
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
	var buf bytes.Buffer
	var err error

	err = binary.Write(&buf, binary.BigEndian, m.Seq)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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
	return nil, nil
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
	var buf bytes.Buffer
	var err error

	err = binary.Write(&buf, binary.BigEndian, m.Seq)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, binary.BigEndian, m.Time)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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

type C2SEnemyDuelFinalSettleMessage struct{}

func (m *C2SEnemyDuelFinalSettleMessage) ContentType() uint32 {
	return C2SEnemyDuelFinalSettleMessageType
}

func (m *C2SEnemyDuelFinalSettleMessage) Marshal() ([]byte, error) {
	return nil, nil
}

func (m *C2SEnemyDuelFinalSettleMessage) Unmarshal(payload []byte) error {
	return nil
}

type C2SEnemyDuelTeamJoinMessage struct {
	PlayerID  string
	TeamID    string
	TeamToken string
}

func (m *C2SEnemyDuelTeamJoinMessage) ContentType() uint32 {
	return C2SEnemyDuelTeamJoinMessageType
}

func (m *C2SEnemyDuelTeamJoinMessage) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	var err error

	err = writePrefixedString(&buf, m.PlayerID)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.TeamID)
	if err != nil {
		return nil, err
	}

	err = writePrefixedString(&buf, m.TeamToken)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (m *C2SEnemyDuelTeamJoinMessage) Unmarshal(payload []byte) error {
	reader := bytes.NewReader(payload)
	var err error

	m.PlayerID, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	m.TeamID, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	m.TeamToken, err = readPrefixedString(reader, defaultMaxStrSize)
	if err != nil {
		return err
	}

	return nil
}
