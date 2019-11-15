package main

import (
	"log"

	kvserver "github.com/edudev/chord/kvserver"
	memcached_wrapper "github.com/edudev/chord/memcached"
	memcached "github.com/mattrobenolt/go-memcached"
)

func main() {
	backend := kvserver.New()
	holder := memcached_wrapper.New(&backend)
	server := memcached.NewServer(":11211", &holder)

	log.SetPrefix("TRACE: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
	log.Println("log initialised")

	server.ListenAndServe()
}
