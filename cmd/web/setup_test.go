package main

import (
	"net/http"
	"os"
	"testing"
)

// Whatever is in here will run before our tests run.

func TestMain(m *testing.M) {
	// Before start running tests in main package, do something inside the function TestMain
	// then run the tests (m.Run()) and then exit!

	os.Exit(m.Run())
}

// dummyHandler is dummy type used for middleware NoSurf test - it should saitsfy everything HTTP handler does
type dummyHandler struct{}

// ServeHTTP function part of HTTP interface which our dummyHandler implements in order to replicate HTTP handler
func (dh *dummyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
