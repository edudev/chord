package main

import (
    memcached "github.com/mattrobenolt/go-memcached"
)

type Cache struct {}

func (c *Cache) Get(key string) (response memcached.MemcachedResponse) {
    if key == "hello" {
        item := &memcached.Item{
            Key: key,
            Value: []byte("world"),
        }

        response = &memcached.ItemResponse{Item: item}
    } else {
        response = &memcached.ClientErrorResponse{
            Reason: memcached.NotFound.Error(),
        }
    }

    return response
}

func main() {
    server := memcached.NewServer(":11211", &Cache{})
    server.ListenAndServe()
}
