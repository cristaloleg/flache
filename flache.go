package flache

import (
	"sync"
	"time"
)

const (
	shardsNum  = uint64(1 << 10)
	shardsMask = shardsNum - 1
)

// entry internal struct to store value in the `cache`
type entry struct {
	Expiration int64
	Value      interface{}
}

// Flache a concurrent, sharded and flexible cache
type Flache struct {
	buckets     [shardsNum]map[string]entry
	mutexes     [shardsNum]sync.Mutex
	expiration  int64
	autocleanup int64
}

// NewFlache creates empty Flache
func NewFlache(expiration, autocleanup time.Duration) *Flache {
	f := NewStaticFlache()
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

// NewStaticFlache creates empty Flache without any time setups
func NewStaticFlache() *Flache {
	f := Flache{}
	for i := uint64(0); i < shardsNum; i++ {
		f.buckets[i] = make(map[string]entry)
		f.mutexes[i] = sync.Mutex{}
	}
	return &f
}

// Add adds key-value in to the cache for a predefined time
func (f *Flache) Add(key string, value interface{}) {
	f.AddExt(key, value, time.Duration(f.expiration))
}

// AddExt adds key-value in to the cache for a specified time
func (f *Flache) AddExt(key string, value interface{}, duration time.Duration) {
	now := time.Now().UnixNano()
	index := f.hash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	f.buckets[index][key] = entry{Expiration: int64(duration) + now, Value: value}
	//
	f.mutexes[index].Unlock()
}

// Check returns state of an object for a given key
func (f *Flache) Check(key string) (exists, alive bool) {
	index := f.hash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	value, exists := f.buckets[index][key]
	//
	f.mutexes[index].Unlock()

	alive = exists && (f.expiration == 0 || value.Expiration > time.Now().UnixNano())
	return exists, alive
}

// Has returns true if value for key is in the cache and not expired
func (f *Flache) Has(key string) bool {
	index := f.hash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	value, ok := f.buckets[index][key]
	//
	f.mutexes[index].Unlock()

	return ok && (f.expiration == 0 || value.Expiration > time.Now().UnixNano())
}

// HasExt returns true(and duration) if value for key is in the cache and not expired
func (f *Flache) HasExt(key string) (time.Duration, bool) {
	index := f.hash(key) & shardsMask

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

// HasStrict returns true if object is in the cache, even expired
func (f *Flache) HasStrict(key string) bool {
	index := f.hash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	_, ok := f.buckets[index][key]
	//
	f.mutexes[index].Unlock()

	return ok
}

// Get returns value for a `key` if it's not expired yet
func (f *Flache) Get(key string) (interface{}, time.Duration, bool) {
	index := f.hash(key) & shardsMask

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
func (f *Flache) GetUpd(key string) (interface{}, time.Duration, bool) {
	index := f.hash(key) & shardsMask

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

// Del removes `key` from the cache
func (f *Flache) Del(key string) {
	index := f.hash(key) & shardsMask

	f.mutexes[index].Lock()
	//
	delete(f.buckets[index], key)
	//
	f.mutexes[index].Unlock()
}

// Set updates value for a `key` but do not touches expiration time
func (f *Flache) Set(key string, value interface{}) {
	expiration := f.expiration
	index := f.hash(key) & shardsMask

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
func (f *Flache) Touch(key string) {
	if f.expiration == 0 {
		return
	}
	index := f.hash(key) & shardsMask

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

// Keys returns slice of all non-expired keys
func (f *Flache) Keys() []string {
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
func (f *Flache) Values() []interface{} {
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
func (f *Flache) Purge() {
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
func (f *Flache) Clear() {
	for i := uint64(0); i < shardsNum; i++ {
		f.mutexes[i].Lock()
		//
		f.buckets[i] = make(map[string]entry)
		//
		f.mutexes[i].Unlock()
	}
}

// Size returns amount of not expired objects
func (f *Flache) Size() int64 {
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

// hash return hash of key, FNV-1 algorithm
func (f *Flache) hash(key string) uint64 {
	hash := uint64(14695981039346656037)
	for _, c := range key {
		hash *= 1099511628211
		hash ^= uint64(c)
	}
	return hash
}
