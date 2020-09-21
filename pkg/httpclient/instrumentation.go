package httpclient

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
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

func pathPatternMatching(pathPatterns map[string]string, cmpPath string) string {
	for k, v := range pathPatterns {
		match, err := regexp.MatchString(v, cmpPath)
		if err != nil {
			log.SetFlags(log.Lshortfile)
			log.Println(err)

			// just return path what want to be match
			return cmpPath
		}

		// return path pattterned
		// example:
		// - /users/:id
		// - /users/{id}
		if match {
			return k
		}
	}

	return cmpPath
}

// NewInstrumentation .
func NewInstrumentation(hv *prometheus.HistogramVec, name string, pathPatterns map[string]string, c *http.Client) *http.Client {
	std := &http.Client{Timeout: 30 * time.Second}
	transport := http.DefaultTransport
	if c != nil {
		std = c

		if std.Transport != nil {
			transport = std.Transport
		}
	}

	std.Transport = instrumentRoundTripper(hv, name, pathPatterns, transport)

	return c
}

// NewWithDefaultInstrumentation .
func NewWithDefaultInstrumentation(name string, pathPatterns map[string]string, c *http.Client) *http.Client {
	std := &http.Client{Timeout: 30 * time.Second}
	transport := http.DefaultTransport
	if c != nil {
		std = c

		if std.Transport != nil {
			transport = std.Transport
		}
	}

	var hv = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "outgoing_http_request_duration_seconds",
		Help:    "observe elapsed time in seconds for a outgoing request",
		Buckets: []float64{0.5, 1, 15, 30, 60},
	}, []string{"name", "scheme", "host", "path", "method", "code"})

	std.Transport = instrumentRoundTripper(hv, name, pathPatterns, transport)

	return c
}

func instrumentRoundTripper(hv *prometheus.HistogramVec, name string, pathPatterns map[string]string, next http.RoundTripper) RoundTripperFunc {
	return RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		now := time.Now()
		res, err := next.RoundTrip(r)
		if err == nil {
			defer func() {
				path := pathPatternMatching(pathPatterns, r.URL.Path)
				requestDuration(hv, name, r.URL.Scheme, r.URL.Host, path, r.Method, res.StatusCode, now)
			}()
		}

		return res, err
	})
}
