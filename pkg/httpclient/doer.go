package httpclient

import "net/http"

// Doer an interface to wrap net/http Client at first.
// Of course, net/http Client is usable without wrapper if you want to use it at the first.
// The purpose of this interface is make your implementation to third party using http method not mixed up with business things.
// Such as logging http request and response, request duration, retryable request, instrumentation, etc.
//
// Even better with this interface, you can make more wrapper for whatever your needs.
// Sample wrapper is `LoggedDoer` which logging the request.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// The RoundTripperFunc type is an adapter to allow the use of ordinary
// functions as RoundTrippers. If f is a function with the appropriate
// signature, RountTripperFunc(f) is a RoundTripper that calls f.
type RoundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the RoundTripper interface.
func (rt RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt(r)
}
