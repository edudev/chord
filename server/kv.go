package server

import (
	"context"
	"errors"
	"log"
	"sync"

	"google.golang.org/grpc"
)

type chordKV struct {
	server *ChordServer
	table  map[string]string
	lock   sync.RWMutex

	UnimplementedChordKVServer
}

func (c *chordKV) localGet(key string) (string, error) {
	/*Check for existing key*/
	if val, ok := c.table[key]; ok {
		return val, nil
	}
	/*Key not found*/
	err := errors.New("Key not found")
	return key, err
}

func (c *chordKV) localSet(key, value string) error {
	/*Save key/value*/
	c.table[key] = value
	return nil
}

func (c *chordKV) localDelete(key string) error {
	if _, ok := c.table[key]; ok {
		delete(c.table, key)
		return nil
	}
	/*Key not found*/
	err := errors.New("Key not found")
	return err
}

/* Create a new instance of chordKV */
func newChordKV(server *ChordServer, grpcServer *grpc.Server) *chordKV {
	kvstore := &chordKV{
		server: server,
		table:  make(map[string]string),
	}

	RegisterChordKVServer(grpcServer, kvstore)

	return kvstore
}

func (c *chordKV) getClient(addr address) (client ChordKVClient) {
	log.Printf("connecting to kv %v", addr)
	conn := c.server.getClientConn(addr)
	client = NewChordKVClient(conn)
	return
}

func (c *chordKV) Get(ctx context.Context, in *GetRequest) (reply *GetReply, err error) {
	key := in.GetKey()
	c.lock.RLock()
	defer c.lock.RUnlock()
	value, err := c.localGet(key)

	if err != nil {
		return
	}

	return &GetReply{Value: value}, nil
}

func (c *chordKV) Set(ctx context.Context, in *SetRequest) (reply *SetReply, err error) {
	key := in.GetKey()
	value := in.GetValue()

	c.lock.Lock()
	defer c.lock.Unlock()
	if err = c.localSet(key, value); err != nil {
		return
	}

	reply = &SetReply{}
	return
}

func (c *chordKV) Delete(ctx context.Context, in *DeleteRequest) (reply *DeleteReply, err error) {
	key := in.GetKey()

	c.lock.Lock()
	defer c.lock.Unlock()
	if err = c.localDelete(key); err != nil {
		return
	}

	reply = &DeleteReply{}
	return
}

func (c *chordKV) remoteGet(remoteNode address, key string) (value string, err error) {
	log.Printf("doing a remote get on %v for ", remoteNode, key)
	keyPos := key2position(key)
	getRequest := &GetRequest{Key: key, Position: position2bytes(keyPos)}
	getReply, err := c.getClient(remoteNode).Get(context.Background(), getRequest)
	if err != nil {
		return
	}

	return getReply.GetValue(), nil
}

func (c *chordKV) remoteSet(remoteNode address, key string, value string) (err error) {
	keyPos := key2position(key)
	setRequest := &SetRequest{
		Key:      key,
		Value:    value,
		Position: position2bytes(keyPos),
	}

	_, err = c.getClient(remoteNode).Set(context.Background(), setRequest)

	if err != nil {
		return
	}

	return
}

func (c *chordKV) remoteDelete(remoteNode address, key string) (err error) {
	keyPos := key2position(key)
	deleteRequest := &DeleteRequest{
		Key:      key,
		Position: position2bytes(keyPos),
	}

	_, err = c.getClient(remoteNode).Delete(context.Background(), deleteRequest)

	if err != nil {
		return
	}

	return
}

func (c *chordKV) Stop() {
}
