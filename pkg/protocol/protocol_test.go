package protocol

import (
	"net"
	"reflect"
	"testing"
)

func cloneBytes(payload []byte) []byte {
	return append([]byte(nil), payload...)
}

func TestReadWriteEnvelop(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()

	const testMsgType uint32 = 123
	testPayload := []byte("some_payload")
	expectedEnv := &Envelop{
		Type:    testMsgType,
		Payload: cloneBytes(testPayload),
	}

	errChan := make(chan error, 1)

	go func() {
		defer clientConn.Close()

		errChan <- WriteEnvelop(clientConn, &Envelop{Type: testMsgType, Payload: cloneBytes(testPayload)})
	}()

	receivedEnv, err := ReadEnvelop(serverConn)
	if err != nil {
		t.Fatalf("ReadEnvelop failed: %v", err)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("WriteEnvelop failed: %v", err)
	}

	if !reflect.DeepEqual(expectedEnv, receivedEnv) {
		t.Errorf("envelop mismatch: \ngot: %+v\nwant: %+v", receivedEnv, expectedEnv)
	}
}

func TestReadEnvelopWithPacket(t *testing.T) {
	packet := []byte{0x0, 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x14}
	const testMsgType uint32 = 1
	testPayload := []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x14}
	expectedEnv := &Envelop{
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

	receivedEnv, err := ReadEnvelop(serverConn)
	if err != nil {
		t.Fatalf("ReadEnvelop failed: %v", err)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("write packet failed: %v", err)
	}

	if !reflect.DeepEqual(expectedEnv, receivedEnv) {
		t.Errorf("envelop mismatch: \ngot: %+v\nwant: %+v", receivedEnv, expectedEnv)
	}
}
