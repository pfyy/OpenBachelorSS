package protocol

import (
	"net"
	"reflect"
	"testing"
)

func TestReadWriteEnvelop(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()

	const testMsgType uint32 = 123
	testPayload := []byte("some_payload")
	expectedMsg := &Envelop{
		Type:    testMsgType,
		Payload: append([]byte(nil), testPayload...),
	}

	errChan := make(chan error, 1)

	go func() {
		defer clientConn.Close()

		errChan <- WriteEnvelop(clientConn, testMsgType, testPayload)
	}()

	receivedMsg, err := ReadEnvelop(serverConn)
	if err != nil {
		t.Fatalf("ReadEnvelop failed: %v", err)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("WriteEnvelop failed: %v", err)
	}

	if !reflect.DeepEqual(expectedMsg, receivedMsg) {
		t.Errorf("message mismatch: \ngot: %+v\nwant: %+v", receivedMsg, expectedMsg)
	}
}

func TestReadEnvelopWithPacket(t *testing.T) {
	packet := []byte{0x0, 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x14}
	const testMsgType uint32 = 1
	testPayload := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x14}
	expectedMsg := &Envelop{
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

	receivedMsg, err := ReadEnvelop(serverConn)
	if err != nil {
		t.Fatalf("ReadEnvelop failed: %v", err)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("write packet failed: %v", err)
	}

	if !reflect.DeepEqual(expectedMsg, receivedMsg) {
		t.Errorf("message mismatch: \ngot: %+v\nwant: %+v", receivedMsg, expectedMsg)
	}
}
