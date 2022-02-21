package testutils

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

func RandomSleeper(milliseconds int32) func() {
	rl := sync.Mutex{}
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	sleepRandomly := func() {
		rl.Lock()
		d := rand.Int31n(500)
		rl.Unlock()
		time.Sleep(time.Duration(d) * time.Millisecond)
	}
	return sleepRandomly
}

// Returns a channel that returns true n times, before returning false
func Repeater(n int) chan bool {
	c := make(chan bool)
	go func(n int) {
		for i := 0; i < n; i++ {
			c <- true
		}
		c <- false
		close(c)
	}(n)
	return c
}

// StubResponse builds a mock http.Response.
func StubResponse(status int, body string, headerKeysAndValues ...string) *http.Response {
	resp := http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}

	isKey := true
	key := ""
	for _, s := range headerKeysAndValues {
		if isKey {
			key = s
		} else {
			resp.Header.Add(key, s)
		}
		isKey = !isKey
	}

	return &resp
}
