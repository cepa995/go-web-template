package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"

	"github.com/cepa995/go-web-template/internal/config"
)

var app *config.AppConfig

// NewHelpers sets up app config for helpers
func NewHelpers(a *config.AppConfig) {
	app = a
}

// ClientError logs status code of an erorr which occurred on the cliend side, and reeturns http.Error with associated status text.
func ClientError(w http.ResponseWriter, status int) {
	app.InfoLog.Println("Client error with status of ", status)
	http.Error(w, http.StatusText(status), status)
}

// ServerError logs stack trace of an erorr which occurred on the server side, and reeturns http.StatusInternalServerError status code
// as well as its associated status text.
func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// CheckAuthorization checks whether logged in user is authorized to access specific page.
func CheckAuthorization(r *http.Request, accessLevel int64) bool {
	if !app.Session.Exists(r.Context(), "access_level") {
		return false
	}

	userAccessLevel := app.Session.Get(r.Context(), "access_level")
	return userAccessLevel == accessLevel
}

// ReadJSON reads a single JSON value from a request body.
func ReadJSON(w http.ResponseWriter, r *http.Request, data interface{}) error {
	maxBytes := 1048576 //1MB

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(data)
	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only have a single JSON value")
	}

	return nil
}

// WriteJSON writes arbitrary data out as JSON
func WriteJSON(w http.ResponseWriter, status int, data interface{}, headers ...http.Header) error {
	out, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for k, v := range headers[0] {
			w.Header()[k] = v
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(out)

	return nil
}
