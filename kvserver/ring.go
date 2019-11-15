package server

import (
	hash "crypto/sha1"
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
	myNode node
	fingerTable [M]node
}

func addr2node(addr address) node {
	return node{
		addr: addr,
		pos: position(hash.Sum([]byte(addr))),
	}
}

func (r *chordRing) lookup(key string) (node address) {
	// TODO: calculate the right position
	// keyPos := position([M/8]byte{})

	// TODO: based on the finger table and do a lookup
	// repeat until (predecessor, successor) is found
	return r.myNode.addr
}

func newChordRing(myAddress address) chordRing {
	return chordRing{
		myNode: addr2node(myAddress),
	}
}

func (r *chordRing) fillFingerTable(servers []ChordServer) {
	// TODO: fill the first entry of the finger table (i.e. the successor)
}
