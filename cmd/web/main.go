package main

import (
	"forum/cmd/utils"
	"forum/internal/database"
	"log"
	"net/http"
	"os"
)

type Application struct {
	Log              Logger
	UserSessionCache map[string]User
	Store            *database.DataStore
}

type Logger struct {
	Info, Warn, Error *log.Logger
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

	logInfo := log.New(os.Stdout, "SkufInfo: ", log.Ldate|log.Ltime|log.Llongfile)
	logWarn := log.New(os.Stdout, "SkufWarning: ", log.Ldate|log.Ltime|log.Llongfile)
	logError := log.New(os.Stderr, "SkufError: ", log.Ldate|log.Ltime|log.Llongfile)

	app := Application{
		Log:              Logger{logInfo, logWarn, logError},
		Store:            database.CreateDataStore(db),
		UserSessionCache: make(map[string]User),
	}

	log.Println("Listening on http://localhost:8080...")
	serverErr := http.ListenAndServe(":8080", app.routes())
	log.Fatal(serverErr)
}
