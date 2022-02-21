package ratelimit

import (
	"io"
	"net/http"
)

// DefaultClient is the default client used by Get, Head, and Post.
var DefaultClient = &MultiHostClient{
	C:                http.Client{},
	RetryAfterPolicy: IdiomaticRetryAfter,
}

func Get(url string) (resp *http.Response, err error) {
	return DefaultClient.Get(url)
}

func Head(url string) (resp *http.Response, err error) {
	return DefaultClient.Head(url)
}

func Post(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	return DefaultClient.Post(url, contentType, body)
}
