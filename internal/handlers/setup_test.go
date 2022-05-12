package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/cepa995/go-web-template/internal/config"
	"github.com/cepa995/go-web-template/internal/helpers"
	render "github.com/cepa995/go-web-template/internal/render"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/justinas/nosurf"
)

var app config.AppConfig
var session *scs.SessionManager
var pathToTemplates = "./../../templates"

//template.FuncMap is map of custom functions that we can use in a particular TEMPLATE (usually functions that are not built in the templating language)
var functions = template.FuncMap{}

// CreateTemplateCache creates a Template Cache map[string]*tempalte.Template{} which stores all application templates in memory
// and makes them easier to load; loading from internal memory is faster then loading from disk each time.
func CreateTestTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}

	// Step 1. Select everything inside ./templates/ directory that starts with *.page.gohtml
	pages, err := filepath.Glob(fmt.Sprintf("%s/*page.gohtml", pathToTemplates))
	if err != nil {
		return nil, err
	}

	// Step 2. Parse each template page and store it in the map[string]*template.Template
	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Step 2.1. Select everything inside ./templates/ directory that starts with *.layout.gohtml
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.gohtml", pathToTemplates))
		if err != nil {
			return nil, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.gohtml", pathToTemplates))
			if err != nil {
				return nil, err
			}
		}
		myCache[name] = ts
	}
	return myCache, nil
}

// NoSurf ceate new CSRF handler by utilizing github.com/justinas/nosurf package
// and set base cookie. This middleware allows us to ignore any POST request that
// does not have proper CSRF token.
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})

	return csrfHandler
}

// SessionLoad load current session.
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

// StopPageCache tries to stop browser from caching pages
func StopPageCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Retry-After", "300")
		w.Header().Set("Expires", time.Unix(0, 0).Format(http.TimeFormat))

		next.ServeHTTP(w, r)
	})
}

// IsAdmin checks if logged in user has Administrative access level for accesing particular page.
func IsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.CheckAuthorization(r, 3) {
			session.Put(r.Context(), "error", "Requires authorized access!")
			http.Redirect(w, r, "/unauthorized_access", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func TestMain(m *testing.M) {
	app.InProduction = false

	// Step 1. Create User Session
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// Step 3. Create Template Cache
	tc, err := CreateTestTemplateCache()
	if err != nil {
		app.ErrorLog.Fatal(fmt.Sprintf("Cannot create Template Cache due to - %v", err))
	}
	app.TemplateCache = tc
	// We do not want to rebuild the page on every request because when rebuilding the page it will call CreateTemplateCache from render package
	app.UseCache = true

	repo := NewTestingRepo(&app)
	NewHandlers(repo)
	render.NewRenderer(&app)
	helpers.NewHelpers(&app)
	render.NewRenderer(&app)

	os.Exit(m.Run())
}

func getRoutes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	// We DO NOT want to use NoSurf while testing handlers - it expects CSRF token during POST requests
	//mux.Use(NoSurf)
	mux.Use(SessionLoad)
	//mux.Use(StopPageCache)
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Cache-Control", "Pragma", "Expires"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	assetsFileServer := http.FileServer(http.Dir("./assets/"))
	mux.Handle("/assets/*", http.StripPrefix("/assets", assetsFileServer))

	mux.Get("/", Repo.Home)
	mux.Route("/auth", func(mux chi.Router) {
		mux.Get("/", Repo.ShowAuth)

		mux.Post("/signin", Repo.PostSignIn)
		mux.Post("/signup", Repo.PostSignUp)
		mux.Get("/signout", Repo.SignOut)

		mux.Get("/activate-accoutn", Repo.ShowActivateUserAccount)
		mux.Post("/activate-accoutn", Repo.ActivateUserAccount)

		mux.Get("/forgot-password", Repo.ForgotPassword)
		mux.Get("/reset-password", Repo.ShowResetPassword)
		mux.Post("/reset-password", Repo.ResetPassword)

	})

	return mux
}
