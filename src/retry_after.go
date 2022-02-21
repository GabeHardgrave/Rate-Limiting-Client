package ratelimit

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gabehardgrave/ratelimit/src/internal/aychttp"
	"github.com/gabehardgrave/ratelimit/src/internal/maath"
	"github.com/gabehardgrave/ratelimit/src/internal/tyme"
)

var (
	// DefaultMaxRetryAfterDuration is 24*30 = 720 hours
	DefaultMaxRetryAfterDuration = 24 * 30 * time.Hour
)

// RetryAfterPolicy describes the retry and rate limiting policy for a Client. RetryAfterPolicy
// should return true if resp indicates a retry, and a non-zero time if requests should retry
// after a specific time.
type RetryAfterPolicy func(
	resp *http.Response,
	prevResps ...*http.Response,
) (retry bool, after time.Time)

// ################################
// ###### Idiomatic Policies ######
// ################################

// ExponentialBackoff implements a policy of exponential backoff. `retry` will be true if
// `resp.StatusCode` is 429, 500, or 503. If `retry` is true, `after` will be calculated
// as N seconds from `time.Now()`, where N starts at 1, and doubles for each response in
// `prevResps`.
func ExponentialBackoff(
	resp *http.Response,
	prevResps ...*http.Response,
) (retry bool, after time.Time) {

	retry = aychttp.IsRetryable(resp)
	if !retry {
		return retry, after
	}

	dur := exponentialBackoffDuration(uint64(len(prevResps)))
	after = tyme.Now().Add(dur)
	retry = (dur < DefaultMaxRetryAfterDuration)

	return retry, after
}

var _ RetryAfterPolicy = ExponentialBackoff

// RetryAfterDurationInHeader implements a policy of honoring Retry-After headers, when specified
// as duration in seconds. `retry` will be true if `resp.StatusCode` is 429, 500, or 503. If a
// Retry-After header is present and set to N, `after` will be equivalent to
// `time.Now().Add(N * time.Second)`.
func RetryAfterDurationInHeader(
	resp *http.Response,
	prevResps ...*http.Response,
) (retry bool, after time.Time) {

	dur := retryAfterDuration(resp.Header.Get("Retry-After"))
	if dur > 0 {
		after = tyme.Now().Add(dur)
	}

	retry = aychttp.IsRetryable(resp) &&
		dur < DefaultMaxRetryAfterDuration

	return retry, after
}

var _ RetryAfterPolicy = RetryAfterDurationInHeader

// RetryAfterTimeInHeader implements a policy of honoring Retry-After headers, when specified
// as an <http-date>. `retry` will be true if `resp.StatusCode` is 429, 500, or 503. If a
// Retry-After header is present and `after` will be the parsed time indicated by the <http-date>.
func RetryAfterTimeInHeader(
	resp *http.Response,
	prevResps ...*http.Response,
) (retry bool, after time.Time) {

	after = retryAfterTime(resp.Header.Get("Retry-After"))

	retry = aychttp.IsRetryable(resp) &&
		after.Sub(tyme.Now()) < DefaultMaxRetryAfterDuration

	return retry, after
}

var _ RetryAfterPolicy = RetryAfterTimeInHeader

// IdiomaticRetryAfter implements a policy of honoring Retry-After headers, when specified as
// either a duration in <seconds>, or as an <http-date>. `retry` will be true if `resp.StatusCode`
// is 429, 500, or 503.
//
// If `retry` is true, but no `Retry-After` header was specified, IdiomaticRetryAfter implements
// a policy of exponential backoff.
func IdiomaticRetryAfter(
	resp *http.Response,
	prevResps ...*http.Response,
) (retry bool, after time.Time) {

	retry = aychttp.IsRetryable(resp)
	retryAfterStr := resp.Header.Get("Retry-After")

	// It's possible for `Retry-After` to be specified, even if retry is false. Hence we only
	// return early if retryAfterStr is unspecified.
	if !retry && retryAfterStr == "" {
		return retry, after
	}

	d := retryAfterDuration(retryAfterStr)
	if d != 0 {
		after = tyme.Now().Add(d)
	} else {
		after = retryAfterTime(retryAfterStr)
		d = tyme.Until(after)
	}

	if retry && after.IsZero() {
		d = exponentialBackoffDuration(uint64(len(prevResps)))
		after = tyme.Now().Add(d)
	}

	retry = retry && (d < DefaultMaxRetryAfterDuration)
	return retry, after
}

var _ RetryAfterPolicy = IdiomaticRetryAfter

// ################################
// ######### Private Shit #########
// ################################

func retryAfterTime(header string) (t time.Time) {
	if header == "" { // quickly catch missing header
		return t
	}

	t, err := http.ParseTime(header)
	if err != nil {
		return time.Time{}
	}
	return t
}

func retryAfterDuration(header string) (d time.Duration) {
	if header == "" { // quickly catch missing header
		return d
	}

	seconds, err := strconv.ParseUint(header, 10, 64)
	if err == nil {
		d = time.Duration(seconds) * time.Second
	}
	return d
}

func exponentialBackoffDuration(prevReqCount uint64) time.Duration {
	nSec := maath.MaxPowerOf2(prevReqCount)
	return time.Duration(nSec) * time.Second
}
