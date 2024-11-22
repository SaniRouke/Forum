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

	mux.HandleFunc("/", app.handlerHome)
	mux.HandleFunc("GET /post", app.handlerPostView)
	mux.HandleFunc("GET /user", app.authMW(app.handlerUserPage))
	mux.HandleFunc("POST /react", app.handlerReactToPost)
	mux.HandleFunc("POST /comment", app.handlerComment)
	mux.HandleFunc("POST /comment-react", app.handlerReactToComment)
	mux.HandleFunc("GET /create", app.authMW(app.handlerCreatePost))
	mux.HandleFunc("POST /create", app.authMW(app.handlerCreatePost))
	mux.HandleFunc("GET /signup", app.handlerSignup)
	mux.HandleFunc("POST /signup", app.handlerSignup)
	mux.HandleFunc("GET /login", app.handlerLogin)
	mux.HandleFunc("POST /login", app.handlerLogin)
	mux.HandleFunc("POST /logout", app.handlerLogout)

	return mux
}
