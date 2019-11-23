package server

import (
	"context"
	hash "crypto/sha1"
	"math/big"
	"sort"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

const (
	M uint = hash.Size * 8
)

// a position should be treated as an opague value.
// however, for the finger table operations, it is useful to treat it as a number.
type position *big.Int

// compares to positions / spaceship operator
func cmpPosition(a position, b position) int {
	return ((*big.Int)(a)).Cmp((*big.Int)(b))
}

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
	var pos big.Int
	// the ordering of bytes shouldn't matter as long as it is consistent.
	// SetBytes treats the number as unsigned.
	hash := hash.Sum([]byte(addr))
	pos.SetBytes(hash[:])
	return node{
		addr: addr,
		pos:  position(&pos),
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

type sortByPosition []ChordServer

func (a sortByPosition) Len() int      { return len(a) }
func (a sortByPosition) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a sortByPosition) Less(i, j int) bool {
	return cmpPosition(a[i].ring.myNode.pos, a[j].ring.myNode.pos) == -1
}

func (r *chordRing) fillFingerTable(servers []ChordServer) {
	sort.Sort(sortByPosition(servers))
	for _, s := range servers {
		if cmpPosition(s.ring.myNode.pos, r.myNode.pos) == 1 {
			// s is the first node that is a successor to us.
			r.fingerTable[0] = s.ring.myNode
			break
		}
	}
	// if we couldn't find a successor yet we need to wrap.
	if (cmpPosition(servers[0].ring.myNode.pos, r.myNode.pos)) == 0 {
		// there is no successor
	} else {
		// the succesor is the next element after wrapping,
		// therefore the first element in the sorted array.
		r.fingerTable[0] = servers[0].ring.myNode
	}
	// TODO: fill the rest of the table (will be done in a separate pull/commit)
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
