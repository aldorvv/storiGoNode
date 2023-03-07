// Data repository module
package db

import (
	"database/sql"
	"os"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var (
	singleton sync.Once
	repo      *sql.DB
)

// Create a new Database connection
func New() *sql.DB {
	uri := os.Getenv("DB_URI")
	db, err := sql.Open("mysql", uri)
	if err != nil {
		panic(err.Error())
	}
	return db
}

// Singleton for DB connection
func GetRepo() *sql.DB {
	singleton.Do(func() {
		repo = New()
	})
	return repo
}
