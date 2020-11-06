package httpclient

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ObserveOption .
type ObserveOption func(name string, r *http.Request, w *http.Response) prometheus.Labels

// DefaultObserveOption .
func DefaultObserveOption(name string, r *http.Request, w *http.Response) prometheus.Labels {
	return map[string]string{
		"name":   name,
		"scheme": r.URL.Scheme,
		"host":   r.URL.Host,
		"path":   r.URL.Path,
		"method": r.Method,
		"code":   fmt.Sprint(w.StatusCode),
	}
}

// RegexedObserveOption option add regex format for changing unique url path into general one.
// Make sure to sort regex descending (longest to shortest)
//
// `ex: /user/123 to /user/{userId}`
func RegexedObserveOption(regs map[string]string) func(name string, r *http.Request, w *http.Response) prometheus.Labels {
	return func(name string, r *http.Request, w *http.Response) prometheus.Labels {
		path := r.URL.Path
		for reg, p := range regs {
			match, _ := regexp.MatchString(reg, r.URL.Path)
			if match {
				path = p
				break
			}
		}

		return map[string]string{
			"name":   name,
			"scheme": r.URL.Scheme,
			"host":   r.URL.Host,
			"path":   path,
			"method": r.Method,
			"code":   fmt.Sprint(w.StatusCode),
		}
	}
}

// NewInstrumentation .
func NewInstrumentation(histogram *prometheus.HistogramVec, name string, c *http.Client, opts ...ObserveOption) *http.Client {
	std := &http.Client{Timeout: 30 * time.Second}
	transport := http.DefaultTransport
	if c != nil {
		std = c

		if std.Transport != nil {
			transport = std.Transport
		}
	}

	if len(opts) == 0 {
		opts = append(opts, DefaultObserveOption)
	}

	std.Transport = instrumentRoundTripper(histogram, name, transport, opts...)

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

	hv := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "outgoing_http_request_duration_seconds",
		Help:    "observe elapsed time in seconds for a outgoing request",
		Buckets: []float64{0.5, 1, 15, 30, 60},
	}, []string{"name", "scheme", "host", "path", "method", "code"})

	std.Transport = instrumentRoundTripper(hv, name, transport, DefaultObserveOption)

	return c
}

func instrumentRoundTripper(hv *prometheus.HistogramVec, name string, next http.RoundTripper, opts ...ObserveOption) RoundTripperFunc {
	return RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		now := time.Now()
		res, err := next.RoundTrip(req)
		if err == nil {
			var labels map[string]string
			for _, opt := range opts {
				labels = opt(name, req, res)
			}

			obs, _ := hv.GetMetricWith(labels)
			if obs != nil {
				obs.Observe(time.Since(now).Seconds())
			}
		}

		return res, err
	})
}
