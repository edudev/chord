package server

import (
	"context"
	hash "crypto/sha1"
	fmt "fmt"
	"log"
	"math"
	"math/big"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

const (
	TICK_STABILISE   = 100 * time.Millisecond
	TICK_FIX_FINGERS = 500 * time.Millisecond

	// M as it is used in the paper. M specifies the size of the identifier ring,
	// which is 2^M in size (M specifies the amount of bits in an identifier).
	// It is chosen to be the amount of bits we receive from the hashing function.
	M uint = hash.Size * 8
	// R is supposed to be log_2(N) where N is the number of nodes.
	// however, we don't know the number of nodes...
	R uint = 5
)

// TODO check mutex recursive

// a position should be treated as an opague value.
// however, for the finger table operations, it is useful to treat it as a number.
type position big.Int

func (p position) String() string {
	return fmt.Sprintf("%049v", ((*big.Int)(&p)).String())[:5]
}

// compares to positions / spaceship operator
func cmpPosition(a position, b position) int {
	return ((*big.Int)(&a)).Cmp((*big.Int)(&b))
}

type node struct {
	addr address
	pos  position
}

func (n node) String() string {
	return fmt.Sprintf("%v@%v", n.addr, n.pos)
}

type chordRing struct {
	server *ChordServer

	myNode node

	fingerTable        [M]node
	fingerTableLock    sync.RWMutex
	nextFingerFixIndex uint

	// `predecessor` should really be an optional instead of a pointer
	predecessor     *node
	predecessorLock sync.RWMutex

	successors     [](node)
	successorsLock sync.RWMutex

	stopped         chan bool
	stabiliseQueue  chan bool
	fixFingersQueue chan uint

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

		stopped:         make(chan bool, 2),
		stabiliseQueue:  make(chan bool, 2),
		fixFingersQueue: make(chan uint, 2),
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

	r.predecessorLock.Lock()
	r.predecessor = &r.myNode
	r.predecessorLock.Unlock()

	r.fingerTableLock.Lock()
	for i := uint(0); i < M; i++ {
		r.fingerTable[i] = r.myNode
	}
	r.fingerTableLock.Unlock()

	return nil
}

// to be called if we realize the node at the given address is gone
func (r *chordRing) nodeDied(addr address) {
	log.Printf("Node died at address: ", addr)
	r.predecessorLock.Lock()
	defer r.predecessorLock.Unlock()
	r.fingerTableLock.Lock()
	defer r.fingerTableLock.Unlock()
	r.successorsLock.Lock()
	defer r.successorsLock.Unlock()

	if r.predecessor != nil && addr == r.predecessor.addr {
		// TODO may want to trigger a predecessor update immediately
		r.predecessor = nil
	}
	// update successors
	successorsAffected := false
	for {
		breaK := true
		for i, successor := range r.successors {
			if successor.addr == addr {
				r.successors = append(r.successors[:i], r.successors[i+1:]...)
				successorsAffected = true
				breaK = false
				break
			}
		}
		if breaK {
			break
		}
	}
	if successorsAffected {
		// TODO reenable later
		//r.askToStabilise()
	}

	// fix fingerTable, excluding successor
	notifiedFixFingerRoutine := false
	for i := M - 1; i > 0; i-- {
		if r.fingerTable[i].addr == addr {
			var s node
			if i == M-1 {
				s = r.myNode
			} else {
				s = r.fingerTable[i+1]
			}
			r.fingerTable[i] = s

			if !notifiedFixFingerRoutine {
				r.askToFixFingers(i)
				notifiedFixFingerRoutine = true
			}
		}
	}

	// fix successor
	if r.fingerTable[0].addr == addr {
		// find the next successor
		if len(r.successors) == 0 {
			// find the next sucessor finger table
			// since the finger table is correct as per the above loop,
			// we can just to the following
			r.fingerTable[0] = r.fingerTable[1]
		} else {
			// pop a successor from the successor list
			r.fingerTable[0], r.successors = r.successors[0], r.successors[1:]
		}
		// need to stabilise since we lost a successor list entry
		// TODO reenable later
		//r.askToStabilise()
	}
}

func (r *chordRing) initFingerTable(nodeToJoin address) error {
	successor, e := r.rpcFindSuccessor(context.Background(), nodeToJoin, r.myNode.pos)
	if e != nil {
		return e
	}

	r.predecessorLock.Lock()
	r.predecessor = nil
	r.predecessorLock.Unlock()

	r.fingerTableLock.Lock()
	for i := uint(0); i < M; i++ {
		// TODO: this isn't actually correct...
		// we should take into account the positions
		r.fingerTable[i] = successor
	}
	r.fingerTableLock.Unlock()

	// TODO fill successors during join
	return nil
}

// checks whether element € (left, right), respecting wrapping
// if left == right, assume we mean the empty set, not the whole ring
func isPosInRangExclusive(left position, element position, right position) bool {
	if cmpPosition(left, right) <= 0 {
		return cmpPosition(element, left) > 0 && cmpPosition(element, right) < 0
	}

	return cmpPosition(element, left) > 0 || cmpPosition(element, right) < 0
}

// checks whether element € (left, right), respecting wrapping
// if left == right, assume we mean the whole ring, not the empty set
func isPosInRangExclusivePreferWhole(left position, element position, right position) bool {
	if cmpPosition(left, right) < 0 {
		return cmpPosition(element, left) > 0 && cmpPosition(element, right) < 0
	}

	return cmpPosition(element, left) > 0 || cmpPosition(element, right) < 0
}

// the stabilize function as defined in the paper
func (r *chordRing) stabilize() error {
	log.Printf("[%v] stabilizing...", r.myNode)
	r.fingerTableLock.RLock()
	successor := r.fingerTable[0]
	r.fingerTableLock.RUnlock()

	xValid, x, e := r.rpcGetPredecessor(context.Background(), successor)
	if e != nil {
		r.nodeDied(successor.addr)
		return e
	}
	if xValid { // successor has no predecessor?
		if isPosInRangExclusivePreferWhole(r.myNode.pos, x.pos, successor.pos) {
			log.Printf("[%v] A new node joined as our successor: %v", r.myNode, x)
			r.fingerTableLock.Lock()
			r.fingerTable[0] = x
			r.fingerTableLock.Unlock()

			/*
				if successor.addr != r.myNode.addr {
					// add our old successor as the first entry to the successor list
					r.successorsLock.Lock()
					r.successors = append([]node{successor}, r.successors...)
					if len(r.successors) > int(R) {
						r.successors = r.successors[:len(r.successors)-1]
					}
					r.successorsLock.Unlock()
				}
				// TODO/INTERESTING shall we also replace other entries occupied by the same node in the finger table?
			*/
		}
	}

	e = r.rpcNotify(context.Background(), successor, r.myNode)
	// e = r.fixSuccessors()
	log.Printf("[%v] done stabilising", r.myNode)
	return e
}

// fixes the successor list if need be
func (r *chordRing) fixSuccessors() error {
	// very dumb version for now: will iteratively add new nodes to successor list
	r.successorsLock.RLock()
	numSuccessors := len(r.successors)
	r.successorsLock.RUnlock()
	if numSuccessors == int(R) {
		// we're all good (Y)
		return nil
	}

	log.Printf("[%v] Need to fix our successors, as we have only %v", r.myNode, numSuccessors)
	// TODO more fine grained locking
	r.successorsLock.Lock()
	defer r.successorsLock.Unlock()
	var nextNodeToAsk node
	if len(r.successors) == 0 {
		r.fingerTableLock.RLock()
		nextNodeToAsk = r.fingerTable[0]
		r.fingerTableLock.RUnlock()
	} else {
		nextNodeToAsk = r.successors[len(r.successors)-1]
	}
	successorToAdd, e := r.rpcGetSuccessor(context.Background(), nextNodeToAsk)
	if e != nil {
		r.nodeDied(nextNodeToAsk.addr)
		// a new stabilise run will be scheduled by nodeDied
		return e
	}
	if successorToAdd.addr == r.myNode.addr {
		// don't add ourself
		return nil
	}
	r.successors = append(r.successors, successorToAdd)
	if len(r.successors) != int(R) {
		// immediately trigger another run in case our list not full yet
		// TODO disabled for now
		//r.askToStabilise()
	}
	return nil
}

func (r *chordRing) getClient(addr address) (client ChordRingClient) {
	conn := r.server.getClientConn(addr)
	client = NewChordRingClient(conn)
	return
}

// sets the nextFingerFixIndex to the next value and returns the old one
func (r *chordRing) rotateNextFingerFixIndex() (k uint) {
	k = r.nextFingerFixIndex
	if r.nextFingerFixIndex == 0 {
		r.nextFingerFixIndex = M - 1
	} else {
		r.nextFingerFixIndex--
	}
	return
}

func wrapNextFingerFixIndex(k uint) uint {
	if k == math.MaxUint64 {
		return M - 1
	}
	if k == M {
		return 0
	}

	return k
}

// the fix finger function according to the paper
func (r *chordRing) fixFingers(k uint) error {
	log.Printf("[%v] fixing fingers... %5v", r.myNode, k)
	n := r.calculateFingerTablePosition(k)
	finger, e := r.findSuccessor(n)
	if e != nil {
		return e
	}

	shouldRepeatFixFingers := false

	r.fingerTableLock.Lock()
	if r.fingerTable[k].addr != finger.addr {
		// TODO: k = log2(wrap(finger.pos - r.myNode.pos))
		// then all relevant fingers: k, k-1, k-2, k-3, etc.
		// until r.fingerTable[k] is correct
		// when this is done, remove shouldRepeatFixFingers
		// NOTE: this needs to handle both cases where finger is a newly
		// joined node, and the case where the old finger has died

		log.Printf("[%v] fixing finger %5v: %v", r.myNode, k, finger)
		r.fingerTable[k] = finger
		shouldRepeatFixFingers = true
	}
	r.fingerTableLock.Unlock()

	if shouldRepeatFixFingers {
		r.askToFixFingers(wrapNextFingerFixIndex(k - 1))
	}

	log.Printf("[%v] fixing fingers... %5v done", r.myNode, k)
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

func (r *chordRing) findSuccessor(keyPos position) (successor node, err error) {
	if r.fingerTable[0].addr == r.myNode.addr {
		return r.myNode, nil
	}
	log.Printf("[%v] getting predecessor of %v", r.myNode, keyPos)
	predecessor, err := r.findPredecessor(keyPos)
	if err != nil {
		return
	}
	log.Printf("[%v] getting successor (final) of %v", r.myNode, predecessor.addr)
	successor, err = r.rpcGetSuccessor(context.Background(), *predecessor)
	return
}

// checks whether keyPos € (p, successor]
// if p == successor, then return true (treat as whole ring)
func isSuccessorResponsibleForPosition(p position, keyPos position, successor position) bool {
	if cmpPosition(p, successor) < 0 {
		return cmpPosition(keyPos, p) > 0 && cmpPosition(keyPos, successor) <= 0
	}

	return cmpPosition(keyPos, p) > 0 || cmpPosition(keyPos, successor) <= 0
}

func (r *chordRing) findPredecessor(keyPos position) (predecessor *node, e error) {
	// iteratively ask nodes (rpc) until we get a node for which keyPos
	// is between (node, nodeSuccessor]
	n := r.myNode
	r.fingerTableLock.RLock()
	successor := r.fingerTable[0]
	r.fingerTableLock.RUnlock()

	for !isSuccessorResponsibleForPosition(n.pos, keyPos, successor.pos) {
		log.Printf("%v is not element of (%v, %v]", keyPos, n.pos, successor.pos)
		n, e = r.rpcClosestPrecedingFinger(context.Background(), n, keyPos)
		if e != nil {
			return nil, e
		}
		successor, e = r.rpcGetSuccessor(context.Background(), n)
		if e != nil {
			return nil, e
		}
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
	r.fingerTableLock.RLock()
	defer r.fingerTableLock.RUnlock()

	for i := int(M - 1); i >= 0; i-- {
		if isPosInRangExclusive(r.myNode.pos, r.fingerTable[i].pos, keyPosition) {
			return r.fingerTable[i]
		}
	}
	return r.myNode
}

func (r *chordRing) GetSuccessor(ctx context.Context, in *empty.Empty) (*RPCNode, error) {
	r.fingerTableLock.RLock()
	successor := r.fingerTable[0]
	r.fingerTableLock.RUnlock()
	if successor.addr == r.myNode.addr {
		log.Printf("########################## returning ourselves as successor!!! %v", r.myNode.addr)
	}
	return successor.rpcNode(), nil
}

func (r *chordRing) rpcGetSuccessor(ctx context.Context, n node) (node, error) {
	nodeRPC, err := r.getClient(n.addr).GetSuccessor(ctx, new(empty.Empty))
	if err != nil {
		return node{}, err
	}
	log.Printf("[%v] Asking %v to get their successor: its %v", r.myNode, n.addr, nodeRPC.Address)
	return nodeRPC.node(), err
}

func (r *chordRing) ClosestPrecedingFinger(ctx context.Context, in *LookupRequest) (*RPCNode, error) {
	keyPosition := bytes2position(in.GetPosition())
	return r.fingerTableClosestPredecessor(keyPosition).rpcNode(), nil
}

func (r *chordRing) rpcClosestPrecedingFinger(ctx context.Context, n node, p position) (node, error) {
	nRPC, err := r.getClient(n.addr).ClosestPrecedingFinger(ctx, &LookupRequest{Position: position2bytes(p)})
	if err != nil {
		return node{}, err
	}
	log.Printf("[%v] Asking %v to get closest preceding finger of %v: its %v", r.myNode, n.addr, p, nRPC.Address)
	return nRPC.node(), nil
}

func (r *chordRing) FindSuccessor(ctx context.Context, in *LookupRequest) (*RPCNode, error) {
	position := bytes2position(in.GetPosition())
	successor, e := r.findSuccessor(position)
	if e != nil {
		return nil, e
	}
	return successor.rpcNode(), nil
}

func (r *chordRing) rpcFindSuccessor(ctx context.Context, addr address, p position) (node, error) {
	log.Printf("[%v] Asking %v to find successor of %v", r.myNode, addr, p)
	successorRPC, e := r.getClient(addr).FindSuccessor(ctx, &LookupRequest{Position: position2bytes(p)})
	if e != nil {
		return node{}, e
	}
	log.Printf("[%v] Asking %v to find successor of %v: its %v!", r.myNode, addr, p, successorRPC.Address)
	return successorRPC.node(), nil
}

func (r *chordRing) GetPredecessor(ctx context.Context, in *empty.Empty) (*PredecessorReply, error) {
	r.predecessorLock.RLock()
	predecessor := r.predecessor
	r.predecessorLock.RUnlock()

	if predecessor == nil {
		return &PredecessorReply{Valid: false}, nil
	}

	return &PredecessorReply{
		Valid: true,
		Node:  predecessor.rpcNode(),
	}, nil
}

func (r *chordRing) rpcGetPredecessor(ctx context.Context, n node) (valid bool, predecessor node, err error) {
	xRPC, e := r.getClient(n.addr).GetPredecessor(ctx, new(empty.Empty))
	if e != nil {
		return false, node{}, e
	}
	if xRPC.GetValid() {
		log.Printf("[%v] Getting predecessor of %v returned: %v", r.myNode, n.addr, xRPC.GetNode().Address)
	} else {
		log.Printf("[%v] Getting predecessor of %v: is invalid!", r.myNode, n.addr)
	}
	return xRPC.GetValid(), xRPC.GetNode().node(), nil
}

func (r *chordRing) Notify(ctx context.Context, in *RPCNode) (*empty.Empty, error) {
	nPrime := in.node()

	// TODO: maybe do an rlock & check
	// then if we don't have to update the predecessor just runlock
	// but if we do have to update the predecessor, runlock, rlock and repeat the check
	r.predecessorLock.Lock()
	if r.predecessor == nil || isPosInRangExclusivePreferWhole(r.predecessor.pos, nPrime.pos, r.myNode.pos) {
		if r.predecessor == nil || r.predecessor.addr != nPrime.addr {
			log.Printf("[%v] We were notified of the presence of %v and they are now our predecessor!", r.myNode, nPrime.addr)
		}
		r.predecessor = &nPrime
	}
	r.predecessorLock.Unlock()
	/*
		// if we are the only node in the ring: learn of the new node immediately!
		r.fingerTableLock.Lock()
		if r.fingerTable[0].addr == r.myNode.addr {
			r.fingerTable[0] = nPrime
		}
		r.fingerTableLock.Unlock()
	*/

	return new(empty.Empty), nil
}

func (r *chordRing) rpcNotify(ctx context.Context, n node, nodeToBeNotifiedOf node) (err error) {
	log.Printf("[%v] Notifying %v of the presence of %v", r.myNode, n.addr, nodeToBeNotifiedOf.addr)
	_, err = r.getClient(n.addr).Notify(ctx, nodeToBeNotifiedOf.rpcNode())
	return
}

func (r *chordRing) ListenAndServe() error {
	go r.periodicActionWorker()
	go r.periodicTicker()

	return nil
}

func (r *chordRing) Stop() {
	close(r.stopped)
}

func (r *chordRing) askToStabilise() {
	r.stabiliseQueue <- true
}

func (r *chordRing) askToFixFingers(index uint) {
	r.fixFingersQueue <- index
}

func (r *chordRing) periodicActionWorker() {
	for {
		select {
		case <-r.stopped:
			return
		case <-r.stabiliseQueue:
			if err := r.stabilize(); err != nil {
				log.Printf("Stabilise failed %v", err)
			}
		case index := <-r.fixFingersQueue:
			if err := r.fixFingers(index); err != nil {
				log.Printf("Fix fingers failed %v", err)
			}
		}
	}
}

func (r *chordRing) periodicTicker() {
	stabiliseTicker := time.NewTicker(TICK_STABILISE)
	fixFingersTicker := time.NewTicker(TICK_FIX_FINGERS)

	for {
		select {
		case <-r.stopped:
			return
		case <-stabiliseTicker.C:
			r.askToStabilise()
		case <-fixFingersTicker.C:
			k := r.rotateNextFingerFixIndex()
			r.askToFixFingers(k)
		}
	}
}
