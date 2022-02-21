package aychttp

import "net/http"

func HasBody(req *http.Request) bool {
	return req.Header.Get("Content-Type") != ""
}

func IsRetryable(resp *http.Response) bool {
	status := resp.StatusCode
	return status == http.StatusTooManyRequests ||
		status == http.StatusServiceUnavailable ||
		status == http.StatusInternalServerError
}
