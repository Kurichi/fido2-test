package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"
)

func New() (*sql.DB, error) {
	USER := os.Getenv("POSTGRES_USER")
	PASS := os.Getenv("POSTGRES_PASSWORD")
	HOST := os.Getenv("POSTGRES_HOST")
	DBNAME := os.Getenv("POSTGRES_DB")

	CONNECT := fmt.Sprintf("host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
		HOST, USER, PASS, DBNAME)

	db, err := sql.Open("postgres", CONNECT)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(10 * time.Second)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
