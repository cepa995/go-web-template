package render

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/cepa995/go-web-template/internal/config"
)

var session *scs.SessionManager
var testApp config.AppConfig

// TestMain gets called before any of the tests are run, and just before it closes it runs all of our tests.
func TestMain(m *testing.M) {
	// Dummy AppConfig variable which we will use to asign to "app" inside render.go so we can test
	testApp.InProduction = false

	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false

	testApp.Session = session

	// We need to make "app" point to "testApp" because "app" is variable we are testing in render.go
	app = &testApp
	os.Exit(m.Run())
}

// myWriter is dummy struct which will implement Header(), WriteHeader(int), Write([]byte),in order to satisfy ResponseWriter interface
type myWriter struct{}

func (tw *myWriter) Header() http.Header {
	var h http.Header
	return h
}

func (tw *myWriter) WriteHeader(i int) {

}

func (tw *myWriter) Write(b []byte) (int, error) {
	length := len(b)
	return length, nil
}
