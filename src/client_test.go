package ratelimit

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gabehardgrave/ratelimit/src/internal/aychttp"
	"github.com/gabehardgrave/ratelimit/src/internal/testutils"
	"github.com/gabehardgrave/ratelimit/src/internal/tyme"
	"github.com/stretchr/testify/assert"
)

func TestGetRetries(t *testing.T) {
	c := clientWithPolicy(retryImmedietly)
	b := testutils.Repeater(1)

	ratelimited := false
	c.stubRequest(func(req *http.Request) (*http.Response, error) {
		url := req.URL.String()
		assert.Equal(t, url, "https://server.io/endpoint")

		if <-b {
			ratelimited = true
			return testutils.StubResponse(429, "rate limited!"), nil
		}

		assert.True(t, ratelimited)
		return testutils.StubResponse(200, "success"), nil
	})

	resp, err := c.Get("https://server.io/endpoint")
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestPostRetriesSendsBody(t *testing.T) {
	c := clientWithPolicy(retryImmedietly)
	b := testutils.Repeater(2)

	c.stubRequest(func(req *http.Request) (*http.Response, error) {
		if <-b {
			assert.Equal(t, "the body", bod(req))
			assert.Equal(t, "text", req.Header.Get("content-type"))
			return testutils.StubResponse(429, "rate limited"), nil
		}

		assert.Equal(t, "text", req.Header.Get("content-type"))
		assert.Equal(t, "the body", bod(req))
		return testutils.StubResponse(200, "success"), nil
	})

	resp, err := c.Post("https://server.io/endpoint", "text", toReader("the body"))
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 200)
}

func TestHonorsRateLimits(t *testing.T) {
	c := clientWithPolicy(RetryAfterDurationInHeader)
	b := testutils.Repeater(1)

	ratelimited := false
	c.stubRequest(func(req *http.Request) (*http.Response, error) {
		if <-b {
			ratelimited = true
			return testutils.StubResponse(429, "rate limited!",
				"Retry-After", "10"), nil
		}

		assert.True(t, ratelimited)
		return testutils.StubResponse(200, "success"), nil
	})

	now := time.Now()
	howManyTimesWasSleepCalled := 0
	sleep := func(d time.Duration) {
		if d > 0 {
			howManyTimesWasSleepCalled += 1
			assert.Equal(t, 10*time.Second, d)
		}
	}

	var resp *http.Response
	var err error

	tyme.FreezeTimeAt(now, func() {
		tyme.StubSleep(sleep, func() {
			resp, err = c.Get("https://server.io/endpoint")
		})
	})

	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	assert.Equal(t, howManyTimesWasSleepCalled, 1)
}

// ################################
// ######### Helper Shit ##########
// ################################

// ClientWithPolicy returns a rate limiting client with the given policy.
func clientWithPolicy(policy RetryAfterPolicy) *Client {
	return &Client{
		RetryAfterPolicy: policy,
	}
}

// Useful to keep the tests snappy
func retryImmedietly(resp *http.Response, _ ...*http.Response) (bool, time.Time) {
	return aychttp.IsRetryable(resp), time.Time{}
}

// StubRequest will stub the internal http.Client's transport to use the given function.
// Useful for stubbing responses used by the client.
func (c *Client) stubRequest(rtf func(r *http.Request) (*http.Response, error)) {
	c.C.Transport = roundTripFunc(rtf)
}

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func bod(req *http.Request) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	req.Body.Close()
	return buf.String()
}

func toReader(s string) io.Reader {
	return strings.NewReader(s)
}
