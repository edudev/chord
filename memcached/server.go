package memcached

import (
	"log"

	memcached "github.com/mattrobenolt/go-memcached"
)

type MemcachedBackend interface {
	Get(string) (string, error)
	Set(string, string) error
	Delete(string) error
}

type MemcachedServer struct {
	keyvalue MemcachedBackend
}

func New(backend MemcachedBackend) MemcachedServer {
	return MemcachedServer{
		keyvalue: backend,
	}
}

func (c *MemcachedServer) Get(key string) (response memcached.MemcachedResponse) {

	if val, err := c.keyvalue.Get(key); err == nil {
		item := &memcached.Item{
			Key:   key,
			Value: []byte(val),
		}
		response = &memcached.ItemResponse{Item: item}
	} else {
		response = nil
		log.Printf("GET FAILED: %s | key received: %s\n", err, key)
	}
	return response
}

func (c *MemcachedServer) Set(toadd *memcached.Item) (response memcached.MemcachedResponse) {
	/*Note:In case of correct add we send back the Item, generic error otherwhise*/
	if err := c.keyvalue.Set(toadd.Key, string(toadd.Value[:])); err == nil {
		response = nil
	} else {
		response = &memcached.ClientErrorResponse{
			Reason: memcached.Error.Error(),
		}
		log.Printf("SET FAILED: %s | key-value pair received: {%s : %s}\n", response, toadd.Key, toadd.Value)
	}
	return response
}

func (c *MemcachedServer) Delete(key string) (response memcached.MemcachedResponse) {
	if err := c.keyvalue.Delete(key); err == nil {
		response = nil
	} else {
		response = &memcached.ClientErrorResponse{
			Reason: memcached.Error.Error(),
		}
		log.Printf("DELETE FAILED: %s | key received: %s\n", err, key)
	}
	return response
}
