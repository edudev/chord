package main

import (
	"fmt"
	"log"
	"sync"

	memcachedWrapper "github.com/edudev/chord/memcached"
	kvserver "github.com/edudev/chord/server"
	memcached "github.com/mattrobenolt/go-memcached"
)

const (
	N         uint   = 64
	chordPort uint16 = 21211
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

	reply <- true
}

func waitForStability() {
	// Using blocking channel to wait until
	// isStable function reply with true
	reply := make(chan bool)
	go isStable(reply)
	<-reply
}

func createNode(id uint) (server kvserver.ChordServer) {
	addr := fmt.Sprintf("127.0.0.1:%d", chordPort+uint16(id))
	server = kvserver.New(addr)
	return
}

func addNodeToRing(prevServersList []kvserver.ChordServer, id uint) (newServer kvserver.ChordServer, newServersList []kvserver.ChordServer) {
	var nodeToJoinTo *string
	nodeToJoinTo = nil
	if len(prevServersList) > 0 {
		a := prevServersList[0].Address()
		nodeToJoinTo = &a
	}

	newServer = createNode(id + 1)
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
		//log.Printf("do listen and serve %v %v", i, server)
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

	// Initialising an empty server array and
	// getting only the first server node
	var servers = []kvserver.ChordServer{}
	var newServer kvserver.ChordServer
	newServer, servers = addNodeToRing(servers, 0)

	holder := memcachedWrapper.New(&newServer)
	memcachedServer := memcached.NewServer("127.0.0.1:11211", &holder)

	var wg sync.WaitGroup
	listenAndServe(&wg, -1, memcachedServer)
	listenAndServe(&wg, 0, &newServer)

	// Adding node in ring one by one
	for id := uint(1); id < N; id++ {
		newServer, servers = addNodeToRing(servers, id)
		listenAndServe(&wg, int(id), &newServer)
		waitForStability()
	}

	fmt.Println(servers)

	wg.Wait()
	shutdown(servers)
}
