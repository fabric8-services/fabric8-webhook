package util

import (
	"errors"
	"net/http"
)

// RoundTripFunc needed to mock a request for url
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// SetMockNetClient Change the Transport for NetClient
func SetMockNetClient(fn RoundTripFunc) {
	NetClient.Transport = fn
}

// ErrReader is used to return error while doing reading
type ErrReader string

// Read Mock Error while reading
func (e ErrReader) Read(p []byte) (n int, err error) {
	return 0, errors.New(string(e))
}

// Close to implement Close method needed for ioutil.ReadAll
func (e ErrReader) Close() error {
	return nil
}
