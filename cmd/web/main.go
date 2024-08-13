package main

import (
	"forum/cmd/handlers"
	"forum/cmd/utils"
	"forum/database"
	"log"
	"net/http"
)

func main() {

	err := utils.CachingTemplates()
	if err != nil {
		log.Fatal("Failed to initialize templates:", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", hdl.HandlerHome) // panic: pattern "/static/"  conflicts with pattern "GET /"
	mux.HandleFunc("GET /post", hdl.HandlerPost)
	mux.HandleFunc("GET /create", hdl.HandlerCreatePost)
	mux.HandleFunc("POST /create", hdl.HandlerCreatePost)
	mux.HandleFunc("POST /delete", hdl.HandlerDeletePost)
	mux.HandleFunc("GET /edit", hdl.HandlerEditPost)
	mux.HandleFunc("POST /edit", hdl.HandlerEditPost)
	mux.HandleFunc("GET /login", hdl.HandlerLogin)
	mux.HandleFunc("POST /login", hdl.HandlerLogin)
	mux.HandleFunc("GET /signup", hdl.HandlerSignup)
	mux.HandleFunc("POST /signup", hdl.HandlerSignup)

	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static", utils.Neuter(fileServer)))

	err = db.InitializeDB("./database.db")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Listening on http://localhost:8080...")
	serverErr := http.ListenAndServe(":80", mux)
	log.Fatal(serverErr)
}
