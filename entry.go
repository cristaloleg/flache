package flache

import (
	"time"
)

// entry internal struct to store value and it's lifetime
type entry struct {
	Expiration int64
	Value      interface{}
}

func (e *entry) IsAlive() bool {
	return e.Expiration == 0 || e.Expiration > time.Now().UnixNano()
}
