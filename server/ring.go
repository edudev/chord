package server

import (
	"context"
	hash "crypto/sha1"
	"log"
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
	// R is supposed to be log_2(N) where N is the number of nodes.
	// however, we don't know the number of nodes...
	R uint = 5
)

// TODO lock on predecessor
// TODO fill successors during join
// TODO create RPC wrapper methods that do the type conversions
// TODO keep immediate successor only in fingerTable not also in successors
// TODO use a list instead of array for successors
// TODO let stabilize goroutine receive explicit fix notifications
// TODO check mutex recursive

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
	// TODO `predecessor` should really be an optional instead of a pointer
	predecessor        *node
	nextFingerFixIndex uint
	// successors keeps track of R successors to this node
	successors [R](*node)
	// TODO rename to fingerTableLock
	lock           sync.RWMutex
	successorsLock sync.RWMutex

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
		server:             server,
		myNode:             addr2node(myAddress),
		nextFingerFixIndex: M - 1,
		predecessor:        nil,
	}

	RegisterChordRingServer(grpcServer, ring)

	return ring
}

// joins the given chord ring.
// otherNodeAddr can be nil to indicate there is node to join to (this is the first node)
func (r *chordRing) join(otherNodeAddr *address) (e error) {
	if otherNodeAddr != nil {
		// there is an existing ring to join to
		return r.initFingerTable(*otherNodeAddr)
	}
	// no ring to join to, we are the 'first' node
	r.lock.Lock()
	defer r.lock.Unlock()
	r.predecessor = &r.myNode
	for i := uint(0); i < M; i++ {
		r.fingerTable[i] = r.myNode
	}
	return nil
}

// to be called if we realize the node at the given address is gone
func (r *chordRing) nodeDied(addr address) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.successorsLock.Lock()
	defer r.successorsLock.Unlock()
	if r.predecessor != nil && addr == r.predecessor.addr {
		// TODO may want to trigger a predecessor update immediately
		r.predecessor = nil
	}
	// update successors
	for i, successor := range r.successors {
		if successor.addr == addr {
			// TODO hint to fix_successors
			r.successors[i] = nil
		}
	}

	// fix fingerTable, excluding successor
	for i := M - 1; i > 0; i-- {
		if r.fingerTable[i].addr == addr {
			var s node
			if i == M-1 {
				s = r.myNode
			} else {
				s = r.fingerTable[i+1]
			}
			r.fingerTable[i] = s
			// TODO hint stabilize
		}
	}

	// fix successor
	if r.fingerTable[0].addr == addr {
		// find the next successor
		var newSuccessor *node
		newSuccessor = nil
		for _, successor := range r.successors {
			if successor != nil {
				newSuccessor = successor
				break
			}
		}
		if newSuccessor == nil {
			// find the next sucessor finger table
			// since the finger table is correct as per the above loop,
			// we can just to the following
			newSuccessor = &r.fingerTable[1]
		}
		r.fingerTable[0] = *newSuccessor
	}
}

func (r *chordRing) initFingerTable(nodeToJoin address) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.predecessor = nil
	successorRPC, e := r.getClient(nodeToJoin).FindSuccessor(context.Background(), &LookupRequest{Position: position2bytes(r.myNode.pos)})
	if e != nil {
		return e
	}
	successor := successorRPC.node()
	for i := uint(0); i < M; i++ {
		r.fingerTable[i] = successor
	}
	return nil
}

// checks whether element € (left, right), respecting wrapping
func isPosInRangExclusive(left position, element position, right position) bool {
	if cmpPosition(left, right) <= 0 {
		return cmpPosition(element, left) > 0 && cmpPosition(element, right) < 0
	}

	return cmpPosition(element, left) > 0 || cmpPosition(element, right) < 0
}

// the stabilize function as defined in the paper
// TODO call periodically with period 150ms in a go routine, kill using channel
func (r *chordRing) stabilize() error {
	r.lock.RLock()
	successor := r.fingerTable[0]
	r.lock.RUnlock()
	xRPC, e := r.getClient(successor.addr).GetPredecessor(context.Background(), new(empty.Empty))
	if e != nil {
		return e
	}
	x := xRPC.node()
	if x.addr == "" { // successor has no predecessor?
		return nil
	}
	if isPosInRangExclusive(r.myNode.pos, x.pos, successor.pos) {
		r.lock.Lock()
		r.fingerTable[0] = x
		// TODO add to successor list? -> push_front
		r.lock.Unlock()
		successor = x
		// TODO/INTERESTING shall we also replace other entries occupied by the same node in the finger table?
	}
	_, e = r.getClient(successor.addr).Notify(context.Background(), r.myNode.rpcNode())
	return nil
}

// fixes the successor list if need be
// TODO should be called in stabilize
func (r *chordRing) fixSuccessors() error {
	// TODO implement
	return nil
}

func (r *chordRing) getClient(addr address) (client ChordRingClient) {
	log.Printf("connecting to ring %v", addr)
	conn := r.server.getClientConn(addr)
	client = NewChordRingClient(conn)
	return
}

// sets the nextFingerFixIndex to the next value and returns the old one
func (r *chordRing) setFixIndexToNextIndex() (k uint) {
	k = r.nextFingerFixIndex
	if r.nextFingerFixIndex == 0 {
		r.nextFingerFixIndex = M - 1
	} else {
		r.nextFingerFixIndex--
	}
	return
}

// the fix finger function according to the paper
// TODO should be called periodically. see comments on `stabilize`
func (r *chordRing) fixFingers() error {
	k := r.setFixIndexToNextIndex()
	n := r.calculateFingerTablePosition(k)
	successor, e := r.findSuccessor(n)
	if e != nil {
		return e
	}
	r.lock.Lock()
	defer r.lock.Unlock()
	r.fingerTable[k] = *successor
	return nil
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

// TODO don't return a pointer
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
		Address:  string(n.addr),
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

func (r *chordRing) FindSuccessor(ctx context.Context, in *LookupRequest) (*RPCNode, error) {
	position := bytes2position(in.GetPosition())
	successor, e := r.findSuccessor(position)
	if e != nil {
		return nil, e
	}
	return successor.rpcNode(), nil
}

func (r *chordRing) GetPredecessor(ctx context.Context, in *empty.Empty) (*RPCNode, error) {
	if r.predecessor == nil {
		return &RPCNode{Address: "", Position: []byte{}}, nil
	}
	return r.predecessor.rpcNode(), nil
}

func (r *chordRing) Notify(ctx context.Context, in *RPCNode) (*empty.Empty, error) {
	nPrime := in.node()
	if r.predecessor == nil || isPosInRangExclusive(r.predecessor.pos, nPrime.pos, r.myNode.pos) {
		r.predecessor = &nPrime
	}
	return new(empty.Empty), nil
}
