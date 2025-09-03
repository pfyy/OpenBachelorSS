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

		func(env *protocol.Envelop) {
			defer func() {
				if r := recover(); r != nil {
					if *verbose {
						log.Printf("panic while getting content of %+v: %v", env, err)
					}
				}
			}()

			content, err := contract.FromEnvelop(env)
			if err != nil {
				if *verbose {
					log.Printf("failed to get content of %+v: %v", env, err)
				}
				return
			}

			parsedMsgCnt++
			spew.Dump(content)

			func(c contract.Content, e *protocol.Envelop) {
				defer func() {
					if r := recover(); r != nil {
						if *verbose {
							log.Printf("panic while getting envelop of %+v: %v", c, err)
						}
					}
				}()

				newEnv, err := contract.ToEnvelop(c)

				if err != nil {
					if *verbose {
						log.Printf("failed to get envelop of %+v: %v", env, err)
					}
					return
				}

				if !bytes.Equal(e.Payload, newEnv.Payload) {
					if *verbose {
						log.Printf("envelop mismatch: \ngot: %#v\nwant: %#v", newEnv.Payload, e.Payload)
					}
					return
				}

				parsedAndVerifiedMsgCnt++
			}(content, env)

		}(receivedEnv)

		msgCnt++
		msgTypeMap[receivedEnv.Type]++
	}

	wg.Wait()

	log.Printf("num of msg: %d", msgCnt)
	log.Printf("num of parsed msg: %d", parsedMsgCnt)
	log.Printf("num of parsed & verified msg: %d", parsedAndVerifiedMsgCnt)
	log.Printf("msg type: %v", msgTypeMap)
}
