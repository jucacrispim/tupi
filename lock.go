// Copyright 2020 Juca Crispim <juca@poraodojuca.net>

// This file is part of tupi.

// tupi is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// tupi is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with tupi. If not, see <http://www.gnu.org/licenses/>.

package main

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
		km.c.Wait()
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
