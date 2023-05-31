package tupi

import "sync"

type keyedMutex struct {
	c *sync.Cond
	l sync.Locker
	m map[string]int
}

func newKeyedMutex() *keyedMutex {
	l := sync.Mutex{}
	return &keyedMutex{c: sync.NewCond(&l), l: &l, m: make(map[string]int)}
}

func (km *keyedMutex) isLocked(key string) bool {
	km.l.Lock()
	defer km.l.Unlock()
	_, ok := km.m[key]
	return ok
}

func (km *keyedMutex) Lock(key string) {
	for km.isLocked(key) {
		km.c.L.Lock()
		// notest
		km.c.Wait()
		km.c.L.Unlock()
	}
	km.l.Lock()
	defer km.l.Unlock()
	km.m[key] = 1
}

func (km *keyedMutex) Unlock(key string) {
	km.l.Lock()
	defer km.l.Unlock()
	delete(km.m, key)
	km.c.Broadcast()
}

var kmu *keyedMutex = newKeyedMutex()

// AcquireLock locks a resource based in a key When you are done you must
// release the lock with ReleaseLock()
func AcquireLock(key string) {
	kmu.Lock(key)
}

// ReleaseLock releases the lock for a given resource identified by a key.
func ReleaseLock(key string) {
	kmu.Unlock(key)
}

// IsLocked return a bool informing if a given resource is locked
func IsLocked(key string) bool {
	return kmu.isLocked(key)
}
