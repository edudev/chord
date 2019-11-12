package server

type Chord interface {
    Get(string) (string, error)
    Set(string, string) error
}

type ChordServer struct {}

func (c *ChordServer) Get(key string) (string, error) {
    // TODO: implement
}

func (c *ChordServer) Set(key string, value string) error {
    // TODO: implement
}
