package server

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

type address string

type ChordServer struct {
	ring    *chordRing
	kvstore *chordKV

	listener    *net.Listener
	grpcServer  *grpc.Server
	clientCache map[address]*grpc.ClientConn
}

func (s *ChordServer) Get(key string) (value string, err error) {
	remoteNode, err := s.ring.lookup(key)
	if err != nil {
		return
	}

	value, err = s.kvstore.remoteGet(remoteNode, key)
	return
}

func (s *ChordServer) Set(key string, value string) error {
	remoteNode, err := s.ring.lookup(key)
	if err != nil {
		return err
	}

	return s.kvstore.remoteSet(remoteNode, key, value)
}

func (s *ChordServer) Delete(key string) error {
	remoteNode, err := s.ring.lookup(key)
	if err != nil {
		return err
	}

	return s.kvstore.remoteDelete(remoteNode, key)
}

/* Create a new instance of ChordServer */
func New(myAddress string) (server ChordServer) {
	listener, err := net.Listen("tcp", myAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	server = ChordServer{
		listener:    &listener,
		grpcServer:  grpcServer,
		clientCache: make(map[address]*grpc.ClientConn),
	}

	ring := newChordRing(&server, address(myAddress), grpcServer)
	kvstore := newChordKV(&server, grpcServer)

	server.ring = ring
	server.kvstore = kvstore

	return server
}

func (s *ChordServer) ListenAndServe() error {
	s.ring.ListenAndServe()

	if err := s.grpcServer.Serve(*s.listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
		return err
	}

	return nil
}

func (s *ChordServer) Stop() {
	s.kvstore.Stop()
	s.ring.Stop()
}

func (s *ChordServer) Join(otherNode *string) {
	s.ring.join((*address)(otherNode))
}

func (s *ChordServer) Address() string {
	return string(s.ring.myNode.addr)
}

func (s *ChordServer) getClientConn(addr address) (conn *grpc.ClientConn) {
	if conn, ok := s.clientCache[addr]; ok {
		return conn
	}

	log.Printf("connecting to %v", addr)
	conn, err := grpc.Dial(string(addr), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	log.Printf("connected to client %v", addr)

	// TODO: close the connection at some point
	return conn
}
