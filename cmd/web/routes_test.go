package main

import (
	"fmt"
	"testing"

	"github.com/cepa995/go-web-template/internal/config"
	"github.com/go-chi/chi"
)

func TestRoutes(t *testing.T) {
	var app config.AppConfig

	mux := routes(&app)

	switch v := mux.(type) {
	case *chi.Mux:
		// Do nothing - test passed
	default:
		t.Error(fmt.Sprintf("type is not *chiMux, type is %t", v))
	}
}
