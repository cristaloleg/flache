package flache

import (
	"hash/fnv"
	"sync"
	"testing"
)

func BenchmarkHash(b *testing.B) {
	b.ReportAllocs()

	testCases := randStrings(b.N)

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		_ = fnvHash(testCases[i])
	}
	b.StartTimer()
}

func BenchmarkStdFnvHash(b *testing.B) {
	b.ReportAllocs()

	testCases := randStrings(b.N)

	stdFnvHash := fnv.New64()

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		stdFnvHash.Reset()
		stdFnvHash.Write([]byte(testCases[i]))
		_ = stdFnvHash.Sum64()
	}
	b.StartTimer()
}

func BenchmarkCache10K(b *testing.B) {
	b.ReportAllocs()

	cache := NewStaticCache()
	if cache == nil {
		b.Error("Should be instantiated")
	}
	var wg sync.WaitGroup
	size := 10000
	writeThreads := 100
	readThreads := 1000
	keys := randStrings(size)

	for i := 0; i < writeThreads; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, c *Cache) {
			for i := 0; i < size; i++ {
				c.Add(keys[i], i)
			}
			wg.Done()
		}(&wg, cache)
	}

	for i := 0; i < readThreads; i++ {

		wg.Add(1)
		go func(lwg *sync.WaitGroup, c *Cache) {
			for i := 0; i < size; i++ {
				_ = c.Get(keys[i])
			}
			wg.Done()
		}(&wg, cache)
	}

	wg.Wait()
}

func BenchmarkGetSameConcurrently(b *testing.B) {
	b.ReportAllocs()

	cache := NewStaticCache()
	if cache == nil {
		b.Error("Should be instantiated")
	}

	size := 10000
	threads := 100
	keys := randStrings(size)
	done := make(chan struct{})

	for i := 0; i < size; i++ {
		cache.Add(keys[i], i)
	}

	for i := 0; i < threads; i++ {
		go func() {
			for j := 0; j < size; j++ {
				_ = cache.Get(keys[j])
			}
			done <- struct{}{}
		}()
	}

	for i := 0; i < threads; i++ {
		<-done
	}
}

func BenchmarkSetSameConcurrently(b *testing.B) {
	b.ReportAllocs()

	cache := NewStaticCache()
	if cache == nil {
		b.Error("Should be instantiated")
	}

	size := 10000
	threads := 1000
	keys := randStrings(size)
	done := make(chan struct{})

	for i := 0; i < size; i++ {
		cache.Add(keys[i], i)
	}

	for i := 0; i < threads; i++ {
		go func(key string) {
			for j := 0; j < size; j++ {
				cache.Set(key, j)
			}
			done <- struct{}{}
		}(keys[i])
	}

	for i := 0; i < threads; i++ {
		<-done
	}
}
