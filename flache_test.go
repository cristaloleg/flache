package flache

import (
	"testing"
	"time"
)

var value interface{}
var left time.Duration
var ok bool

func TestNewCache(t *testing.T) {
	Cache := NewCache(10*time.Second, 100*time.Second)
	if Cache == nil {
		t.Error("`Cache` should be instantiated")
	}

	Cache.AddExt("key1", "value1", time.Duration(10)*time.Second)

	if !Cache.Has("key1") {
		t.Error("Should have `key1`")
	}

	value, left, ok = Cache.Get("key1")
	if !ok || value != "value1" || left == 0 {
		t.Error("Should have `key1` with proper value")
	}
}

func TestNewCache_AddHasPurge(t *testing.T) {
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

func TestCache_GetSet(t *testing.T) {
	cache := NewCache(time.Duration(100)*time.Millisecond, time.Duration(100)*time.Millisecond)

	cache.AddExt("key2", 123, time.Duration(10)*time.Millisecond)
	value, left, ok = cache.Get("key2")
	if !ok || value != 123 || left == 0 {
		t.Error("Should have `key2` value")
	}

	cache.Set("key2", "123")

	value, left, ok = cache.Get("key2")
	if !ok || value != "123" {
		t.Error("Should have `key2` value")
	}

	<-time.After(10 * time.Millisecond)

	value, left, ok = cache.Get("key2")
	if ok || value != nil || left != 0 {
		t.Error("Should not have `key2` anymore")
	}
}

func TestCache_Default(t *testing.T) {
	cache := NewCache(time.Duration(10)*time.Millisecond, time.Duration(100)*time.Millisecond)
	cache.AddExt("key1", 123, time.Duration(5)*time.Millisecond)
	if !cache.Has("key1") {
		t.Error("Should have `key1`")
	}
	//value, life, ok := cache.GetWithTime()
	//if !ok || value != 123 || time == 0 {
	//	t.Error("Should ")
	//}

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

func TestCache_Concurrent(t *testing.T) {
	cache := NewCache(time.Duration(10)*time.Millisecond, time.Duration(100)*time.Millisecond)

	cache.AddExt("key1", 123, time.Duration(1))

	go func() {
		cache.AddExt("key", 321, time.Duration(1))
	}()
}