package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	kvserver "github.com/edudev/chord/server"
)

var n uint
var bootstrapAddr string
var checkFingertable bool
var checkSuccessors bool

var startTime time.Time
var timeSet bool

var clientCache map[string](*grpc.ClientConn)

func getClientConn(addr string) (conn *grpc.ClientConn) {
	if conn, ok := clientCache[addr]; ok {
		return conn
	}

	if conn, ok := clientCache[addr]; ok {
		return conn
	}

	conn = kvserver.GetGRPCConnection(string(addr))
	clientCache[addr] = conn

	// TODO: close the connection at some point
	return conn
}

func traverseSuccessors(bootstrapAddr string) (nodesFound []string) {
	nextNodeAddr := bootstrapAddr

	nodesFound = []string{bootstrapAddr}

	for {
		conn := getClientConn(nextNodeAddr)
		if !timeSet {
			startTime = time.Now()
			timeSet = true
		}
		client := kvserver.NewChordRingClient(conn)

		nodeRPC, err := client.GetSuccessor(context.Background(), new(empty.Empty))
		if err != nil {
			log.Printf("Failed to execute RPC %v", err)
			return
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

func checkFingertables(nodeAddresses []string) (incorrectEntries int) {

	for i, node := range nodeAddresses {
		conn := getClientConn(node)
		client := kvserver.NewChordRingClient(conn)

		list, e := client.GetFingerTable(context.Background(), new(empty.Empty))
		if e != nil {
			log.Fatalf("Can't get fingertable: %v", e)
			return
		}

		correctFingerTable := kvserver.CreatePerfectFingertable(node, nodeAddresses)
		for k, correctAddr := range correctFingerTable {
			if list.Nodes[k].Address != correctAddr {
				log.Printf("Node #%v addr %v has incorrect finger at index %v: %v instead of %v", i, node, k, list.Nodes[k].Address, correctAddr)
				incorrectEntries++
			}
		}
	}
	return
}

func checkSuccessorsCorrect(nodes []string) (incorrectSuccessors int) {
	incorrectSuccessors = 0
	for i, node := range nodes {
		conn := getClientConn(node)
		client := kvserver.NewChordRingClient(conn)

		list, e := client.GetSuccessorList(context.Background(), new(empty.Empty))
		if e != nil {
			log.Fatalf("Can't get successor list: %v", e)
		}

		correctSuccessorList := kvserver.CreatePerfectSuccessorList(node, nodes)
		for k, successor := range list.GetNodes() {
			if successor.Address != correctSuccessorList[k] {
				log.Printf("Node #%v addr %v has incorrect successor at index %v: %v instead of %v", i, node, k, successor.Address, correctSuccessorList[k])
				incorrectSuccessors++
			}
		}
		incorrectSuccessors += len(correctSuccessorList) - len(list.GetNodes())
	}
	return
}

func main() {
	clientCache = make(map[string](*grpc.ClientConn))
	timeSet = false

	log.SetPrefix("TRACE: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.Println("log initialised")

	nFlag := flag.Uint("N", 64, "Amount of nodes to check for. Use 0 for continous monitoring")
	bootstrapAddrFlag := flag.String("addr", "127.0.0.1:21210", "the address of the node to bootrstrap from")
	checkFingerTableFlag := flag.Bool("check_fingers", false, "also wait for fingertable to be correct")
	checkSuccessorListFlag := flag.Bool("check_successors", false, "also wait for successor list to be correct")

	flag.Parse()

	n = *nFlag
	bootstrapAddr = *bootstrapAddrFlag
	checkFingertable = *checkFingerTableFlag
	checkSuccessors = *checkSuccessorListFlag

	iterations := 0
	for {
		nodes := traverseSuccessors(bootstrapAddr)
		log.Printf("Found %d nodes", len(nodes))

		var incorrectSuccessors int
		if checkSuccessors || n == 0 {
			incorrectSuccessors = checkSuccessorsCorrect(nodes)
			if incorrectSuccessors > 0 {
				log.Print("Incorrect succesors: ", incorrectSuccessors)
			} else {
				log.Print("All nodes have correct successors!")
			}
		}

		var incorrectFingerTableEntries int
		if checkFingertable || n == 0 {
			incorrectFingerTableEntries = checkFingertables(nodes)
			if incorrectFingerTableEntries > 0 {
				log.Printf("There are %v incorrect finger table entries", incorrectFingerTableEntries)
			} else {
				log.Print("All nodes have correct fingers!")
			}
		}

		if n == 0 {
			fmt.Printf("%v,%v,%v,%v\n", (time.Now().Sub(startTime)).Seconds(), len(nodes), incorrectSuccessors, incorrectFingerTableEntries)
		}
		if (n != 0 && int(n) == len(nodes)) && (!checkFingertable || incorrectFingerTableEntries == 0) && (!checkSuccessors || incorrectSuccessors == 0) && (n != 0 || iterations > 100) {
			break
		}
		iterations++
		time.Sleep(startTime.Add(time.Duration(1000*iterations) * time.Millisecond).Sub(time.Now()))
	}
}
