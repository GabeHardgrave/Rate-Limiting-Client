package ratelimit

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gabehardgrave/ratelimit/src/internal/testutils"
	"github.com/gabehardgrave/ratelimit/src/internal/tyme"
	"github.com/stretchr/testify/assert"
)

func TestExponentialBackoff(t *testing.T) {
	retry, after := ExponentialBackoff(testutils.StubResponse(200, ""))
	assert.False(t, retry)
	assert.Zero(t, after)

	now := time.Now()

	tyme.FreezeTimeAt(now, func() {
		retry, after = ExponentialBackoff(testutils.StubResponse(429, ""))
		assert.True(t, retry)
		assert.EqualValues(t, now.Add(1*time.Second), after)
	})

	tyme.FreezeTimeAt(now, func() {
		retry, after = ExponentialBackoff(
			testutils.StubResponse(429, ""), // 1s delay
			testutils.StubResponse(429, ""), // 2s delay
			testutils.StubResponse(429, ""), // 4s delay
		)
		assert.True(t, retry)
		assert.EqualValues(t, now.Add(4*time.Second), after)
	})

	tyme.FreezeTimeAt(now, func() {
		prevResps := make([]*http.Response, 22, 23) // 2^22 is just above DefaultMaxRetryAfterDuration
		retry, after = ExponentialBackoff(
			testutils.StubResponse(429, ""),
			prevResps...,
		)
		assert.False(t, retry)
		assert.True(t, after.After(now.Add(DefaultMaxRetryAfterDuration)))
	})
}

func TestRetryAfterDurationInHeader(t *testing.T) {
	retry, after := RetryAfterDurationInHeader(testutils.StubResponse(200, ""))
	assert.False(t, retry)
	assert.Zero(t, after)

	now := time.Now()

	tyme.FreezeTimeAt(now, func() {
		retry, after = RetryAfterDurationInHeader(testutils.StubResponse(429, ""))
		assert.True(t, retry)
		assert.Zero(t, after)
	})

	tyme.FreezeTimeAt(now, func() {
		retry, after = RetryAfterDurationInHeader(testutils.StubResponse(503, "",
			"Retry-After", "10"))
		assert.True(t, retry)
		assert.EqualValues(t, now.Add(10*time.Second), after)
	})

	s := strconv.FormatUint(uint64(DefaultMaxRetryAfterDuration/time.Second)+1, 10)
	tyme.FreezeTimeAt(now, func() {
		retry, after = RetryAfterDurationInHeader(testutils.StubResponse(500, "",
			"Retry-After", s))
		assert.False(t, retry)
		assert.EqualValues(t, now.Add(DefaultMaxRetryAfterDuration+(1*time.Second)), after)
	})
}

func TestRetryAfterTimeInHeader(t *testing.T) {
	retry, after := RetryAfterDurationInHeader(testutils.StubResponse(200, ""))
	assert.False(t, retry)
	assert.Zero(t, after)

	resp := testutils.StubResponse(500, "",
		"Retry-After", "Wed, 21 Oct 2015 07:28:00 GMT")
	retry, after = RetryAfterTimeInHeader(resp)
	assert.True(t, retry)
	assert.Equal(t, after, time.Date(2015, time.October, 21, 7, 28, 0, 0, time.UTC))
}
