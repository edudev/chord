package server

import (
	"errors"
	"sync"
)

type Chord interface {
	Get(string) (string, error)
	Set(string, string) error
	Delete(string) error
}

type ChordKV struct {
	table sync.Map
}

func (c *ChordKV) Get(key string) (string, error) {
	/*Check for existing key*/
	if val, ok := c.table.Load(key); ok {
		return val.(string), nil
	}
	/*Key not found*/
	err := errors.New("Key not found")
	return key, err
}

func (c *ChordKV) Set(key, value string) error {
	/*Save key/value*/
	c.table.Store(key, value)
	return nil
}

func (c *ChordKV) Delete(key string) error {
	if _, ok := c.table.Load(key); ok {
		c.table.Delete(key)
		return nil
	}
	/*Key not found*/
	err := errors.New("Key not found")
	return err
}

/* Create a new instance of ChordKV */
func New() ChordKV {
	return ChordKV{
		//	table: make(map[string]string),
	}
}
