package main

import (
    memcached "github.com/mattrobenolt/go-memcached"
    memcached_wrapper "github.com/edudev/chord/memcached"
)

func main() {
    holder := memcached_wrapper.MemcachedServer{}
    server := memcached.NewServer(":11211", &holder)
    server.ListenAndServe()
}
