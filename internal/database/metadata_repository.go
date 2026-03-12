package database

import (
	"context"
	"database/sql"
	"fmt"
)

type PgMetadataRepository struct {
	db *sql.DB
}

func NewPgMetadataRepository(db *sql.DB) *PgMetadataRepository {
	return &PgMetadataRepository{db: db}
}

func (r *PgMetadataRepository) Migrate(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS metadata (
			key        TEXT PRIMARY KEY,
			value      TEXT NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to run metadata migration: %w", err)
	}
	return nil
}

func (r *PgMetadataRepository) GetMetadata(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRowContext(ctx,
		`SELECT value FROM metadata WHERE key = $1`, key,
	).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get metadata key %s: %w", key, err)
	}
	return value, nil
}

func (r *PgMetadataRepository) SetMetadata(ctx context.Context, key, value string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO metadata (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE
			SET value      = EXCLUDED.value,
			    updated_at = NOW()
	`, key, value)
	if err != nil {
		return fmt.Errorf("failed to set metadata key %s: %w", key, err)
	}
	return nil
}
