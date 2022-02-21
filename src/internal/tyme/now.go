package tyme

import (
	"sync"
	"time"
)

var (
	timeNowFunc = time.Now
	tnfLock     sync.Mutex
)

func Now() time.Time {
	return timeNowFunc()
}

func Until(t time.Time) time.Duration {
	return t.Sub(Now())
}

func FreezeTimeAt(t time.Time, f func()) {
	tnfLock.Lock()
	defer func() {
		timeNowFunc = time.Now
		tnfLock.Unlock()
	}()

	timeNowFunc = func() time.Time { return t }
	f()
}
