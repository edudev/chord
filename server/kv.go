package server

import (
	"context"
	"errors"

	"google.golang.org/grpc"
)

type chordKV struct {
	server *ChordServer
	table  map[string]string

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
func newChordKV(server *ChordServer, grpcServer *grpc.Server) chordKV {
	kvstore := chordKV{
		server: server,
		table:  make(map[string]string),
	}

	RegisterChordKVServer(grpcServer, &kvstore)

	return kvstore
}

func (c *chordKV) getClient(addr address) (client ChordKVClient) {
	conn := c.server.getClientConn(addr)
	client = NewChordKVClient(conn)
	return
}

func (c *chordKV) Get(ctx context.Context, in *GetRequest) (*GetReply, error) {
	// TODO: do a local get
	return &GetReply{}, nil
}

func (c *chordKV) Set(ctx context.Context, in *SetRequest) (*SetReply, error) {
	// TODO: do a local set
	return &SetReply{}, nil
}

func (c *chordKV) Delete(ctx context.Context, in *DeleteRequest) (*DeleteReply, error) {
	// TODO: do a local delete
	return &DeleteReply{}, nil
}

func (c *chordKV) remoteGet(remoteNode address, key string) (string, error) {
	// TODO: execute rpc
	return "", nil
}

func (c *chordKV) remoteSet(remoteNode address, key string, value string) error {
	// TODO: execute rpc
	return nil
}

func (c *chordKV) remoteDelete(remoteNode address, key string) error {
	// TODO: execute rpc
	return nil
}
