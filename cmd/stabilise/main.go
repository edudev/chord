package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes/empty"

	kvserver "github.com/edudev/chord/server"
)

func traverseSuccessors(bootstrapAddr string) int {
	nextNodeAddr := bootstrapAddr

	nodeCount := int(0)

	for {
		nodeCount++

		log.Printf("Found node %s", nextNodeAddr)
		conn := kvserver.GetGRPCConnection(nextNodeAddr)
		client := kvserver.NewChordRingClient(conn)

		nodeRPC, err := client.GetSuccessor(context.Background(), new(empty.Empty))
		if err != nil {
			log.Fatalf("Failed to execute RPC %v", err)
		}

		nextNodeAddr = nodeRPC.GetAddress()

		if nextNodeAddr == bootstrapAddr {
			break
		}
	}

	return nodeCount
}

func main() {
	log.SetPrefix("TRACE: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.Println("log initialised")

	if len(os.Args) != 3 {
		panic("expected arguments: <bootstrap node> <ring size>")
	}

	ringSize, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic("ring size not an integer")
	}

	bootstrapAddr := os.Args[1]

	for {
		nodeCount := traverseSuccessors(bootstrapAddr)
		log.Printf("Found %d nodes", nodeCount)

		if nodeCount == ringSize {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}
}
