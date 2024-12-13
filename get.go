package memory

import (
	"github.com/coocood/freecache"
	"github.com/golang/protobuf/proto"

	mm "github.com/shiyunjin/MeMCache/model"
)

func (c *Cache) Get(key string, value proto.Message, caller mm.CacheCaller, expireSeconds int) error {
	bKey := []byte(key)
	v, err := c.storage.Get(bKey)
	switch err {
	case freecache.ErrNotFound:
		return c.do(key, value, caller, expireSeconds)
	case nil:
		return proto.Unmarshal(v, value)
	default:
		return err
	}
}

func (c *Cache) do(key string, value proto.Message, caller mm.CacheCaller, expireSeconds int) error {
	// lock do mutex by start do
	c.doMu.Lock()

	// challenge caller group
	if calldo, ok := c.doMap[key]; ok {
		// unlock do mutex by wait result
		c.doMu.Unlock()

		calldo.wg.Wait()
		proto.Merge(value, calldo.val)
		return calldo.err
	}

	// add call in do map
	calldo := &call{}
	calldo.wg.Add(1)
	c.doMap[key] = calldo

	// unlock do mutex by run caller
	c.doMu.Unlock()

	// run caller
	calldo.val, calldo.err = caller()
	calldo.wg.Done()

	if calldo.err != nil {
		expireSeconds = 1
	}

	defer func() {
		// lock do mutex by delete do map
		c.doMu.Lock()

		delete(c.doMap, key)

		// unlock do mutex
		c.doMu.Unlock()
	}()

	if err := c.set(key, calldo.val, expireSeconds); err != nil {
		return err
	}

	proto.Merge(value, calldo.val)
	return calldo.err
}

func (c *Cache) set(key string, value proto.Message, expireSeconds int) error {
	bVal, err := proto.Marshal(value)
	if err != nil {
		return err
	}

	return c.storage.Set([]byte(key), bVal, expireSeconds)
}
