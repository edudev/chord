package main

import (
	kvserver "github.com/edudev/chord/kvserver"
	memcached_wrapper "github.com/edudev/chord/memcached"
	memcached "github.com/mattrobenolt/go-memcached"
)

func main() {
	backend := kvserver.New()
	holder := memcached_wrapper.New(&backend)
	server := memcached.NewServer(":11211", &holder)
	server.ListenAndServe()
}
