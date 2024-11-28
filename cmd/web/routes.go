package main

import (
	"net/http"
)

func (app *Application) routes() *http.ServeMux {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static", app.Neuter(fileServer)))

	imageServer := http.FileServer(http.Dir("./uploads"))
	mux.Handle("/uploads/", http.StripPrefix("/uploads", app.Neuter(imageServer)))

	mux.HandleFunc("/auth/google/login", app.rateLimiterMW(app.handleGoogleLogin))
	mux.HandleFunc("/auth/google/callback", app.rateLimiterMW(app.handleGoogleCallback))

	mux.HandleFunc("/auth/github/login", app.rateLimiterMW(app.handleGithubLogin))
	mux.HandleFunc("/auth/github/callback", app.rateLimiterMW(app.handleGithubCallback))

	mux.HandleFunc("/", app.rateLimiterMW(app.handlerHome))
	mux.HandleFunc("GET /post", app.rateLimiterMW(app.handlerPostView))
	mux.HandleFunc("GET /user", app.rateLimiterMW(app.authMW(app.handlerUserPage)))
	mux.HandleFunc("POST /react", app.rateLimiterMW(app.rateLimiterMW(app.handlerReactToPost)))
	mux.HandleFunc("POST /comment", app.rateLimiterMW(app.handlerComment))
	mux.HandleFunc("POST /comment-react", app.rateLimiterMW(app.handlerReactToComment))
	mux.HandleFunc("GET /create", app.rateLimiterMW(app.authMW(app.handlerCreatePost)))
	mux.HandleFunc("POST /create", app.rateLimiterMW(app.authMW(app.handlerCreatePost)))
	mux.HandleFunc("GET /signup", app.rateLimiterMW(app.handlerSignup))
	mux.HandleFunc("POST /signup", app.rateLimiterMW(app.handlerSignup))
	mux.HandleFunc("GET /login", app.rateLimiterMW(app.handlerLogin))
	mux.HandleFunc("POST /login", app.rateLimiterMW(app.handlerLogin))
	mux.HandleFunc("POST /logout", app.rateLimiterMW(app.handlerLogout))

	return mux
}
