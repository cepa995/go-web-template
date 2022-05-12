package main

import (
	"net/http"

	"github.com/cepa995/go-web-template/internal/config"
	"github.com/cepa995/go-web-template/internal/handlers"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

// routes creates new chi router and specifies which middleware our application
// is currently using.
func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
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

	mux.Get("/", handlers.Repo.Home)
	mux.Route("/auth", func(mux chi.Router) {
		mux.Get("/", handlers.Repo.ShowAuth)

		mux.Post("/signin", handlers.Repo.PostSignIn)
		mux.Post("/signup", handlers.Repo.PostSignUp)
		mux.Get("/signout", handlers.Repo.SignOut)

		mux.Get("/activate-accoutn", handlers.Repo.ShowActivateUserAccount)
		mux.Post("/activate-accoutn", handlers.Repo.ActivateUserAccount)

		mux.Get("/forgot-password", handlers.Repo.ForgotPassword)
		mux.Get("/reset-password", handlers.Repo.ShowResetPassword)
		mux.Post("/reset-password", handlers.Repo.ResetPassword)

	})

	return mux
}
