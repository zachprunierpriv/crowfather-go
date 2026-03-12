package database

import (
	"context"
	"database/sql"
	"fmt"
)

type PgThreadRepository struct {
	db *sql.DB
}

func NewPgThreadRepository(db *sql.DB) *PgThreadRepository {
	return &PgThreadRepository{db: db}
}

func (r *PgThreadRepository) Migrate(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS thread_ids (
			context_id  TEXT PRIMARY KEY,
			thread_id   TEXT NOT NULL,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to run thread_ids migration: %w", err)
	}
	return nil
}

func (r *PgThreadRepository) GetThreadID(ctx context.Context, contextID string) (string, error) {
	var threadID string
	err := r.db.QueryRowContext(ctx,
		`SELECT thread_id FROM thread_ids WHERE context_id = $1`,
		contextID,
	).Scan(&threadID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get thread id for context %s: %w", contextID, err)
	}
	return threadID, nil
}

func (r *PgThreadRepository) SaveThreadID(ctx context.Context, contextID, threadID string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO thread_ids (context_id, thread_id)
		VALUES ($1, $2)
		ON CONFLICT (context_id) DO UPDATE
			SET thread_id  = EXCLUDED.thread_id,
			    updated_at = NOW()
	`, contextID, threadID)
	if err != nil {
		return fmt.Errorf("failed to save thread id for context %s: %w", contextID, err)
	}
	return nil
}
