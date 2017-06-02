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

func (c *cache) Add(key string, value interface{}) {
	c.Lock()
	//
	c.entries[key] = entry{
		Value: value,
	}
	//
	c.Unlock()
}

func (c *cache) Has(key string) bool {
	c.Lock()
	//
	_, ok := c.entries[key]
	//
	c.Unlock()

	return ok
}

func (c *cache) Del(key string) {
	c.Lock()
	//
	delete(c.entries, key)
	//
	c.Unlock()
}

func (c *cache) Keys() []string {
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

func (c *cache) Clear() {
	c.Lock()
	//
	c.entries = make(map[string]entry)
	//
	c.Unlock()
}

func (c *cache) Size() int {
	c.Lock()
	//
	size := len(c.entries)
	//
	c.Unlock()

	return size
}
