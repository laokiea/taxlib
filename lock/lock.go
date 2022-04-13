package lock

import (
	"sync/atomic"
)

type Lock struct {
	key int32
}

func (l *Lock) Lock() {
	for {
		if atomic.CompareAndSwapInt32(&l.key, 0, 1) {
			return
		}
	}
}

func (l *Lock) Unlock() {
	if atomic.CompareAndSwapInt32(&l.key, 0, 0) {
		panic("unlock of unlocked mutex")
	}
	atomic.CompareAndSwapInt32(&l.key, 1, 0)
	return
}

func (l *Lock) TryLock() bool {
	if atomic.CompareAndSwapInt32(&l.key, 0, 1) {
		return true
	}
	return false
}