package main

import (
	"forum/cmd/utils"
	"forum/internal/database"
	"github.com/joho/godotenv"
	"log/slog"
	"net/http"
	"os"
)

type Application struct {
	Log              *slog.Logger
	UserSessionCache map[string]User
	Store            *database.DataStore
	Limits           map[string]*Visitor
}

type User struct {
	ID                int
	Name              string
	Role              string
	IsAuth            bool
	Token             string
	NotificationCount int
}

func main() {

	handlerOpts := slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &handlerOpts))

	// Log the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		logger.Error("Failed to get current working directory:", err)
	} else {
		logger.Info("Current working directory:", cwd)
	}

	// Load the .env file
	err = godotenv.Load(".env")
	if err != nil {
		logger.Warn("Error loading .env file:", err)
	} else {
		logger.Info(".env file loaded successfully")
	}

	// Check if environment variables are set
	if os.Getenv("GOOGLE_CLIENT_ID") == "" {
		logger.Warn("GOOGLE_CLIENT_ID is not set")
	} else {
		logger.Info("GOOGLE_CLIENT_ID is set")
	}

	if os.Getenv("GITHUB_CLIENT_ID") == "" {
		logger.Warn("GITHUB_CLIENT_ID is not set")
	} else {
		logger.Info("GITHUB_CLIENT_ID is set")
	}

	// ... rest of your code ...

	err = utils.CachingTemplates()
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
		Limits:           createLimiter(),
	}

	logger.Info("Listening on https://localhost:8443...")

	certFile := "tls/cert.pem"
	keyFile := "tls/key.pem"

	serverErr := http.ListenAndServeTLS(":8443", certFile, keyFile, app.routes())
	logger.Error(serverErr.Error())
	os.Exit(1)
}
