package contract

import (
	"bytes"
	"testing"
)

func TestS2CEnemyDuelEmojiMessageMarshal(t *testing.T) {
	msg := &S2CEnemyDuelEmojiMessage{
		PlayerID:   "123",
		EmojiGroup: "some_group",
		EmojiID:    "some_id",
	}
	expectedPayload := []byte{0x0, 0x3, 0x31, 0x32, 0x33, 0x0, 0xa, 0x73, 0x6f, 0x6d, 0x65, 0x5f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x0, 0x7, 0x73, 0x6f, 0x6d, 0x65, 0x5f, 0x69, 0x64}

	payload, err := msg.Marshal()
	if err != nil {
		t.Fatalf("S2CEnemyDuelEmojiMessageMarshal failed: %v", err)
	}

	if !bytes.Equal(expectedPayload, payload) {
		t.Errorf("payload mismatch: \ngot: %#v\nwant: %#v", payload, expectedPayload)
	}
}
