package main

import (
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
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

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer clientConn.Close()

		_, err := clientConn.Write(streamData)

		if err != nil {
			log.Printf("write raw stream failed: %v", err)
		}
	}()

	msgCnt := 0
	msgTypeMap := make(map[uint32]int)
	for {
		receivedEnv, err := protocol.ReadEnvelop(serverConn)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("ReadEnvelop failed: %v", err)
			break
		}

		func() {
			defer recover()

			content, err := contract.FromEnvelop(receivedEnv)
			if err == nil {
				log.Printf("msg: %+v", content)
			}
		}()

		msgCnt++
		msgTypeMap[receivedEnv.Type]++
	}

	wg.Wait()

	log.Printf("num of msg: %d", msgCnt)
	log.Printf("msg type: %v", msgTypeMap)
}
