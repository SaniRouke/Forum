package main

import (
	"forum/cmd/utils"
	"forum/internal/database"
	"log"
	"net/http"
)

type Application struct {
	User  User
	Store *database.DataStore
}

type User struct {
	ID     int
	Name   string
	IsAuth bool
}

func main() {

	err := utils.CachingTemplates()
	if err != nil {
		log.Fatal("Failed to initialize templates:", err)
	}

	db, err := database.InitializeDB("./database.db")
	if err != nil {
		log.Fatal(err)
	}
	app := Application{Store: database.CreateDataStore(db)}

	mux := http.NewServeMux()

	mux.HandleFunc("/", app.handlerHome) // panic: pattern "/static/"  conflicts with pattern "GET /"
	mux.HandleFunc("GET /post", app.handlerPostView)

	mux.HandleFunc("GET /user", app.handlerUserPage)

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

	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static", utils.Neuter(fileServer)))

	log.Println("Listening on http://localhost:8080...")
	serverErr := http.ListenAndServe(":8080", mux)
	log.Fatal(serverErr)
}
