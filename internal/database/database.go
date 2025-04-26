package database

import (
	"database/sql"
	"fmt"
	"os"
)

type DatabaseService struct {
	db *sql.DB
}

func ConnectDb() *DatabaseService {
	connStr := fmt.Sprintf("user=%s password=%s host=%s dbname=%s", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("HOST"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		fmt.Println("Failed to connect to database instance %v", err)
	}

	if err := db.Ping(); err != nil {
		fmt.Println("Failed to connect to database instance %v", err)
	}

	return &DatabaseService{db: db}
}