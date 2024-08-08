package main

import (
	"forum/cmd/utils"
	"forum/internal"
	"log"
	"net/http"
)

func main() {

	err := utils.CachingTemplates()
	if err != nil {
		log.Fatal("Failed to initialize templates:", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", handlerHome) // panic: pattern "/static/"  conflicts with pattern "GET /"
	mux.HandleFunc("GET /post", handlerPost)
	mux.HandleFunc("GET /create", handlerCreatePost)
	mux.HandleFunc("POST /create", handlerCreatePost)
	mux.HandleFunc("POST /delete", handlerDeletePost)
	mux.HandleFunc("GET /edit", handlerEditPost)
	mux.HandleFunc("POST /edit", handlerEditPost)

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
