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

func createNode(id uint) (server kvserver.ChordServer) {
	addr := fmt.Sprintf("127.0.0.1:%d", chordPort+uint16(id))
	server = kvserver.New(addr)
	return
}

func populateNodes(count uint) (servers []kvserver.ChordServer) {
	servers = make([]kvserver.ChordServer, count)

	for id := uint(0); id < count; id++ {
		servers[id] = createNode(id + 1)
	}

	for _, server := range servers {
		server.FillFingerTable(servers)
	}

	return
}

func listenAndServe(wg *sync.WaitGroup, f func() error) {
	wg.Add(1)
	go func() {
		f()
		wg.Done()
	}()
}

func main() {
	log.SetPrefix("TRACE: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.Println("log initialised")

	servers := populateNodes(N)
	backend := servers[0]

	holder := memcachedWrapper.New(&backend)
	memcachedServer := memcached.NewServer("127.0.0.1:11211", &holder)

	var wg sync.WaitGroup
	listenAndServe(&wg, memcachedServer.ListenAndServe)

	for _, server := range servers {
		listenAndServe(&wg, server.ListenAndServe)
	}

	wg.Wait()
}
