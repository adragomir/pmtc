package util

import (
	"reflect"
)

type Callbacker struct {
	Callbacks map[string][]interface{}
}

func NewCallbacker() *Callbacker {
	return &Callbacker{
		Callbacks: make(map[string][]interface{}),
	}
}

func (c *Callbacker) AddCb(t string, cb interface{}) {
	if c.Callbacks == nil {
		c.Callbacks = make(map[string][]interface{})
	}
	if _, ok := c.Callbacks[t]; !ok {
		c.Callbacks[t] = make([]interface{}, 0)
	}
	c.Callbacks[t] = append(c.Callbacks[t], cb)
}

func (c *Callbacker) GetCbs(t string) []interface{} {
	if c.Callbacks == nil {
		return []interface{}{}
	}
	if _, ok := c.Callbacks[t]; !ok {
		return []interface{}{}
	}
	return c.Callbacks[t]
}

func (c *Callbacker) RunCbs(t string, v interface{}) {
	if c.Callbacks == nil {
		return
	}
	if _, ok := c.Callbacks[t]; !ok {
		return
	}
	vv := reflect.ValueOf(v)
	for _, cb := range c.Callbacks[t] {
		reflect.ValueOf(cb).Call([]reflect.Value{vv})
	}
}
