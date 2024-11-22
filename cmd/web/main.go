package main

import (
	"forum/cmd/utils"
	"forum/internal/database"
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

func main() {

	handlerOpts := slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &handlerOpts))

	err := utils.CachingTemplates()
	if err != nil {
		logger.Error("failed to initialize templates:", err)
		os.Exit(1)
	}

	db, err := database.InitializeDB("./database.db", logger)
	if err != nil {
		logger.Error("database initialization failed", "error", err)
		os.Exit(1)
	}

	app := Application{
		Log:              logger,
		Store:            database.CreateDataStore(db, logger),
		UserSessionCache: make(map[string]User),
	}

	logger.Info("Listening on https://localhost:8443...")

	certFile := "tls/cert.pem"
	keyFile := "tls/key.pem"

	serverErr := http.ListenAndServeTLS(":8443", certFile, keyFile, app.routes())
	logger.Error(serverErr.Error())
	os.Exit(1)
}
