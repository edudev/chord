package server

type address string

type ChordServer struct {
	ring *chordRing
	kvstore *chordKV
}

func (s *ChordServer) Get(key string) (string, error) {
	remoteNode := s.ring.lookup(key)
	return remoteGet(remoteNode, key)
}

func (s *ChordServer) Set(key string, value string) error {
	remoteNode := s.ring.lookup(key)
	return remoteSet(remoteNode, key, value)
}

func (s *ChordServer) Delete(key string) error {
	remoteNode := s.ring.lookup(key)
	return remoteDelete(remoteNode, key)
}

/* Create a new instance of ChordServer */
func New(myAddress string) ChordServer {
	ring := newChordRing(address(myAddress))
	kvstore := newChordKV()

	return ChordServer{
		ring: &ring,
		kvstore: &kvstore,
	}
}

func (s *ChordServer) ListenAndServe() error {
	// gRPC server listen
	return nil
}

func remoteGet(remoteNode address, key string) (string, error) {
	// TODO: execute rpc
	return "", nil
}

func remoteSet(remoteNode address, key string, value string) error {
	// TODO: execute rpc
	return nil
}

func remoteDelete(remoteNode address, key string) error {
	// TODO: execute rpc
	return nil
}

// TODO: remove this in the future
func (s *ChordServer) FillFingerTable(servers []ChordServer) {
	s.ring.fillFingerTable(servers)
}
