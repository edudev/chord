package server

import (
	"context"
	hash "crypto/sha1"

	"google.golang.org/grpc"
	"github.com/golang/protobuf/ptypes/empty"
)

const (
	M uint = hash.Size * 8
)

type position [M/8]byte

type node struct {
	addr address
	pos position
}

type chordRing struct {
	server *ChordServer

	myNode node
	fingerTable [M]node

	UnimplementedChordRingServer
}

func addr2node(addr address) node {
	return node{
		addr: addr,
		pos: position(hash.Sum([]byte(addr))),
	}
}

func (r *chordRing) lookup(key string) (addr address) {
	// TODO: calculate the right position
	// keyPos := position([M/8]byte{})
	// node := r.findSuccessor(keyPos)
	// return node.addr

	// TODO: based on the finger table and do a lookup
	// repeat until (current, successor) is found
	return r.myNode.addr
}

func newChordRing(server *ChordServer, myAddress address, grpcServer *grpc.Server) chordRing {
	ring := chordRing{
		server: server,
		myNode: addr2node(myAddress),
	}

	RegisterChordRingServer(grpcServer, &ring)

	return ring
}

func (r *chordRing) getClient(addr address) (client ChordRingClient) {
	conn := r.server.getClientConn(addr)
	client = NewChordRingClient(conn)
	return
}

func (r *chordRing) fillFingerTable(servers []ChordServer) {
	// TODO: fill the first entry of the finger table (i.e. the successor)
}

func (r *chordRing) findSuccessor(keyPos position) (successor node) {
	// TODO: find predecessor, then return his successor
	return node{}
}

func (r *chordRing) findPredecessor(keyPos position) (predecessor node) {
	// TODO: iteratively ask nodes (rpc) until we get a node for which keyPos
	// is between (node, nodeSuccessor]
	return node{}
}

func (c *chordRing) GetSuccessor(ctx context.Context, in *empty.Empty) (*RPCNode, error) {
	// TODO: return the current node's successor
	return &RPCNode{}, nil
}

func (c *chordRing) ClosestPrecedingFinger(ctx context.Context, in *LookupRequest) (*RPCNode, error) {
	// TODO: return the successor node, for now
	return &RPCNode{}, nil
}
