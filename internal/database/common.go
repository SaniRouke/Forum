package database

import (
	"database/sql"
	"fmt"
	"log/slog"
)

type DataStore struct {
	User UserDBInterface
	Post PostDBInterface
}

func CreateDataStore(db *sql.DB) *DataStore {
	return &DataStore{
		DataUserWorkerCreation(db),
		DataPostWorkerCreation(db),
	}
}

func InitializeDB(dataSourceName string, logger *slog.Logger) (*sql.DB, error) {
	var err error
	DB, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	} else {
		logger.Info("DB OK")
	}

	err = DB.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return DB, nil
}
