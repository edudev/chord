package server

import (
	"errors"
)

type Chord interface {
	Get(string) (string, error)
	Set(string, string) error
	Delete(string) error
}

type ChordKV struct {
	table map[string]string
}

func (c *ChordKV) Get(key string) (string, error) {
	/*Check for existing key*/
	if val, ok := c.table[key]; ok {
		return val, nil
	}
	/*Key not found*/
	err := errors.New("Key not found")
	return key, err
}

func (c *ChordKV) Set(key string, value string) error {
	/*Check for existing key*/
	if _, ok := c.table[key]; ok {
		err := errors.New("Key already exists")
		return err
	}
	/*Save key/value*/
	c.table[key] = value
	return nil
}

func (c *ChordKV) Delete(key string) error {
	// TODO: implement
	return errors.New("Not implemented")
}

/* Create a new instance of ChordKV */
func New() ChordKV {
	return ChordKV{
		table: make(map[string]string),
	}
}
