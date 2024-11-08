package main

import (
	"forum/cmd/utils"
	"forum/internal/database"
	"log"
	"net/http"
)

type Application struct {
	UserSessionCache map[string]User
	Store            *database.DataStore
}

type User struct {
	ID     int
	Name   string
	IsAuth bool
	Token  string
}

func main() { //TODO: добавить логер

	err := utils.CachingTemplates()
	if err != nil {
		log.Fatal("Failed to initialize templates:", err)
	}

	db, err := database.InitializeDB("./database.db")
	if err != nil {
		log.Fatal(err)
	}
	app := Application{Store: database.CreateDataStore(db), UserSessionCache: make(map[string]User)}
	
	log.Println("Listening on http://localhost:8080...")
	serverErr := http.ListenAndServe(":8080", app.routes())
	log.Fatal(serverErr)
}
