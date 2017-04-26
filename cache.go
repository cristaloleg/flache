package flache

import "sync"

type cache struct {
	sync.Mutex
	entries map[string]entry
}

func newCache() *cache {
	c := &cache{
		entries: make(map[string]entry),
	}
	return c
}

func (c *cache) add(key string, value interface{}) {
	c.Lock()
	//
	c.entries[key] = entry{
		Value: value,
	}
	//
	c.Unlock()
}

func (c *cache) has(key string) bool {
	c.Lock()
	//
	_, ok := c.entries[key]
	//
	c.Unlock()
	return ok
}

func (c *cache) del(key string) {
	c.Lock()
	//
	delete(c.entries, key)
	//
	c.Unlock()
}

func (c *cache) keys() []string {
	c.Lock()
	//
	keys := make([]string, len(c.entries))
	i := 0
	for k := range c.entries {
		keys[i] = k
		i++
	}
	//
	c.Unlock()
	return keys
}

func (c *cache) clear() {
	c.Lock()
	//
	c.entries = make(map[string]entry)
	//
	c.Unlock()
}

func (c *cache) size() int {
	c.Lock()
	//
	size := len(c.entries)
	//
	c.Unlock()
	return size
}
