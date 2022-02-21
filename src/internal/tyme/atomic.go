package tyme

import (
	"sync"
	"time"
)

type Atomic struct {
	lock sync.RWMutex
	t    time.Time
}

func (at *Atomic) Time() time.Time {
	at.lock.RLock()
	t := at.t
	at.lock.RUnlock()
	return t
}

func (at *Atomic) UpdateIfLater(t time.Time) {
	at.lock.Lock()
	if t.After(at.t) {
		at.t = t
	}
	at.lock.Unlock()
}
