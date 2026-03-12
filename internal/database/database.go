package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type DatabaseService struct {
	db *sql.DB
}

func (s *DatabaseService) DB() *sql.DB {
	return s.db
}

func ConnectDb() (*DatabaseService, error) {
	connStr := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("HOST"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to reach database: %w", err)
	}

	return &DatabaseService{db: db}, nil
}
