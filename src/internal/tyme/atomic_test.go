package tyme

import (
	"sync"
	"testing"
	"time"

	"github.com/gabehardgrave/ratelimit/src/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestUpdateIfLater(t *testing.T) {
	atomicT := Atomic{}
	later := time.Now()

	atomicT.UpdateIfLater(later)
	assert.Equal(t, atomicT.Time(), later)
}

// Be sure to run with `--race`
func TestUpdateIfLaterThreadSafety(t *testing.T) {

	sleepRandomly := testutils.RandomSleeper(500)

	durations := []time.Duration{
		1, 10, 100, 1000, 1_000_000,
	}
	now := time.Now()
	expectedTime := now.Add(1_000_000)

	wg := sync.WaitGroup{}

	atomicT := Atomic{}

	for _, d := range durations {
		wg.Add(1)
		go func(d time.Duration) {
			sleepRandomly()
			atomicT.UpdateIfLater(now.Add(d))
			wg.Done()
		}(d)
	}
	wg.Wait()

	assert.Equal(t, atomicT.Time(), expectedTime)
}
