package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlerHome)
	mux.HandleFunc("/post", handlerPost)
	mux.HandleFunc("/create", handlerCreatePost)

	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	log.Println("Listening on http://localhost:8080...")
	serverErr := http.ListenAndServe(":8080", mux)
	log.Fatal(serverErr)
}
