package forms

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Valid(t *testing.T) {
	r := httptest.NewRequest("GET", "/some-url", nil)
	form := New(r.PostForm)

	isValid := form.Valid()
	if !isValid {
		t.Error("form is invalid!")
	}
}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("GET", "/some-url", nil)
	form := New(r.PostForm)

	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("form shows valid when required fields are missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")

	r, _ = http.NewRequest("POST", "/some-url", nil)

	r.PostForm = postedData
	form = New(r.PostForm)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("form shows invalid when required fields are not missing")
	}
}

func TestForm_Has(t *testing.T) {
	r := httptest.NewRequest("GET", "/some-url", nil)
	form := New(r.PostForm)

	if form.Has("a") {
		t.Error("form shows that it contains field 'a' when it should not.")
	}

	r, _ = http.NewRequest("POST", "/some-url", nil)

	postedData := url.Values{}
	postedData.Add("a", "a")

	r.PostForm = postedData
	form = New(r.PostForm)

	if !form.Has("a") {
		t.Error("form shows that it does NOT contain field 'a' when it should")
	}
}

func TestForm_IsEmail(t *testing.T) {
	r, _ := http.NewRequest("POST", "/some-url", nil)

	postedData := url.Values{}
	postedData.Add("correct_email", "test@gmail.com")
	postedData.Add("incorrect_email", "test@gmail")

	r.PostForm = postedData
	form := New(r.PostForm)

	form.IsEmail("correct_email")
	if len(form.Errors) != 0 {
		t.Error("form shows that email is incorrect when it shouldn't")
	}

	form.IsEmail("incorrect_email")
	if len(form.Errors) == 0 {
		t.Error("form shows that email is correct when it isn't")
	}
}

func TestForm_MinLength(t *testing.T) {
	r, _ := http.NewRequest("POST", "/some-url", nil)

	postedData := url.Values{}
	postedData.Add("a", "abc")

	r.PostForm = postedData
	form := New(r.PostForm)

	if form.MinLength("a", 5) {
		t.Error("form shows that field 'a' satisfies minimum length of 5 when it shouldn't")
	} else if !form.MinLength("a", 2) {
		t.Error("form shows that field 'a' does not satisfy minimum length of 2 when it should")
	}
}

func TestForm_MinValueInt64(t *testing.T) {
	r, _ := http.NewRequest("POST", "/some-url", nil)

	postedData := url.Values{}
	postedData.Add("a", "10")

	r.PostForm = postedData
	form := New(r.PostForm)

	if form.MinValueInt64("a", int64(15)) {
		t.Error("form shows that field 'a' satisfies minimum value of 15 when it shouldn't")
	} else if !form.MinValueInt64("a", int64(2)) {
		t.Error("form shows that field 'a' doesn't satisfy minimum value of 2 when it should")
	}
}
func TestForm_MinValueFloat64(t *testing.T) {
	r, _ := http.NewRequest("POST", "/some-url", nil)

	postedData := url.Values{}
	postedData.Add("a", "10.5")

	r.PostForm = postedData
	form := New(r.PostForm)

	if form.MinValueFloat64("a", float64(10.6)) {
		t.Error("form shows that field 'a' satisfies minimum value of 15 when it shouldn't")
	} else if !form.MinValueFloat64("a", float64(10.49)) {
		t.Error("form shows that field 'a' doesn't satisfy minimum value of 2 when it should")
	}
}
