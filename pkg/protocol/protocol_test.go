package protocol

import (
	"net"
	"reflect"
	"testing"
)

func cloneBytes(payload []byte) []byte {
	return append([]byte(nil), payload...)
}

func TestReadWriteEnvelope(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()

	const testMsgType uint32 = 123
	testPayload := []byte("some_payload")
	expectedEnv := &Envelope{
		Type:    testMsgType,
		Payload: cloneBytes(testPayload),
	}

	errChan := make(chan error, 1)

	go func() {
		defer clientConn.Close()

		errChan <- WriteEnvelope(clientConn, &Envelope{Type: testMsgType, Payload: cloneBytes(testPayload)})
	}()

	receivedEnv, err := ReadEnvelope(serverConn)
	if err != nil {
		t.Fatalf("ReadEnvelope failed: %v", err)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("WriteEnvelope failed: %v", err)
	}

	if !reflect.DeepEqual(expectedEnv, receivedEnv) {
		t.Errorf("envelope mismatch: \ngot: %+v\nwant: %+v", receivedEnv, expectedEnv)
	}
}

func TestReadEnvelopeWithPacket(t *testing.T) {
	packet := []byte{0x0, 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x14}
	const testMsgType uint32 = 1
	testPayload := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x14}
	expectedEnv := &Envelope{
		Type:    testMsgType,
		Payload: cloneBytes(testPayload),
	}

	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()

	errChan := make(chan error, 1)

	go func() {
		defer clientConn.Close()

		_, err := clientConn.Write(packet)

		errChan <- err
	}()

	receivedEnv, err := ReadEnvelope(serverConn)
	if err != nil {
		t.Fatalf("ReadEnvelope failed: %v", err)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("write packet failed: %v", err)
	}

	if !reflect.DeepEqual(expectedEnv, receivedEnv) {
		t.Errorf("envelope mismatch: \ngot: %+v\nwant: %+v", receivedEnv, expectedEnv)
	}
}
