package httpclient

import (
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// RetryableConfig .
type RetryableConfig struct {
	RetryMax     int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
}

// DefaultRetryableConfig .
func DefaultRetryableConfig() RetryableConfig {
	return RetryableConfig{
		RetryMax:     3,
		RetryWaitMin: 15000 * time.Millisecond,
		RetryWaitMax: 30000 * time.Millisecond,
	}
}

// NewRetryable this is wrapper for `go-retryablehttp`.
// which has base from `net/http.Client` itself
// Make it as `net/http.Client` again
func NewRetryable(conf RetryableConfig, c *http.Client) *http.Client {
	retrier := retryablehttp.NewClient()

	// default http client
	retrier.HTTPClient = &http.Client{Timeout: 30 * time.Second}

	// overriding http client if exists
	if c != nil {
		retrier.HTTPClient = c
	}

	// no logger applied
	retrier.Logger = nil

	retrier.Backoff = retryablehttp.LinearJitterBackoff
	retrier.RetryMax = conf.RetryMax
	retrier.RetryWaitMin = conf.RetryWaitMin
	retrier.RetryWaitMax = conf.RetryWaitMax
	retrier.ErrorHandler = retryablehttp.PassthroughErrorHandler

	std := retrier.StandardClient()

	return std
}
