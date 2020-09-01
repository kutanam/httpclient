package httpclient

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func requestDuration(histogram *prometheus.HistogramVec, name, scheme, host, path, method string, status int, startTime time.Time) {
	duration := time.Since(startTime)

	histogram.With(
		prometheus.Labels{
			"name":   name,
			"scheme": scheme,
			"host":   host,
			"path":   path,
			"method": method,
			"code":   fmt.Sprint(status),
		},
	).Observe(duration.Seconds())
}

// NewInstrumentation .
func NewInstrumentation(histogram *prometheus.HistogramVec, name string, c *http.Client) *http.Client {
	std := &http.Client{Timeout: 30 * time.Second}
	transport := http.DefaultTransport
	if c != nil {
		std = c

		if std.Transport != nil {
			transport = std.Transport
		}
	}

	std.Transport = instrumentRoundTripper(histogram, name, transport)

	return c
}

// NewWithDefaultInstrumentation .
func NewWithDefaultInstrumentation(name string, c *http.Client) *http.Client {
	std := &http.Client{Timeout: 30 * time.Second}
	transport := http.DefaultTransport
	if c != nil {
		std = c

		if std.Transport != nil {
			transport = std.Transport
		}
	}

	var metric = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "outgoing_http_request_duration_seconds",
		Help:    "observe elapsed time in seconds for a outgoing request",
		Buckets: []float64{0.5, 1, 15, 30, 60},
	}, []string{"name", "scheme", "host", "path", "method", "code"})

	std.Transport = instrumentRoundTripper(metric, name, transport)

	return c
}

func instrumentRoundTripper(histogram *prometheus.HistogramVec, name string, next http.RoundTripper) RoundTripperFunc {
	return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		now := time.Now()
		res, err := next.RoundTrip(r)
		if err == nil {
			requestDuration(histogram, name, r.URL.Scheme, r.URL.Host, r.URL.Path, r.Method, res.StatusCode, now)
		}

		return res, err
	})
}
