// Package rumutex implements an upgradable reader/writer mutual exclusion lock.
package rumutex

import (
	"sync"
	"sync/atomic"
)

// An RUMutex is an upgradable reader/writer mutual exclusion lock. The lock can be held by an
// arbitrary number of readers or a single writer.
type RUMutex struct {
	rw       sync.RWMutex
	upgraded int32
}

// RLock locks ru for shared reading.
func (ru *RUMutex) RLock() {
	ru.rw.RLock()
}

// Upgrade attempts to upgrade a read locked ru to exclusive writing.
// Returns false if the lock is already being upgraded.
func (ru *RUMutex) Upgrade() bool {
	if !atomic.CompareAndSwapInt32(&ru.upgraded, 0, 1) {
		return false
	}
	ru.rw.RUnlock()
	ru.rw.Lock()
	return true
}

// Downgrade downgrades a write locked ru back to shared reading.
func (ru *RUMutex) Downgrade() {
	ru.rw.Unlock()
	ru.rw.RLock()
	atomic.StoreInt32(&ru.upgraded, 0) // don't allow new upgrades until we have the read lock
}

// RUnlock unlocks a read locked ru.
func (ru *RUMutex) RUnlock() {
	ru.rw.RUnlock()
}

// Unlock unlocks a write locked ru.
func (ru *RUMutex) Unlock() {
	ru.upgraded = 0
	ru.rw.Unlock()
}
