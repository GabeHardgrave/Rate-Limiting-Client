package ratelimit

import (
	"net/http"
	"time"

	"github.com/gabehardgrave/ratelimit/src/internal/aychttp"
	"github.com/gabehardgrave/ratelimit/src/internal/tyme"
)

// RateLimiter provides a threadsafe API for self rate limiting applications.
//
// Internally, RateLimiter stores a time, `t`, after which it is safe for the application
// to resume.
//
// `t` can be updated either through `SetRetryAfterTime` or `SetRetryAfterDuration`, although
// neither method will decrease `t`.
//
// In order to block (i.e. honor rate limits), applications must call `SleepUntilReady`. This will
// block the calling goroutine until time `t`.
type RateLimiter struct {
	t tyme.Atomic
}

// SleepUntilReady will block the current goroutine until the rate limit has been honored,
// based on prior calls to SetRetryAfterTime or SetRetryAfterDuration. SleepUntilReady() returns
// the duration it slept for.
func (rl *RateLimiter) SleepUntilReady() (d time.Duration) {
	d = rl.t.Time().Sub(tyme.Now())
	return tyme.Sleep(d)
}

// SetRetryAfterTime updates `t` to max(`t`, `newT`). SetRetryAfterTime does not block the current
// goroutine.
func (rl *RateLimiter) SetRetryAfterTime(newT time.Time) {
	rl.t.UpdateIfLater(newT)
}

// SetRetryAfterDuration updates `t` to max(`t`, `time.Now().Add(d)`). SetRetryAfterDuration does
// not block the current goroutine.
func (rl *RateLimiter) SetRetryAfterDuration(d time.Duration) {
	t := tyme.Now().Add(d)
	rl.SetRetryAfterTime(t)
}

// I'm not sure `do` really belongs here, but I wanted the logic to be reused by `Client` and
// `MultiHostClient`, so this happened.

func (rl *RateLimiter) do(
	req *http.Request,
	client *http.Client,
	policy RetryAfterPolicy,
) (*http.Response, error) {

	var prevResps []*http.Response
	includeBody := aychttp.HasBody(req)

	for {
		rl.SleepUntilReady()

		resp, err := client.Do(req)
		if err != nil {
			return resp, err
		}

		retry, after := policy(resp, prevResps...)

		if !after.IsZero() {
			rl.SetRetryAfterTime(after)
		}

		if !retry {
			return resp, err
		}

		_ = resp.Body.Close() // possible `policy` already closed the body.
		prevResps = append(prevResps, resp)

		if includeBody && req.GetBody != nil {
			req.Body, err = req.GetBody()
			if err != nil {
				return resp, err
			}
		}
	}
}
