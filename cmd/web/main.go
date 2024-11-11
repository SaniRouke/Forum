package main

import (
	"forum/cmd/utils"
	"forum/internal/database"
	"log"
	"log/slog"
	"net/http"
	"os"
)

type Application struct {
	Log              *slog.Logger
	UserSessionCache map[string]User
	Store            *database.DataStore
}

type User struct {
	ID     int
	Name   string
	IsAuth bool
	Token  string
}

func main() { //TODO: добавить логгер

	//handlerOpts := slog.HandlerOptions{
	//	Level: slog.LevelInfo,
	//}
	//logger := slog.New(slog.NewTextHandler(os.Stdout, &handlerOpts))

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	err := utils.CachingTemplates()
	if err != nil {
		logger.Error("Failed to initialize templates:", err)
		os.Exit(1)
	}

	db, err := database.InitializeDB("./database.db", logger)
	if err != nil {
		logger.Error("Database initialization failed", "error", err)
		os.Exit(1)
	}

	app := Application{
		Log:              logger,
		Store:            database.CreateDataStore(db),
		UserSessionCache: make(map[string]User),
	}

	log.Println("Listening on http://localhost:8080...")
	serverErr := http.ListenAndServe(":8080", app.routes())
	log.Fatal(serverErr)
}
