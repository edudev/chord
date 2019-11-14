package server

import (
	"errors"
)

type Chord interface {
	Get(string) (string, error)
	Set(string, string) error
	Init()
	isNill() bool
}

type ChordServer struct {
	table map[string]string
}

func (c *ChordServer) Get(key string) (string, error) {
	if c.isNill() {
		err := errors.New("Table not initialised")
		return key, err
	}
	/*Check for existing key*/
	if val, ok := c.table[key]; ok {
		return val, nil
	}
	/*Key not found*/
	err := errors.New("Key not found")
	return key, err
}

func (c *ChordServer) Set(key string, value string) error {
	if c.isNill() {
		c.Init()
	}
	/*Check for existing key*/
	if _, ok := c.table[key]; ok {
		err := errors.New("Key already exists")
		return err
	}
	/*Save key/value*/
	c.table[key] = value
	return nil
}

/*Initialise map*/
func (c *ChordServer) Init() {
	if c.isNill() {
		c.table = make(map[string]string)
	}
}

/*Check if map is not initialised*/
func (c *ChordServer) isNill() bool {
	if c.table == nil {
		return true
	}
	return false
}
