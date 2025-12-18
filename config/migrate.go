package config

import (
	"database/sql"
	"log"
	"path/filepath"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func Migrate(dsn string) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	defer db.Close()

	dir, err := filepath.Abs("./migration")
	if err != nil {
		log.Fatalf("failed to get migration dir: %v", err)
	}

	if err := goose.Up(db, dir); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
}
