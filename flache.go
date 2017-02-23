package flache

import (
	"sync"
	"time"
)

const (
	shardsLog  = uint64(10)
	shardsNum  = 1 << shardsLog
	shardsMask = shardsNum - 1

	// for fnv hash function
	offset64 = 14695981039346656037
	prime64  = 1099511628211
)

// entry internal struct to store value in the `cache`
type entry struct {
	Expiration int64
	Value      interface{}
}

// Cacher cache interface
type Cacher interface {
	Add(string, interface{})
	AddExt(string, interface{}, time.Duration)
	Check(string) (bool, bool)
	Has(string) bool
	HasExt(string) (time.Duration, bool)
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

// Cache a concurrent, sharded and flexible cache
type Cache struct {
	buckets     [shardsNum]map[string]entry
	mutexes     [shardsNum]sync.Mutex
	expiration  int64
	autocleanup int64
}

// NewCache creates empty Cache
func NewCache(expiration, autocleanup time.Duration) *Cache {
	f := NewStaticCache()
	f.expiration = int64(expiration)
	f.autocleanup = int64(autocleanup)

	if autocleanup != 0 {
		go func() {
			time.Sleep(time.Duration(f.autocleanup))
			f.Purge()
		}()
	}
	return f
}

// NewStaticCache creates empty Cache without any time setups
func NewStaticCache() *Cache {
	f := Cache{}
	for i := uint64(0); i < shardsNum; i++ {
		f.buckets[i] = make(map[string]entry)
		f.mutexes[i] = sync.Mutex{}
	}
	return &f
}

// Add adds key-value in to the cache for a predefined time
func (f *Cache) Add(key string, value interface{}) {
	f.AddExt(key, value, time.Duration(f.expiration))
}

// AddExt adds key-value in to the cache for a specified time
func (f *Cache) AddExt(key string, value interface{}, duration time.Duration) {
	now := time.Now().UnixNano()
	index := fnvHash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	f.buckets[index][key] = entry{Expiration: int64(duration) + now, Value: value}
	//
	f.mutexes[index].Unlock()
}

// Check returns state of an object for a given key
func (f *Cache) Check(key string) (exists, alive bool) {
	index := fnvHash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	value, exists := f.buckets[index][key]
	//
	f.mutexes[index].Unlock()

	alive = exists && (f.expiration == 0 || value.Expiration > time.Now().UnixNano())
	return exists, alive
}

// Has returns true if value for key is in the cache and not expired
func (f *Cache) Has(key string) bool {
	index := fnvHash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	value, ok := f.buckets[index][key]
	//
	f.mutexes[index].Unlock()

	return ok && (f.expiration == 0 || value.Expiration > time.Now().UnixNano())
}

// HasExt returns true(and duration) if value for key is in the cache and not expired
func (f *Cache) HasExt(key string) (time.Duration, bool) {
	index := fnvHash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	value, ok := f.buckets[index][key]
	//
	f.mutexes[index].Unlock()

	if ok && (f.expiration == 0 || value.Expiration > time.Now().UnixNano()) {
		return time.Duration(value.Expiration), true
	}
	return 0, false
}

// Get returns value for a `key` if it's not expired yet
func (f *Cache) Get(key string) interface{} {
	index := fnvHash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	value, ok := f.buckets[index][key]
	//
	f.mutexes[index].Unlock()

	if ok && (f.expiration == 0 || value.Expiration > time.Now().UnixNano()) {
		return value.Value
	}
	return nil
}

// GetExt returrns value, time and bool for a `key` if it's not expired yet
func (f *Cache) GetExt(key string) (interface{}, time.Duration, bool) {
	index := fnvHash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	value, ok := f.buckets[index][key]
	//
	f.mutexes[index].Unlock()

	if ok && (f.expiration == 0 || value.Expiration > time.Now().UnixNano()) {
		return value.Value, time.Duration(value.Expiration), true
	}
	var zero interface{}
	return zero, 0, false
}

// GetUpd returns value for a `key` if it's not expired yet and updates expiration time for the `key`
func (f *Cache) GetUpd(key string) (interface{}, time.Duration, bool) {
	index := fnvHash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	value, ok := f.buckets[index][key]
	if ok {
		value.Expiration = f.expiration + time.Now().UnixNano()
		f.buckets[index][key] = value
	}
	//
	f.mutexes[index].Unlock()

	if ok {
		return value.Value, time.Duration(value.Expiration), true
	}
	var zero interface{}
	return zero, 0, false
}

// Gets returns values for a `keys` if it's not expired yet
func (f *Cache) Gets(keys ...string) []interface{} {
	size := len(keys)
	res := make([]interface{}, size)
	now := time.Now().UnixNano()

	for i := 0; i < size; i++ {
		index := fnvHash(keys[i]) & shardsMask

		f.mutexes[index].Lock()
		//
		value, ok := f.buckets[index][keys[i]]
		//
		f.mutexes[index].Unlock()

		if ok && value.Expiration >= now {
			res[i] = value.Value
		} else {
			res[i] = nil
		}
	}

	return res
}

// Del removes `key` from the cache
func (f *Cache) Del(key string) {
	index := fnvHash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	delete(f.buckets[index], key)
	//
	f.mutexes[index].Unlock()
}

// Dels removes `keys` from the cache
func (f *Cache) Dels(keys ...string) {
	size := len(keys)

	for i := 0; i < size; i++ {
		index := fnvHash(keys[i]) & shardsMask

		f.mutexes[index].Lock()
		//
		delete(f.buckets[index], keys[i])
		//
		f.mutexes[index].Unlock()
	}
}

// Set updates value for a `key` but do not touches expiration time
func (f *Cache) Set(key string, value interface{}) {
	expiration := f.expiration
	index := fnvHash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	oldValue, ok := f.buckets[index][key]
	if ok {
		expiration = oldValue.Expiration
	}
	f.buckets[index][key] = entry{Expiration: expiration, Value: value}
	//
	f.mutexes[index].Unlock()
}

// Touch updates expiration time for the `key`
func (f *Cache) Touch(key string) {
	if f.expiration == 0 {
		return
	}
	index := fnvHash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	value, ok := f.buckets[index][key]
	if ok {
		value.Expiration = f.expiration + time.Now().UnixNano()
		f.buckets[index][key] = value
	}
	//
	f.mutexes[index].Unlock()
}

// Touchs updates expiration time for the `key`
func (f *Cache) Touchs(keys ...string) {
	if f.expiration == 0 {
		return
	}
	size := len(keys)
	now := time.Now().UnixNano()

	for i := 0; i < size; i++ {
		index := fnvHash(keys[i]) & shardsMask

		f.mutexes[index].Lock()
		//
		value, ok := f.buckets[index][keys[i]]
		if ok {
			value.Expiration = f.expiration + now
			f.buckets[index][keys[i]] = value
		}
		//
		f.mutexes[index].Unlock()
	}
}

// Keys returns slice of all non-expired keys
func (f *Cache) Keys() []string {
	var res []string
	for i := uint64(0); i < shardsNum; i++ {
		now := time.Now().UnixNano()

		f.mutexes[i].Lock()
		//
		for k, v := range f.buckets[i] {
			if f.expiration == 0 || v.Expiration > now {
				res = append(res, k)
			}
		}
		//
		f.mutexes[i].Unlock()
	}
	return res
}

// Values returns slice of all non-expired values
func (f *Cache) Values() []interface{} {
	var res []interface{}
	for i := uint64(0); i < shardsNum; i++ {
		now := time.Now().UnixNano()

		f.mutexes[i].Lock()
		//
		for _, v := range f.buckets[i] {
			if f.expiration == 0 || v.Expiration > now {
				res = append(res, v.Value)
			}
		}
		//
		f.mutexes[i].Unlock()
	}
	return res
}

// Purge removes expired entries from shard
func (f *Cache) Purge() {
	if f.expiration == 0 {
		return
	}
	for i := uint64(0); i < shardsNum; i++ {
		go func(i uint64) {
			now := time.Now().UnixNano()

			f.mutexes[i].Lock()
			//
			for k, v := range f.buckets[i] {
				if v.Expiration < now {
					delete(f.buckets[i], k)
				}
			}
			//
			f.mutexes[i].Unlock()
		}(i)
	}
}

// Clear removes all values from a cache
func (f *Cache) Clear() {
	for i := uint64(0); i < shardsNum; i++ {
		f.mutexes[i].Lock()
		//
		f.buckets[i] = make(map[string]entry)
		//
		f.mutexes[i].Unlock()
	}
}

// Size returns amount of not expired objects
func (f *Cache) Size() int64 {
	var res int64
	ch := make(chan int64, shardsNum)

	for i := uint64(0); i < shardsNum; i++ {
		go func(i uint64, c chan int64) {
			var alive int64

			f.mutexes[i].Lock()
			//
			if f.expiration == 0 {
				alive = int64(len(f.buckets[i]))
			} else {
				now := time.Now().UnixNano()
				for _, v := range f.buckets[i] {
					if v.Expiration > now {
						alive++
					}
				}
			}
			//
			f.mutexes[i].Unlock()

			ch <- alive
		}(i, ch)
	}
	for i := uint64(0); i < shardsNum; i++ {
		res += <-ch
	}
	return res
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
