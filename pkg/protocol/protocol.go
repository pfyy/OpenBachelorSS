package protocol

import (
	"encoding/binary"
	"io"
	"net"
)

const (
	headerSize = 8
)

type Message struct {
	Type    uint32
	Payload []byte
}

func WriteMessage(conn net.Conn, msgType uint32, payload []byte) error {
	header := make([]byte, headerSize)

	binary.BigEndian.PutUint32(header[0:4], uint32(len(payload)))
	binary.BigEndian.PutUint32(header[4:8], msgType)

	if _, err := conn.Write(header); err != nil {
		return err
	}

	if len(payload) > 0 {
		if _, err := conn.Write(payload); err != nil {
			return err
		}
	}

	return nil
}

func ReadMessage(conn net.Conn) (*Message, error) {
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

	return &Message{
		Type:    msgType,
		Payload: payload,
	}, nil
}
