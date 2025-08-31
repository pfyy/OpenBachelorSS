package protocol

import (
	"encoding/binary"
	"io"
	"net"
)

const (
	headerSize = 8
)

type Envelop struct {
	Type    uint32
	Payload []byte
}

func WriteEnvelop(conn net.Conn, msg *Envelop) error {
	header := make([]byte, headerSize)

	binary.BigEndian.PutUint32(header[0:4], uint32(len(msg.Payload)))
	binary.BigEndian.PutUint32(header[4:8], msg.Type)

	if _, err := conn.Write(header); err != nil {
		return err
	}

	if len(msg.Payload) > 0 {
		if _, err := conn.Write(msg.Payload); err != nil {
			return err
		}
	}

	return nil
}

func ReadEnvelop(conn net.Conn) (*Envelop, error) {
	header := make([]byte, headerSize)

	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(header[0:4])
	msgType := binary.BigEndian.Uint32(header[4:8])

	var payload []byte
	if length > 0 {
		payload = make([]byte, length)
		if _, err := io.ReadFull(conn, payload); err != nil {
			return nil, err
		}
	}

	return &Envelop{
		Type:    msgType,
		Payload: payload,
	}, nil
}
