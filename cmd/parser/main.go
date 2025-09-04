package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/OpenBachelor/OpenBachelorSS/pkg/contract"
	"github.com/OpenBachelor/OpenBachelorSS/pkg/protocol"
	"github.com/davecgh/go-spew/spew"
)

func processEnv(env *protocol.Envelop, verbose bool, parsedMsgCnt *int, parsedAndVerifiedMsgCnt *int) {
	defer func() {
		if r := recover(); r != nil {
			if verbose {
				log.Printf("panic while processing env %+v: %v", env, r)
			}
		}
	}()

	content, err := contract.FromEnvelop(env)
	if err != nil {
		if verbose {
			log.Printf("failed to get content of %+v: %v", env, err)
		}
		return
	}

	*parsedMsgCnt++

	spew.Dump(content)

	newEnv, err := contract.ToEnvelop(content)

	if err != nil {
		if verbose {
			log.Printf("failed to get envelop of %+v: %v", content, err)
		}
		return
	}

	if !bytes.Equal(env.Payload, newEnv.Payload) {
		if verbose {
			log.Printf("envelop mismatch: \ngot: %#v\nwant: %#v", newEnv.Payload, env.Payload)
		}
		return
	}

	*parsedAndVerifiedMsgCnt++

}

func main() {
	verbose := flag.Bool("verbose", false, "print err")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatalf("raw stream file not provided")
	}

	streamFilePath := flag.Args()[0]
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
	parsedMsgCnt := 0
	parsedAndVerifiedMsgCnt := 0
	for {
		receivedEnv, err := protocol.ReadEnvelop(serverConn)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("ReadEnvelop failed: %v", err)
			break
		}

		msgCnt++
		msgTypeMap[receivedEnv.Type]++

		processEnv(receivedEnv, *verbose, &parsedMsgCnt, &parsedAndVerifiedMsgCnt)
	}

	wg.Wait()

	log.Printf("num of msg: %d", msgCnt)
	log.Printf("num of parsed msg: %d", parsedMsgCnt)
	log.Printf("num of parsed & verified msg: %d", parsedAndVerifiedMsgCnt)
	log.Printf("msg type: %v", msgTypeMap)
}
