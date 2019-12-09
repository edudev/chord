package server

import (
	"math/big"
	"sort"
)

func CreatePerfectFingertable(forNodeAddr string, otherNodes []string) (fingerTable [M]string) {
	forNode := addr2node(address(forNodeAddr))
	nodes := make([]node, len(otherNodes))
	for i, addr := range otherNodes {
		nodes[i] = addr2node(address(addr))
	}
	sortServersByNodePosition(nodes)
	for k := uint(0); k < M; k++ {
		q := calculateFingerTablePosition(forNode, k)
		fingerTable[k] = string(successorToPositionInServers(nodes, q).addr)
	}
	return
}

func CreatePerfectSuccessorList(forNodeAddr string, otherNodes []string) (successors []string) {
	successors = make([]string, R+1)
	forNode := addr2node(address(forNodeAddr))
	nodes := make([]node, len(otherNodes))
	for i, addr := range otherNodes {
		nodes[i] = addr2node(address(addr))
	}
	sortServersByNodePosition(nodes)

	for i, node := range nodes {
		if node.addr == forNode.addr {
			for j := 0; j < len(successors); j++ {
				successors[j] = string(nodes[(i+j+1)%len(nodes)].addr)
			}
		}
	}

	return
}

type sortByPosition []node

func (a sortByPosition) Len() int      { return len(a) }
func (a sortByPosition) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a sortByPosition) Less(i, j int) bool {
	return cmpPosition(a[i].pos, a[j].pos) < 0
}

func sortServersByNodePosition(servers []node) {
	sort.Sort(sortByPosition(servers))
}

// finds the successor to the given position in the given array of servers
func successorToPositionInServers(servers []node, p position) node {
	for _, s := range servers {
		// find the next node that is larger than p
		if cmpPosition(p, s.pos) < 0 {
			return s
		}
	}
	// the succesor is the next element after wrapping,
	// therefore the first element in the sorted array.
	return servers[0]
}

func calculateFingerTablePosition(forNode node, k uint) position {
	// n is our node position (as in the paper)
	n := forNode.pos
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
