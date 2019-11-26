package server

import (
	"log"
	"context"
	hash "crypto/sha1"
	"math/big"
	"sort"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

const (
	// M as it is used in the paper. M specifies the size of the identifier ring,
	// which is 2^M in size (M specifies the amount of bits in an identifier).
	// It is chosen to be the amount of bits we receive from the hashing function.
	M uint = hash.Size * 8
)

// a position should be treated as an opague value.
// however, for the finger table operations, it is useful to treat it as a number.
type position big.Int

// compares to positions / spaceship operator
func cmpPosition(a position, b position) int {
	return ((*big.Int)(&a)).Cmp((*big.Int)(&b))
}

type node struct {
	addr address
	pos  position
}

type chordRing struct {
	server *ChordServer

	myNode      node
	fingerTable [M]node
	lock        sync.RWMutex

	UnimplementedChordRingServer
}

func bytes2position(bytes []byte) (pos position) {
	// the ordering of bytes shouldn't matter as long as it is consistent.
	// SetBytes treats the number as unsigned.
	(*big.Int)(&pos).SetBytes(bytes)
	return
}

func key2position(key string) position {
	h := hash.Sum([]byte(key))
	return bytes2position(h[:])
}

func position2bytes(pos position) []byte {
	return (*big.Int)(&pos).Bytes()
}

func addr2node(addr address) node {
	return node{
		addr: addr,
		pos:  key2position(string(addr)),
	}
}

func (r *chordRing) lookup(key string) (addr address, err error) {
	keyPos := key2position(key)

	node, err := r.findSuccessor(keyPos)
	if err != nil {
		return
	}

	addr = node.addr
	return
}

func newChordRing(server *ChordServer, myAddress address, grpcServer *grpc.Server) *chordRing {
	ring := &chordRing{
		server: server,
		myNode: addr2node(myAddress),
	}

	RegisterChordRingServer(grpcServer, ring)

	return ring
}

func (r *chordRing) getClient(addr address) (client ChordRingClient) {
	log.Printf("connecting to ring %v", addr)
	conn := r.server.getClientConn(addr)
	client = NewChordRingClient(conn)
	return
}

// calculates n + 2^k mod (2^M - 1)
func (r *chordRing) calculateFingerTablePosition(k uint) position {
	// n is our node position (as in the paper)
	n := r.myNode.pos
	// one = 1
	one := big.NewInt(1)
	// max = 2^M - 1
	var max big.Int
	max.Lsh(one, M)
	max.Sub(&max, one)
	var q big.Int
	// q = 2^k
	q.Lsh(one, k)
	// q = n + q = n + 2^k
	q.Add(&q, (*big.Int)(&n))
	// q = q AND max = q mod max = q mod (2^M - 1) = n + 2^k mod (2^M - 1)
	q.And(&q, &max)
	return position(q)
}

// finds the successor to the given position in the given array of servers
func (r *chordRing) successorToPositionInServers(servers []ChordServer, p position) ChordServer {
	for _, s := range servers {
		// find the next node that is larger than p
		if cmpPosition(p, s.ring.myNode.pos) < 0 {
			return s
		}
	}
	// the succesor is the next element after wrapping,
	// therefore the first element in the sorted array.
	return servers[0]
}

type sortByPosition []ChordServer

func (a sortByPosition) Len() int      { return len(a) }
func (a sortByPosition) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a sortByPosition) Less(i, j int) bool {
	return cmpPosition(a[i].ring.myNode.pos, a[j].ring.myNode.pos) < 0
}

func SortServersByNodePosition(servers []ChordServer) {
	sort.Sort(sortByPosition(servers))
}

func (r *chordRing) fillFingerTable(servers []ChordServer) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for k := uint(0); k < M; k++ {
		q := r.calculateFingerTablePosition(k)
		r.fingerTable[k] = r.successorToPositionInServers(servers, q).ring.myNode
	}
}

func (r *chordRing) findSuccessor(keyPos position) (successor *node, e error) {
	log.Printf("getting predecessor of %v", keyPos)
	predecessor, e := r.findPredecessor(keyPos)
	if e != nil {
		return nil, e
	}
	log.Printf("getting successor (final) of %v", predecessor.addr)
	successorRPC, e := r.getClient(predecessor.addr).GetSuccessor(context.Background(), new(empty.Empty))
	if e != nil {
		return nil, e
	}
	n := successorRPC.node()
	return &n, nil
}

// checks whether keyPos € (p, successor]
func isSuccessorResponsibleForPosition(p position, keyPos position, successor position) bool {
	if cmpPosition(p, successor) <= 0 {
		return cmpPosition(keyPos, p) > 0 && cmpPosition(keyPos, successor) <= 0
	}

	return cmpPosition(keyPos, p) > 0 || cmpPosition(keyPos, successor) <= 0
}

func (r *chordRing) findPredecessor(keyPos position) (predecessor *node, e error) {
	// iteratively ask nodes (rpc) until we get a node for which keyPos
	// is between (node, nodeSuccessor]
	n := r.myNode
	r.lock.RLock()
	successor := r.fingerTable[0]
	r.lock.RUnlock()
	for !isSuccessorResponsibleForPosition(n.pos, keyPos, successor.pos) {
		log.Printf("getting ClosestPrecedingFinger for %v from %v", keyPos, n.addr)
		nRPC, e := r.getClient(n.addr).ClosestPrecedingFinger(context.Background(), &LookupRequest{Position: position2bytes(keyPos)})
		if e != nil {
			return nil, e
		}
		n = nRPC.node()
		log.Printf("getting successor for %v", n.addr)
		successorRPC, e := r.getClient(n.addr).GetSuccessor(context.Background(), new(empty.Empty))
		if e != nil {
			return nil, e
		}
		successor = successorRPC.node()
	}
	return &n, nil
}

func (rpc *RPCNode) node() node {
	return node{
		addr: address(rpc.GetAddress()),
		pos:  bytes2position(rpc.GetPosition()),
	}
}

func (n node) rpcNode() *RPCNode {
	return &RPCNode{
		Address: string(n.addr),
		Position: position2bytes(n.pos),
	}
}

// finds the closest predecessor for keyPosition in our local finger table
func (r *chordRing) fingerTableClosestPredecessor(keyPosition position) node {
	r.lock.RLock()
	defer r.lock.RUnlock()
	for i := uint(0); i < M-1; i++ {
		if isSuccessorResponsibleForPosition(r.fingerTable[i].pos, keyPosition, r.fingerTable[i+1].pos) {
			return r.fingerTable[i]
		}
	}
	return r.fingerTable[M-1]
}

func (r *chordRing) GetSuccessor(ctx context.Context, in *empty.Empty) (*RPCNode, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.fingerTable[0].rpcNode(), nil
}

func (r *chordRing) ClosestPrecedingFinger(ctx context.Context, in *LookupRequest) (*RPCNode, error) {
	keyPosition := bytes2position(in.GetPosition())
	return r.fingerTableClosestPredecessor(keyPosition).rpcNode(), nil
}
