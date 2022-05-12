package forms

import (
	"net/http/httptest"
	"testing"
)

func TestErrors_Add(t *testing.T) {
	r := httptest.NewRequest("GET", "/some-url", nil)
	form := New(r.PostForm)

	form.Errors.Add("test", "error")

	if _, ok := form.Errors["test"]; !ok {
		t.Error("form does not have expected error")
	}

	if form.Errors["test"][0] != "error" {
		t.Error("form does not have expected error message")
	}
}

func TestErrors_Get(t *testing.T) {
	r := httptest.NewRequest("GET", "/some-url", nil)
	form := New(r.PostForm)

	form.Errors.Add("test", "error")
	if form.Errors.Get("test") == "" {
		t.Error("form does not have expected error when it should")
	}

	form = New(r.PostForm)
	if form.Errors.Get("test") != "" {
		t.Error("form does have error 'test' when it shouldn't")
	}
}
