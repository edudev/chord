package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/empty"

	kvserver "github.com/edudev/chord/server"
)

var n uint
var bootstrapAddr string
var checkFingertable bool
var checkSuccessors bool

func traverseSuccessors(bootstrapAddr string) (nodesFound []string) {
	nextNodeAddr := bootstrapAddr

	nodesFound = []string{bootstrapAddr}

	for {
		conn := kvserver.GetGRPCConnection(nextNodeAddr)
		client := kvserver.NewChordRingClient(conn)

		nodeRPC, err := client.GetSuccessor(context.Background(), new(empty.Empty))
		if err != nil {
			log.Fatalf("Failed to execute RPC %v", err)
		}

		nextNodeAddr = nodeRPC.GetAddress()

		for _, node := range nodesFound {
			if node == nextNodeAddr {
				return
			}
		}
		nodesFound = append(nodesFound, nextNodeAddr)
	}

	return
}

func checkFingertables(nodeAddresses []string) bool {
	for i, node := range nodeAddresses {
		conn := kvserver.GetGRPCConnection(node)
		client := kvserver.NewChordRingClient(conn)

		list, e := client.GetFingerTable(context.Background(), new(empty.Empty))
		if e != nil {
			log.Fatalf("Can't get fingertable: %v", e)
			return false
		}

		correctFingerTable := kvserver.CreatePerfectFingertable(node, nodeAddresses)
		for k, correctAddr := range correctFingerTable {
			if list.Nodes[k].Address != correctAddr {
				log.Printf("Node #%v addr %v has incorrect finger at index %v: %v instead of %v", i, node, k, list.Nodes[k].Address, correctAddr)
				return false
			}
		}
	}
	return true
}

func checkSuccessorsCorrect(nodes []string) bool {
	for i, node := range nodes {
		conn := kvserver.GetGRPCConnection(node)
		client := kvserver.NewChordRingClient(conn)

		list, e := client.GetSuccessorList(context.Background(), new(empty.Empty))
		if e != nil {
			log.Fatalf("Can't get successor list: %v", e)
			return false
		}

		correctSuccessorList := kvserver.CreatePerfectSuccessorList(node, nodes)
		for k, successor := range list.Nodes {
			if successor.Address != correctSuccessorList[k] {
				log.Printf("Node #%v addr %v has incorrect successor at index %v: %v instead of %v", i, node, k, successor.Address, correctSuccessorList[k])
				return false
			}
		}
	}
	return true
}

func main() {
	log.SetPrefix("TRACE: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.Println("log initialised")

	nFlag := flag.Uint("N", 64, "Amount of nodes to check for")
	bootstrapAddrFlag := flag.String("addr", "127.0.0.1:21210", "the address of the node to bootrstrap from")
	checkFingerTableFlag := flag.Bool("check_fingers", false, "also wait for fingertable to be correct")
	checkSuccessorListFlag := flag.Bool("check_successors", false, "also wait for successor list to be correct")

	flag.Parse()

	n = *nFlag
	bootstrapAddr = *bootstrapAddrFlag
	checkFingertable = *checkFingerTableFlag
	checkSuccessors = *checkSuccessorListFlag

	for {
		nodes := traverseSuccessors(bootstrapAddr)
		log.Printf("Found %d nodes", len(nodes))

		if len(nodes) != int(n) {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if checkSuccessors && !checkSuccessorsCorrect(nodes) {
			time.Sleep(5000 * time.Millisecond)
			continue
		}

		log.Print("All nodes have correct successors!")

		if checkFingertable && !checkFingertables(nodes) {
			time.Sleep(5000 * time.Millisecond)
			continue
		}

		log.Print("All nodes have correct fingers!")

		break
	}
}
