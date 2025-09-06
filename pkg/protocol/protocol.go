package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	headerSize = 8
)

type Envelop struct {
	Type    uint32
	Payload []byte
}

const maxPayloadSize = 1 << 10 << 10

func WriteEnvelop(w io.Writer, env *Envelop) error {
	if len(env.Payload) > 0xffffffff {
		return fmt.Errorf("payload too large (%d bytes)", len(env.Payload))
	}

	header := make([]byte, headerSize)

	binary.BigEndian.PutUint32(header[0:4], uint32(len(env.Payload)))
	binary.BigEndian.PutUint32(header[4:8], env.Type)

	if _, err := w.Write(header); err != nil {
		return err
	}

	if len(env.Payload) > 0 {
		if _, err := w.Write(env.Payload); err != nil {
			return err
		}
	}

	return nil
}

func ReadEnvelop(r io.Reader) (*Envelop, error) {
	header := make([]byte, headerSize)

	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(header[0:4])
	msgType := binary.BigEndian.Uint32(header[4:8])

	if length > maxPayloadSize {
		return nil, fmt.Errorf("envelop too large (%d)", length)
	}

	var payload []byte
	if length > 0 {
		payload = make([]byte, length)
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, err
		}
	}

	return &Envelop{
		Type:    msgType,
		Payload: payload,
	}, nil
}
