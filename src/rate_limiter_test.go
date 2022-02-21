package ratelimit

import (
	"testing"
	"time"

	"github.com/gabehardgrave/ratelimit/src/internal/tyme"
	"github.com/stretchr/testify/assert"
)

func TestRLZeroValue(t *testing.T) {
	limiter := RateLimiter{}

	slept := false
	sleep := func(d time.Duration) {
		slept = true
		assert.LessOrEqual(t, d, time.Duration(0))
	}

	tyme.StubSleep(sleep, func() {
		limiter.SleepUntilReady()
	})

	assert.True(t, slept)
}

func TestRLRetryAfterDuration(t *testing.T) {
	limiter := RateLimiter{}

	slept := false
	sleep := func(d time.Duration) {
		slept = true
		assert.Equal(t, 1*time.Hour, d)
	}

	tyme.FreezeTimeAt(time.Now(), func() {
		limiter.SetRetryAfterDuration(1 * time.Hour)

		tyme.StubSleep(sleep, func() {
			limiter.SleepUntilReady()
		})
	})

	assert.True(t, slept)
}

func TestRLRetryAfterTime(t *testing.T) {
	limiter := RateLimiter{}
	now := time.Now()

	slept := false
	sleep := func(d time.Duration) {
		slept = true
		assert.Equal(t, 30*time.Minute, d)
	}

	tyme.FreezeTimeAt(now, func() {
		limiter.SetRetryAfterTime(now.Add(30 * time.Minute))

		tyme.StubSleep(sleep, func() {
			limiter.SleepUntilReady()
		})
	})

	assert.True(t, slept)
}
