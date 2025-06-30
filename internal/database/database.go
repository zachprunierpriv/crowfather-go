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

func ConnectDb() (*DatabaseService, error) {
	connStr := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=disable", os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database instance %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database instance %v", err)
	}
	svc := &DatabaseService{db: db}
	if err := svc.init(); err != nil {
		return nil, err
	}
	return svc, nil
}

func (d *DatabaseService) init() error {
	createThreads := `CREATE TABLE IF NOT EXISTS threads (
        context_id TEXT PRIMARY KEY,
        thread_id TEXT NOT NULL
    );`
	if _, err := d.db.Exec(createThreads); err != nil {
		return err
	}

	createMessages := `CREATE TABLE IF NOT EXISTS messages (
        id SERIAL PRIMARY KEY,
        thread_id TEXT NOT NULL,
        role TEXT NOT NULL,
        message_id TEXT NOT NULL,
        content TEXT NOT NULL
    );`
	if _, err := d.db.Exec(createMessages); err != nil {
		return err
	}
	return nil
}

func (d *DatabaseService) SaveThread(contextID, threadID string) error {
	_, err := d.db.Exec(`INSERT INTO threads (context_id, thread_id) VALUES ($1,$2)
        ON CONFLICT (context_id) DO UPDATE SET thread_id = EXCLUDED.thread_id`, contextID, threadID)
	return err
}

func (d *DatabaseService) GetThread(contextID string) (string, error) {
	var threadID string
	err := d.db.QueryRow(`SELECT thread_id FROM threads WHERE context_id=$1`, contextID).Scan(&threadID)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("thread not found")
	}
	return threadID, err
}

func (d *DatabaseService) SaveMessage(threadID, role, messageID, content string) error {
	_, err := d.db.Exec(`INSERT INTO messages (thread_id, role, message_id, content) VALUES ($1,$2,$3,$4)`,
		threadID, role, messageID, content)
	return err
}
