package flache

import (
	"testing"
	"hash/fnv"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func TestCache_Hash(t *testing.T) {
	testCases := randStrings(100)
	stdFnvHash := fnv.New64()

	for _, test := range testCases {
		stdFnvHash.Reset()
		stdFnvHash.Write([]byte(test))

		expected := stdFnvHash.Sum64()
		actual := fnvHash(test)
		if actual != expected {
			t.Errorf("%s expected %s but was %s", test, expected, actual)
		}
	}
}

func BenchmarkCache_Hash(b *testing.B) {
	b.ReportAllocs()
	testCases := randStrings(b.N)

	b.StopTimer()
	for i := 0; i < b.N; i++ {
		_ = fnvHash(testCases[i])
	}
	b.StartTimer()
}

func BenchmarkCache_StdFnvHash(b *testing.B) {
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

func randStrings(size int) []string {
	res := make([]string, size)

	for i := 0; i < size; i++ {
		res[i] = randStr(15+(i%10-5), i+1)
	}
	return res
}

func randStr(size, seed int) string {
	res := make([]byte, size)
	sz := len(letters)
	for i := 0; i < size; i++ {
		res[i] = letters[seed*i%sz]
	}
	return string(res)
}
