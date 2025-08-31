package protocol

import (
	"net"
	"reflect"
	"testing"
)

func TestReadWriteMessage(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()

	const testMsgType uint32 = 123
	testPayload := []byte("some_payload")
	expectedMsg := &Message{
		Type:    testMsgType,
		Payload: append([]byte(nil), testPayload...),
	}

	errChan := make(chan error, 1)

	go func() {
		defer clientConn.Close()

		errChan <- WriteMessage(clientConn, testMsgType, testPayload)
	}()

	receivedMsg, err := ReadMessage(serverConn)
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("WriteMessage failed: %v", err)
	}

	if !reflect.DeepEqual(expectedMsg, receivedMsg) {
		t.Errorf("message mismatch: \ngot: %+v\nwant: %+v", receivedMsg, expectedMsg)
	}
}

func TestReadMessageWithPacket(t *testing.T) {
	packet := []byte{0x0, 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x14}
	const testMsgType uint32 = 1
	testPayload := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x14}
	expectedMsg := &Message{
		Type:    testMsgType,
		Payload: append([]byte(nil), testPayload...),
	}

	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()

	errChan := make(chan error, 1)

	go func() {
		defer clientConn.Close()

		_, err := clientConn.Write(packet)

		errChan <- err
	}()

	receivedMsg, err := ReadMessage(serverConn)
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("write packet failed: %v", err)
	}

	if !reflect.DeepEqual(expectedMsg, receivedMsg) {
		t.Errorf("message mismatch: \ngot: %+v\nwant: %+v", receivedMsg, expectedMsg)
	}
}
