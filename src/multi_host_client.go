package ratelimit

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// MultiHostClient is a wrapper over http.Client that retries requests and honors rate limits.
// It exposes an identical API to http.Client, and can be used as a drop-in replacement.
//
// The zero value uses the net/http.Client's zero value, and uses the IdiomaticRetryAfter
// RetryAfterPolicy.
//
// MultiHostClient tracks rate limits separately per host. For Clients that make requests against
// a single host, ratelimit.Client is preferred.
type MultiHostClient struct {

	// C is the http.Client used to make HTTP requests. Consumers of ratelimit.Client can
	// configure C as desired. However, in order to take advantage of ratelimit.Client's
	// automatic retry and self rate limiting, consumers should not call C's method's directly,
	// instead using ratelimit.Client's equivalents.
	C http.Client

	// RetryAfterPolicy is the policy used to determine when to retry, and for how long to wait
	// before retrying. If RetryAfterPolicy is nil, the Client will use IdiomaticRetryAfter.
	RetryAfterPolicy RetryAfterPolicy

	limiters hostRateLimiterMap
}

func (c *MultiHostClient) CloseIdleConnections() {
	c.C.CloseIdleConnections()
}

func (c *MultiHostClient) Do(req *http.Request) (*http.Response, error) {
	policy := c.RetryAfterPolicy
	if policy == nil {
		policy = IdiomaticRetryAfter
	}

	host := req.Host
	if host == "" {
		host = req.URL.Host // host might still be empty, but at least we tried.
	}

	limiter := c.limiters.HostLimiter(host)

	return limiter.do(req, &c.C, policy)
}

func (c *MultiHostClient) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *MultiHostClient) Head(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *MultiHostClient) Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}

func (c *MultiHostClient) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (c *MultiHostClient) ForgetHost(host string) {
	c.limiters.m.Delete(host)
}

// ################################
// ### private multi host stuff ###
// ################################

type hostRateLimiterMap struct {
	m sync.Map
}

func (m *hostRateLimiterMap) HostLimiter(host string) *RateLimiter {
	limiter, _ := m.m.LoadOrStore(host, &RateLimiter{})
	return limiter.(*RateLimiter)
}
