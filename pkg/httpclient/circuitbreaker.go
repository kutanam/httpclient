package httpclient

import (
	"net/http"
	"time"

	"github.com/payfazz/go-errors"
	"github.com/payfazz/stdlog"
	"github.com/sony/gobreaker"
)

// CircuitBreaker .
type CircuitBreaker struct {
	c                     *http.Client
	cb                    *gobreaker.CircuitBreaker
	failedExecutionStatus []int
}

// NewCircuitBreaker .
func NewCircuitBreaker(name string, c *http.Client) *CircuitBreaker {
	std := &http.Client{Timeout: 30 * time.Second}

	if c == nil {
		c = std
	}

	return &CircuitBreaker{
		c:                     c,
		cb:                    DefaultCircuitBreaker(name),
		failedExecutionStatus: DefaultFailedExecutionStatus(),
	}
}

// UseCircuitBreaker overriding ciruitbreaker system
func (c *CircuitBreaker) UseCircuitBreaker(cb *gobreaker.CircuitBreaker) *CircuitBreaker {
	c.cb = cb
	return c
}

// SetFailedExecutionStatus overriding failed execution status
func (c *CircuitBreaker) SetFailedExecutionStatus(status ...int) *CircuitBreaker {
	c.failedExecutionStatus = status
	return c
}

// StandardClient .
func (c *CircuitBreaker) StandardClient() *http.Client {
	next := c.c.Transport
	if next == nil {
		next = http.DefaultTransport
	}

	return &http.Client{
		Transport: circuitBreakerRoundTripper(c.cb, c.failedExecutionStatus, next),
	}
}

// Do .
func (c *CircuitBreaker) Do(req *http.Request) (*http.Response, error) {
	res, err := c.cb.Execute(func() (interface{}, error) {
		res, err := c.c.Do(req)

		if res != nil {
			for _, status := range c.failedExecutionStatus {
				if res.StatusCode == status {
					return nil, errors.Errorf("ERR EXEC (%s) [%d] %s _ %s%s", c.cb.Name(), res.StatusCode, req.Method, req.URL.Host, req.URL.Path)
				}
			}
		}

		return res, err
	})

	if res == nil {
		return nil, err
	}

	return res.(*http.Response), err
}

// DefaultCircuitBreaker .
func DefaultCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 15,
		Interval:    5000 * time.Millisecond,
		Timeout:     45 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= 0.6 || counts.ConsecutiveFailures >= 100
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			// do smth when circuit breaker trips.
			stdlog.Out().Print("circuit [%s] change state %s --> %s", name, from, to)
		},
	}

	return gobreaker.NewCircuitBreaker(settings)
}

// DefaultFailedExecutionStatus .
func DefaultFailedExecutionStatus() []int {
	return []int{
		http.StatusInternalServerError,
		http.StatusNotImplemented,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusHTTPVersionNotSupported,
		http.StatusVariantAlsoNegotiates,
		http.StatusInsufficientStorage,
		http.StatusLoopDetected,
		http.StatusNotExtended,
		http.StatusNetworkAuthenticationRequired,
	}
}

func circuitBreakerRoundTripper(cb *gobreaker.CircuitBreaker, failedExecutionStatus []int, next http.RoundTripper) RoundTripperFunc {
	return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		res, err := cb.Execute(func() (interface{}, error) {
			res, err := next.RoundTrip(req)

			if res != nil {
				for _, status := range failedExecutionStatus {
					if res.StatusCode == status {
						return nil, errors.Errorf("ERR EXEC (%s) [%d] %s _ %s%s", cb.Name(), res.StatusCode, req.Method, req.URL.Host, req.URL.Path)
					}
				}
			}

			return res, err
		})

		if res == nil {
			return nil, err
		}

		return res.(*http.Response), err
	})
}
