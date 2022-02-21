package tyme

import (
	"sync"
	"time"
)

var (
	sleepFunc = time.Sleep
	sfLock    sync.Mutex
)

func Sleep(d time.Duration) time.Duration {
	sleepFunc(d)
	return d
}

func StubSleep(sleep func(time.Duration), f func()) {
	sfLock.Lock()
	defer func() {
		sleepFunc = time.Sleep
		sfLock.Unlock()
	}()

	sleepFunc = sleep
	f()
}
