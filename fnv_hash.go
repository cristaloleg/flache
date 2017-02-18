package flache

const (
	offset64 = 14695981039346656037
	prime64  = 1099511628211
)

// fnvHash return hash of the key, FNV-1 algorithm
func fnvHash(key string) uint64 {
	hash := uint64(offset64)
	for _, c := range key {
		hash *= prime64
		hash ^= uint64(c)
	}
	return hash
}
