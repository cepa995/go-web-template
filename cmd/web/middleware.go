package main

import (
	"net/http"
	"time"

	"github.com/cepa995/go-web-template/internal/helpers"
	"github.com/justinas/nosurf"
)

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
