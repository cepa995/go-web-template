package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	params             []postData
	expectedStatusCode int
}{
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"auth", "/auth", "GET", []postData{}, http.StatusOK},
	{"signout", "/auth/signout", "GET", []postData{}, http.StatusOK},
	{"signin", "/auth/signin", "POST", []postData{
		{key: "email", value: "test@gmail.com"},
		{key: "password", value: "test123"},
	}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()

	// Step 1. Create Test Server
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	// Step 2. For each test case perform GET/POST request
	for _, e := range theTests {
		if e.method == "GET" {
			resp, err := ts.Client().Get(ts.URL + e.url)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}
			// Evaluate Response
			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}
		} else {
			values := url.Values{}
			for _, v := range e.params {
				values.Add(v.key, v.value)
			}
			// PostForm requiers 2x things: URL to which we want to post to + values we want to post
			resp, err := ts.Client().PostForm(ts.URL+e.url, values)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			if resp.StatusCode != e.expectedStatusCode {
				t.Errorf("for %s, expected %d but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
			}
		}
	}
}

func getCtx(req *http.Request) context.Context {
	// We need Header with X-Session in order to READ FROM and WRITE TO a session
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		log.Println(err)
	}
	return ctx
}

var loginTests = []struct {
	name               string
	email              string
	password           string
	expectedStatusCode int
	expectedHTML       string
	expectedLocation   string
}{
	{
		"valid-credentials",
		"test@gmail.com",
		"password",
		http.StatusSeeOther,
		"",
		"/",
	},
	{
		"requires-password",
		"test@gmail.com",
		"",
		http.StatusOK,
		`action='/auth/signin'`,
		"", // We are not doing REDIRECT in this case (so we are not getting Location in Result) but we are RENDERING a page (html)
	},
	{
		"requires-email",
		"",
		"password",
		http.StatusOK,
		`action='/auth/signin'`,
		"", // We are not doing REDIRECT in this case (so we are not getting Location in Result) but we are RENDERING a page (html)
	},
	{
		"invalid-credentials-pt1",
		"test@hotmail.com",
		"password",
		http.StatusSeeOther,
		"",
		"/auth",
	},
	{
		"invalid-credentials-pt2",
		"test@gmail.com",
		"wrong_password",
		http.StatusSeeOther,
		"",
		"/auth",
	},
}

func TestSignIn(t *testing.T) {
	for _, e := range loginTests {
		postedData := url.Values{}
		postedData.Add("email", e.email)
		postedData.Add("password", e.password)

		// Create the request
		req, _ := http.NewRequest("POST", "/auth/signin", strings.NewReader(postedData.Encode()))
		ctx := getCtx(req)
		req = req.WithContext(ctx)

		// Set the header
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		// Call the handler
		handler := http.HandlerFunc(Repo.PostSignIn)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		if e.expectedLocation != "" {
			// get URL from test
			actualLoc, _ := rr.Result().Location()
			if actualLoc.String() != e.expectedLocation {
				t.Errorf("failed %s: expected location %s, but got %s", e.name, e.expectedLocation, actualLoc.String())
			}
		}

		// Checkingfor expected values in HTML
		if e.expectedHTML != "" {
			// read the response body into a string
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s but did not", e.name, e.expectedHTML)
			}
		}
	}
}

var registerTests = []struct {
	name               string
	firstName          string
	lastName           string
	password           string
	email              string
	expectedStatusCode int
	expectedLocation   string
	expectedJSON       jsonResponse
	expectedHTML       string
}{
	{
		"valid-info",
		"Jon",
		"Doe",
		"password",
		"test1@gmail.com",
		http.StatusOK,
		"/auth",
		jsonResponse{
			OK:      true,
			Message: "Ok",
		},
		"",
	},
	{
		"invalid-info-pt1",
		"Jon",
		"Doe",
		"",
		"test2@gmail.com",
		http.StatusOK,
		"/auth",
		jsonResponse{
			OK:      false,
			Message: "Missing Password",
		},
		"",
	},
	{
		"invalid-info-pt1",
		"J",
		"Doe",
		"password",
		"test3@gmail.com",
		http.StatusOK,
		"/auth",
		jsonResponse{
			OK:      false,
			Message: "Name is too small",
		},
		"",
	},
	{
		"invalid-info-pt1",
		"Jon",
		"Doe",
		"password",
		"test@gmail.com",
		http.StatusOK,
		"/auth",
		jsonResponse{
			OK:      false,
			Message: "Email Exists",
		},
		"",
	},
}

func TestSignUp(t *testing.T) {
	for _, e := range registerTests {
		postedData := url.Values{}
		postedData.Add("firstName", e.firstName)
		postedData.Add("lastName", e.lastName)
		postedData.Add("password", e.password)
		postedData.Add("email", e.email)

		// Create Request
		req, _ := http.NewRequest("POST", "/auth/signup", nil)
		ctx := getCtx(req)
		req = req.WithContext(ctx)
		req.PostForm = postedData

		// Set Header
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Call the handler
		handler := http.HandlerFunc(Repo.PostSignUp)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("failed %s: expected code %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		b, err := io.ReadAll(rr.Result().Body)
		if err != nil {
			t.Error("failed to read result body")
		}

		var d jsonResponse
		if err := json.Unmarshal(b, &d); err != nil {
			fmt.Println(err)
		}

		if d.OK != e.expectedJSON.OK {
			t.Errorf("failed %s: exppected JSON response %v, but got %v", e.name, e.expectedJSON.OK, d.OK)
		}

		// Checkingfor expected values in HTML
		if e.expectedHTML != "" {
			// read the response body into a string
			html := rr.Body.String()
			if !strings.Contains(html, e.expectedHTML) {
				t.Errorf("failed %s: expected to find %s but did not", e.name, e.expectedHTML)
			}
		}
	}
}
