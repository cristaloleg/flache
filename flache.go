package flache

import "time"

const (
	shardsLog  = uint64(10)
	shardsNum  = 1 << shardsLog
	shardsMask = shardsNum - 1

	// for fnv hash function
	offset64 = 14695981039346656037
	prime64  = 1099511628211
)

// Cacher cache interface
type Cacher interface {
	Add(string, interface{})
	AddExt(string, interface{}, time.Duration)
	Check(string) (bool, bool)
	Has(string) bool
	HasExt(string) (time.Duration, bool)
	HasEntry(string, interface{}) bool
	Get(string) interface{}
	GetExt(string) (interface{}, time.Duration, bool)
	Gets(...string) []interface{}
	Del(string)
	Dels(...string)
	Set(string, interface{})
	Touch(string)
	Touchs(...string)
	Keys() []string
	Values() []interface{}
	Purge()
	Clear()
	Size() int64
}

// Flache fast concurrent cache
type Flache struct {
	shards [shardsNum]cache
}

// New returns a pointer to the new instance of Flache
func New() *Flache {
	f := &Flache{}
	for i := 0; i < shardsNum; i++ {
		f.shards[i] = *newCache()
	}
	return f
}

func getIndex(key string) uint64 {
	return fnvHash(key) & shardsMask
}

// Add adds key-value to the cache
func (f *Flache) Add(key string, value interface{}) {
	index := getIndex(key)
	f.shards[index].add(key, value)
}

// Has returns true if key is present, false otherwise
func (f *Flache) Has(key string) bool {
	index := getIndex(key)
	return f.shards[index].has(key)
}

// Del deletes key-value from the cache
func (f *Flache) Del(key string) {
	index := getIndex(key)
	f.shards[index].del(key)
}

// Keys returns all keys inside the cache
func (f *Flache) Keys() []string {
	var keys []string
	for i := 0; i < shardsNum; i++ {
		keys = append(keys, f.shards[i].keys()...)
	}
	return keys
}

//Size returns number of key-value pairs in the cache
func (f *Flache) Size() int {
	size := 0
	for i := 0; i < shardsNum; i++ {
		size += f.shards[i].size()
	}
	return size
}

// fnvHash return hash of the key, FNV-1 algorithm
func fnvHash(key string) uint64 {
	hash := uint64(offset64)
	for _, c := range key {
		hash *= prime64
		hash ^= uint64(c)
	}
	return hash
}
