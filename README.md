A self rate limiting and automatic-retry HTTP Client, written in Go.

## Purpose

Modern HTTP based APIs typically rate limit clients through some mechanism. This is done idiomatically using 429 status codes and a `Retry-After` header (see MDN summary [here](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Retry-After)), although different APIs may communicate a client's rate limiting status in other ways.

Even if an API does not explicitly rate limit, robust HTTP clients should still be able to handle transient server errors and service unavailability, while still being "good citizens" (so to speak) and not overwhelm the API servers.

GoRlimit makes honoring rate limits and gracefully retrying requests painless, via `ratelimit.Client` and `ratelimit.MultiHostClient`. Both HTTP clients provide an identical interface to the standard library's `http.Client`. Simply swap an instance of `http.Client` with `ratelimit.Client`, and voilà, your application is now a respectful API client that honors rate limits, and gracefully retries when feasible.

## Examples

```go

// Similar to net/http, the zero value of client is perfectly usable.
// The default client will retry on 429, 500, and 503 status codes.
// Retry-After headers (either as a duration in <seconds>, or as an <http-date>) will be honored.
// If no Retry-After header is present, the client will implement exponential backoff.
client := ratelimit.Client{}

// If "https://api.example.com/index" returns a 429, `client.Get` will wait and retry as appropriate
// until a non 429/500/503 response is returned.
resp, err := client.Get("https://api.example.com/index")
```

If you know how your API rate limits clients, you can implement a custom `RetryAfterPolicy` to configure your client.
```go
client := ratelimit.Client{
    RetryAfterPolicy: func(resp *http.Response, prevResps ...*http.Response) (bool, time.Time) {
    
        // This client will only retry automatically when the server responds with a 429,
        retry := resp.StatusCode == 429
        if !retry {
            return false, time.Time{} // use the zero time to indicate no waiting
        }
        
        // This client will wait until the time specified by X-Ratelimit-Reset-At-Date header
        // before sending another request.
        afterStr := resp.Headers.Get("X-Ratelimit-Reset-At-Date")
        after, e := http.ParseTime(after)
        if e != nil {
            // ... log the unexpected header value
            after = time.Now().Add(5 * time.Second) // default to 5 seconds in the future
        }
        
        return retry, after
    }
}

// If api.example.com first returns a 429 with a date,
// Get will block and resend at the specified time. It's
// guaranteed that `resp.StatusCode != 429` when it returns.
resp, err := client.Get("https://api.example.com/index")
```

`ratelimit.Client` is ideal if you're making requests to a single host. If your client is making requests to multiple hosts, you should use `ratelimit.MultiHostClient` (this will track and enforce rate limits separately for each host).

```go
client := ratelimit.MultiHostClient{}

// if api.foo rate limits and api.bar does not,
// the multi host client will not block requests
// to api.bar
go client.Get("https://api.foo.com") 
go client.Get("https://api.bar.com")
```
