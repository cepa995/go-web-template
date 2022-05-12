package main

import (
	"fmt"
	"net/http"
	"testing"
)

func TestNoSurf(t *testing.T) {
	var dh dummyHandler

	h := NoSurf(&dh)

	// Do something based on type of 'h'
	switch v := h.(type) {
	case http.Handler:
		// do nothing
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, but is %t", v))
	}
}

func TestSessionLoad(t *testing.T) {
	var dh dummyHandler

	h := SessionLoad(&dh)

	// Do something based on type of 'h'
	switch v := h.(type) {
	case http.Handler:
		// do nothing
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, but is %t", v))
	}
}

func TestStopPageCache(t *testing.T) {
	var dh dummyHandler

	h := StopPageCache(&dh)

	// Do something based on type of 'h'
	switch v := h.(type) {
	case http.Handler:
		// do nothing
	default:
		t.Error(fmt.Sprintf("type is not http.Handler, but is %t", v))
	}
}
