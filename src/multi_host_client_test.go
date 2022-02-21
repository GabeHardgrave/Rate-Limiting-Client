package ratelimit

import (
	"net/http"
	"testing"
	"time"

	"github.com/gabehardgrave/ratelimit/src/internal/testutils"
	"github.com/gabehardgrave/ratelimit/src/internal/tyme"
	"github.com/stretchr/testify/assert"
)

func TestPerHostRateLimiting(t *testing.T) {
	c := multiHostClientWithPolicy(RetryAfterDurationInHeader)
	b1 := testutils.Repeater(1)
	b2 := testutils.Repeater(1)

	c.stubRequest(func(req *http.Request) (*http.Response, error) {
		url := req.URL.String()

		if url == "https://site1.com/index" {
			if <-b1 {
				return testutils.StubResponse(429, "", "Retry-After", "10"), nil
			}
			return testutils.StubResponse(200, ""), nil
		} else if url == "https://site2.com/index" {
			if <-b2 {
				return testutils.StubResponse(429, "", "Retry-After", "5"), nil
			}
			return testutils.StubResponse(200, ""), nil
		} else {
			assert.FailNow(t, "BAD SITE '%s'", url)
		}
		return nil, nil
	})

	now := time.Now()

	sleep1 := func(d time.Duration) {
		if d > 0 {
			assert.Equal(t, 10*time.Second, d)
		}
	}

	tyme.FreezeTimeAt(now, func() {
		tyme.StubSleep(sleep1, func() {
			resp, _ := c.Get("https://site1.com/index")
			assert.Equal(t, 200, resp.StatusCode)
		})
	})

	sleep2 := func(d time.Duration) {
		if d > 0 {
			assert.Equal(t, 5*time.Second, d)
		}
	}

	tyme.FreezeTimeAt(now, func() {
		tyme.StubSleep(sleep2, func() {
			resp, _ := c.Get("https://site2.com/index")
			assert.Equal(t, 200, resp.StatusCode)
		})
	})
}

func multiHostClientWithPolicy(policy RetryAfterPolicy) *MultiHostClient {
	return &MultiHostClient{
		RetryAfterPolicy: policy,
	}
}

func (c *MultiHostClient) stubRequest(rtf func(r *http.Request) (*http.Response, error)) {
	c.C.Transport = roundTripFunc(rtf)
}
