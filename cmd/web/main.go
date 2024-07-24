package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handlerHome)
	//mux.HandleFunc("POST /", handlerPost)

	//fileServer := http.FileServer(http.Dir("./ui/static"))
	//mux.Handle("GET /static/", http.StripPrefix("/static", utils.Neuter(fileServer)))

	log.Println("Listening on http://localhost:8080...")
	serverErr := http.ListenAndServe(":8080", mux)
	log.Fatal(serverErr)
}
