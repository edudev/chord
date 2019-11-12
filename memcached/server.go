package memcached

import (
    memcached "github.com/mattrobenolt/go-memcached"
)

type MemcachedServer struct {}

func (c *MemcachedServer) Get(key string) (response memcached.MemcachedResponse) {
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
    server := memcached.NewServer(":11211", &MemcachedServer{})
    server.ListenAndServe()
}
