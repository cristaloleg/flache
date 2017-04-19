package flache

import (
	"hash/fnv"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	cache := NewCache(10*time.Second, 100*time.Second)
	if cache == nil {
		t.Error("`cache` should be instantiated")
	}
	check := func(Cacher) {}
	check(cache)

	if cache.Size() != 0 {
		t.Error("Should be empty")
	}

	cache.AddExt("key1", "value1", time.Duration(10)*time.Second)

	if !cache.Has("key1") {
		t.Error("Should have `key1`")
	}

	value, left, ok := cache.GetExt("key1")
	if !ok || value != "value1" || left == 0 {
		t.Error("Should have `key1` with proper value")
	}
}

func TestNewAddHasPurge(t *testing.T) {
	cache := NewCache(time.Duration(100)*time.Millisecond, time.Duration(50)*time.Second)
	if cache == nil {
		t.Error("`cache` should be instantiated")
	}

	cache.Add("key1", "value1")
	if !cache.Has("key1") {
		t.Error("Should have `key1`")
	}
	if cache.Has("key2") {
		t.Error("Should not have `key2`")
	}

	if size := cache.Size(); size != 1 {
		t.Error("Should have only 1 entry, but has", size)
	}

	<-time.After(125 * time.Millisecond)

	if cache.Has("key1") {
		t.Error("Should not have `key1` anymore")
	}
	if cache.Has("key2") {
		t.Error("Should not have `key2` as before")
	}
	if size := cache.Size(); size != 0 {
		t.Error("Should have only 0 entries, but has ", size)
	}

	cache.Purge()

	if cache.Size() != 0 {
		t.Error("Should be empty")
	}
}

func TestGetSet(t *testing.T) {
	cache := NewCache(time.Duration(100)*time.Millisecond, time.Duration(100)*time.Millisecond)

	cache.AddExt("key2", 123, time.Duration(10)*time.Millisecond)
	value, left, ok := cache.GetExt("key2")
	if !ok || value != 123 || left == 0 {
		t.Error("Should have `key2` value")
	}

	cache.Set("key2", "123")

	value, left, ok = cache.GetExt("key2")
	if !ok || value != "123" {
		t.Error("Should have `key2` value")
	}

	<-time.After(10 * time.Millisecond)

	value, left, ok = cache.GetExt("key2")
	if ok || value != nil || left != 0 {
		t.Error("Should not have `key2` anymore")
	}
}

func TestDefault(t *testing.T) {
	cache := NewCache(time.Duration(10)*time.Millisecond, time.Duration(100)*time.Millisecond)
	cache.AddExt("key1", 123, time.Duration(5)*time.Millisecond)
	if !cache.Has("key1") {
		t.Error("Should have `key1`")
	}

	<-time.After(time.Duration(cache.expiration))

	if cache.Has("key1") {
		t.Error("Should not have `key1` anymore")
	}

	cache.AddExt("key2", "123", time.Duration(1))
	cache.Clear()

	if cache.Size() != 0 {
		t.Error("Should be empty")
	}
}

func TestConcurrent(t *testing.T) {
	size := 10000
	threads := 5
	segment := size / threads
	keys := randStrings(size)
	done := make(chan struct{})
	cache := NewStaticCache()

	for i := 0; i < threads; i++ {
		go func(offset int) {
			for j := offset * segment; j < (segment)*(offset+1); j++ {
				cache.Add(keys[j], j)

				val := cache.Get(keys[j])
				if val == nil || val.(int) != j {
					t.Errorf("For %s expected %d but was %d", keys[j], val.(int), j)
				}
			}
			done <- struct{}{}
		}(i)
	}

	for i := 0; i < threads; i++ {
		<-done
	}

	if cache.Size() != int64(size) {
		t.Error("wut? ", cache.Size())
	}
}

func TestHash(t *testing.T) {
	testCases := randStrings(1000)
	stdFnvHash := fnv.New64()

	for _, test := range testCases {
		stdFnvHash.Reset()
		stdFnvHash.Write([]byte(test))

		expected := stdFnvHash.Sum64()
		actual := fnvHash(test)
		if actual != expected {
			t.Errorf("%s expected %d but was %d", test, expected, actual)
		}
	}
}

func readFromCache(n int64, wg *sync.WaitGroup, c *Cache) {
	wg.Add(1)
	for i := int64(0); i < n; i++ {
		c.Get("")
	}
	wg.Done()
}

func writeToCache(n int64, wg *sync.WaitGroup, c *Cache) {
	wg.Add(1)
	for i := int64(0); i < n; i++ {
		c.Add("", "")
	}
	wg.Done()
}

func randStrings(size int) []string {
	res := make([]string, size)

	for i := 0; i < size; i++ {
		res[i] = strconv.Itoa((100000 + i) * 1000001)
	}
	return res
}
