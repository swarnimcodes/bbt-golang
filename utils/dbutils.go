package utils

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var (
	db   *sql.DB
	once sync.Once
	err  error
)

func initDatabase() {
	// get database path from environment variables
	databaseLocation := os.Getenv("DATABASE_LOCATION")
	if databaseLocation == "" {
		errorMessage := "no key called `DATABASE_LOCATION` found in the environment variables"
		err = errors.New(errorMessage)
		return
	}

	// check if location exists
	err = os.MkdirAll(filepath.Dir(databaseLocation), 0755)
	if err != nil {
		return
	}

	// open the database
	db, err = sql.Open("sqlite3", databaseLocation) // connection assignied to the shared db variable
	if err != nil {
		return
	}

	// ping the database to verify the connection
	err = db.Ping()
	if err != nil {
		return
	}
}

func GetDatabaseConnection() (*sql.DB, error) {
	once.Do(initDatabase)
	if err != nil {
		return nil, err
	} else {
		return db, nil
	}
}
