package main

import (
	"io"
	"log"
	"net"
	"os"

	"github.com/OpenBachelor/OpenBachelorSS/pkg/protocol"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("raw stream file not provided")
	}

	streamFilePath := os.Args[1]
	streamData, err := os.ReadFile(streamFilePath)
	if err != nil {
		log.Fatalf("failed to read raw stream file '%s': %v", streamFilePath, err)
	}

	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()

	go func() {
		defer clientConn.Close()

		_, err := clientConn.Write(streamData)

		if err != nil {
			log.Printf("write raw stream failed: %v", err)
		}
	}()

	msgCnt := 0
	msgTypeMap := make(map[uint32]struct{})
	for {
		receivedMsg, err := protocol.ReadMessage(serverConn)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("ReadMessage failed: %v", err)
			break
		}

		msgCnt++
		msgTypeMap[receivedMsg.Type] = struct{}{}
	}

	log.Printf("num of msg: %d", msgCnt)
	log.Printf("msg type: %v", msgTypeMap)
}
