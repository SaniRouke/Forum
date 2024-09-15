package main

import (
	"forum/cmd/utils"
	"forum/internal"
	"log"
	"net/http"
)

type Application struct {
	User User
}

type User struct {
	Name   string
	IsAuth bool
}

func main() {

	err := utils.CachingTemplates()
	if err != nil {
		log.Fatal("Failed to initialize templates:", err)
	}

	app := Application{}

	mux := http.NewServeMux()

	mux.HandleFunc("/", app.handlerHome) // panic: pattern "/static/"  conflicts with pattern "GET /"
	mux.HandleFunc("GET /post", app.handlerPostView)
	mux.HandleFunc("POST /comment", app.handlerComment)
	mux.HandleFunc("GET /create", app.handlerCreatePost)
	mux.HandleFunc("POST /create", app.handlerCreatePost)
	mux.HandleFunc("GET /signup", app.handlerSignup)
	mux.HandleFunc("POST /signup", app.handlerSignup)
	mux.HandleFunc("GET /login", app.handlerLogin)
	mux.HandleFunc("POST /login", app.handlerLogin)
	mux.HandleFunc("POST /logout", app.handlerLogout)

	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static", utils.Neuter(fileServer)))

	err = internal.InitializeDB("./database.db")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening on http://localhost:8080...")
	serverErr := http.ListenAndServe(":8080", mux)
	log.Fatal(serverErr)
}
