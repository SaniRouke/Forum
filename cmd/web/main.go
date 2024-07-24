package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handlerHome)

	log.Println("Listening on http://localhost:8080...")
	serverErr := http.ListenAndServe(":8080", mux)
	log.Fatal(serverErr)
}
