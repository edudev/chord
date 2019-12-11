package main

import (
	"flag"
	"fmt"
	"log"
	"sync"

	memcachedWrapper "github.com/edudev/chord/memcached"
	kvserver "github.com/edudev/chord/server"
	memcached "github.com/mattrobenolt/go-memcached"
)

const (
	N         uint   = 64
	chordPort uint16 = 21210
)

func isStable(reply chan bool) {
	// Make RPC calls to each node to fetch
	// their last finger table update

	// Basic example of time difference
	// given below

	// NOTE: Either fix time everywhere as UTC
	// or Unix(). Keep it one

	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	// lastFingerTableUpdateTime := time.Date(2019, 11, 29, 20, 19, 48, 324359102, time.UTC)
	// currentTime := time.Now().UTC()

	// if diff := currentTime.Sub(lastFingerTableUpdateTime).Seconds(); diff > 10 {
	// 	return true
	// }
	// return false
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
	// TODO remove sleep
	// this makes the logs a bit easier to follow for now
	//time.Sleep(2 * time.Second)
	reply <- true
}

func waitForStability() {
	// Using blocking channel to wait until
	// isStable function reply with true
	reply := make(chan bool)
	go isStable(reply)
	<-reply
}

func createNode(localAddr string, id uint) (server kvserver.ChordServer) {
	addr := fmt.Sprintf("%v:%d", localAddr, chordPort+uint16(id))
	server = kvserver.New(addr)
	return
}

func addNodeToRing(localAddr string, nodeToJoinTo *string, prevServersList []kvserver.ChordServer, id uint) (newServer kvserver.ChordServer, newServersList []kvserver.ChordServer) {
	if nodeToJoinTo == nil {
		if len(prevServersList) > 0 {
			a := prevServersList[0].Address()
			nodeToJoinTo = &a
		}
	}

	newServer = createNode(localAddr, id)
	newServersList = append(prevServersList, newServer)

	newServer.Join(nodeToJoinTo)

	return
}

type serverType interface {
	ListenAndServe() error
}

func listenAndServe(wg *sync.WaitGroup, i int, server serverType) {
	wg.Add(1)
	go func() {
		// log.Printf("do listen and serve %v %v", i, server)
		server.ListenAndServe()
		log.Printf("done listen and serve")
		wg.Done()
	}()
}

func shutdown(servers []kvserver.ChordServer) {
	for _, server := range servers {
		server.Stop()
	}
}

func main() {
	log.SetPrefix("TRACE: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.Println("log initialised")

	nFlag := flag.Uint("N", 64, "number of nodes to run in this process")
	joinFlag := flag.String("join", "", "an existing node to join to")
	localAddrFlag := flag.String("addr", "127.0.0.1", "local IP address to bind to")
	learnNodesFlag := flag.Bool("learn_nodes", false, "whether to aggresively learn nodes for fingers")
	intelligentFixFingersFlag := flag.Bool("fix_fingers", false, "whether to intelligently fix fingers")
	logHopsFlag := flag.Bool("log_hops", false, "whether to log hop counts for predecessor search")

	flag.Parse()

	kvserver.LearnNodes = *learnNodesFlag
	kvserver.IntelligentFixFingers = *intelligentFixFingersFlag
	kvserver.LogHopCounts = *logHopsFlag

	N := *nFlag
	nodeToJoinTo := &*joinFlag
	if *joinFlag == "" {
		nodeToJoinTo = nil
	}
	localAddr := *localAddrFlag

	// Initialising an empty server array and
	// getting only the first server node
	var servers = []kvserver.ChordServer{}
	var newServer kvserver.ChordServer
	newServer, servers = addNodeToRing(localAddr, nodeToJoinTo, servers, 0)

	holder := memcachedWrapper.New(&newServer)
	memcachedServer := memcached.NewServer(localAddr+":11211", &holder)

	var wg sync.WaitGroup
	listenAndServe(&wg, -1, memcachedServer)
	listenAndServe(&wg, 0, &newServer)

	// Adding node in ring one by one
	for id := uint(1); id < N; id++ {
		newServer, servers = addNodeToRing(localAddr, nodeToJoinTo, servers, id)
		listenAndServe(&wg, int(id), &newServer)
		waitForStability()
	}

	fmt.Println(servers)

	wg.Wait()
	shutdown(servers)
}
