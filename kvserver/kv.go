package server

import (
	"errors"
)

type chordKV struct {
	table map[string]string
}

func (c *chordKV) get(key string) (string, error) {
	/*Check for existing key*/
	if val, ok := c.table[key]; ok {
		return val, nil
	}
	/*Key not found*/
	err := errors.New("Key not found")
	return key, err
}

func (c *chordKV) set(key, value string) error {
	/*Check for existing key*/
	if _, ok := c.table[key]; ok {
		err := errors.New("Key already exists")
		return err
	}
	/*Save key/value*/
	c.table[key] = value
	return nil
}

func (c *chordKV) delete(key string) error {
	if _, ok := c.table[key]; ok {
		delete(c.table, key)
		return nil
	}
	/*Key not found*/
	err := errors.New("Key not found")
	return err
}

/* Create a new instance of chordKV */
func newChordKV() chordKV {
	return chordKV{
		table: make(map[string]string),
	}
}
