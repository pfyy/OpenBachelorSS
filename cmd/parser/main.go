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

func processEnv(msgDomain contract.MessageDomain, env *protocol.Envelope, verbose bool, parsedMsgCnt *int, parsedAndVerifiedMsgCnt *int) {
	defer func() {
		if r := recover(); r != nil {
			if verbose {
				log.Printf("panic while processing env %+v: %v", env, r)
			}
		}
	}()

	content, err := contract.FromEnvelope(msgDomain, env)
	if err != nil {
		if verbose {
			log.Printf("failed to get content of %+v: %v", env, err)
		}
		return
	}

	if _, ok := content.(*contract.UnknownMessage); ok {
		if verbose {
			log.Printf("unknown env %+v: %v", env, err)
		}
		return
	}

	*parsedMsgCnt++

	spew.Dump(content)

	newEnv, err := contract.ToEnvelope(content)

	if err != nil {
		if verbose {
			log.Printf("failed to get envelope of %+v: %v", content, err)
		}
		return
	}

	if !bytes.Equal(env.Payload, newEnv.Payload) {
		if verbose {
			log.Printf("envelope mismatch: \ngot: %#v\nwant: %#v", newEnv.Payload, env.Payload)
		}
		return
	}

	*parsedAndVerifiedMsgCnt++

}

func main() {

	msgDomainInt := flag.Int("domain", int(contract.EnemyDuelMessageDomain), "message domain")

	verbose := flag.Bool("verbose", false, "print err")
	flag.Parse()

	msgDomain := contract.MessageDomain(*msgDomainInt)

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
		receivedEnv, err := protocol.ReadEnvelope(serverConn)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("ReadEnvelope failed: %v", err)
			break
		}

		msgCnt++
		msgTypeMap[receivedEnv.Type]++

		processEnv(msgDomain, receivedEnv, *verbose, &parsedMsgCnt, &parsedAndVerifiedMsgCnt)
	}

	wg.Wait()

	log.Printf("num of msg: %d", msgCnt)
	log.Printf("num of parsed msg: %d", parsedMsgCnt)
	log.Printf("num of parsed & verified msg: %d", parsedAndVerifiedMsgCnt)
	log.Printf("msg type: %v", msgTypeMap)
}
