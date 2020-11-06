package httpclient_test

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/google/go-cmp/cmp"

	"github.com/payfazz/httpclient/pkg/httpclient"
)

func Test_RegexedObserveOption_Descending(t *testing.T) {
	host := "https://example.com"
	uniqueUserID := strings.ReplaceAll(fmt.Sprint(regexp.QuoteMeta("/user/"), `\d?.+`), `/`, `\/`)
	subserviceAfterUniqueUserID := strings.ReplaceAll(fmt.Sprint(regexp.QuoteMeta("/user/"), `\d?.+`, `/subservice`), `/`, `\/`)

	sampleRegex := map[string]string{
		subserviceAfterUniqueUserID: "/user/{userId}/subservice",
		uniqueUserID:                "/user/{userId}",
	}

	cases := map[string]struct {
		expectedRes prometheus.Labels
		path        string
	}{
		"unique id": {
			path: "/user/123",
			expectedRes: map[string]string{
				"name":   "example",
				"scheme": "https",
				"host":   "example.com",
				"path":   "/user/{userId}",
				"method": http.MethodGet,
				"code":   fmt.Sprint(http.StatusOK),
			},
		},
		"subservice after unique id": {
			path: "/user/123/subservice",
			expectedRes: map[string]string{
				"name":   "example",
				"scheme": "https",
				"host":   "example.com",
				"path":   "/user/{userId}/subservice",
				"method": http.MethodGet,
				"code":   fmt.Sprint(http.StatusOK),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", host, tc.path), nil)
			fn := httpclient.RegexedObserveOption(sampleRegex)

			result := fn("example", req, &http.Response{StatusCode: http.StatusOK})
			if diff := cmp.Diff(tc.expectedRes, result); diff != "" {
				t.Fatalf("[%s] mismatch (-want +got):\n%s", t.Name(), diff)
			}
		})
	}
}

func Test_RegexedObserveOption_Ascending(t *testing.T) {
	host := "https://example.com"
	uniqueUserID := strings.ReplaceAll(fmt.Sprint(regexp.QuoteMeta("/user/"), `\d?.+`), `/`, `\/`)
	subserviceAfterUniqueUserID := strings.ReplaceAll(fmt.Sprint(regexp.QuoteMeta("/user/"), `\d?.+`, `/subservice`), `/`, `\/`)

	sampleRegex := map[string]string{
		uniqueUserID:                "/user/{userId}",
		subserviceAfterUniqueUserID: "/user/{userId}/subservice",
	}

	cases := map[string]struct {
		expectedRes prometheus.Labels
		path        string
	}{
		"unique id": {
			path: "/user/123",
			expectedRes: map[string]string{
				"name":   "example",
				"scheme": "https",
				"host":   "example.com",
				"path":   "/user/{userId}",
				"method": http.MethodGet,
				"code":   fmt.Sprint(http.StatusOK),
			},
		},
		"will alywas match shortest regex first": {
			path: "/user/123/subservice/subservice2",
			expectedRes: map[string]string{
				"name":   "example",
				"scheme": "https",
				"host":   "example.com",
				"path":   "/user/{userId}",
				"method": http.MethodGet,
				"code":   fmt.Sprint(http.StatusOK),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/%s", host, tc.path), nil)
			fn := httpclient.RegexedObserveOption(sampleRegex)

			result := fn("example", req, &http.Response{StatusCode: http.StatusOK})
			if diff := cmp.Diff(tc.expectedRes, result); diff != "" {
				t.Fatalf("[%s] mismatch (-want +got):\n%s", t.Name(), diff)
			}
		})
	}
}
