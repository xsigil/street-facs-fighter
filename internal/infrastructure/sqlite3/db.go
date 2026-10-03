package sqlite3

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func NewDB(ctx context.Context, dbPath string) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	// SQLite の単一ライター特性に対応
	db.SetMaxOpenConns(1)

	schema := `
	CREATE TABLE IF NOT EXISTS action_units (
		code TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		muscle TEXT NOT NULL,
		clues TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS questions (
		id TEXT PRIMARY KEY,
		file_path TEXT NOT NULL,
		file_name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS question_targets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		question_id TEXT NOT NULL,
		side TEXT,
		code TEXT NOT NULL,
		intensity TEXT,
		FOREIGN KEY(question_id) REFERENCES questions(id) ON DELETE CASCADE
	);
	`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return nil, fmt.Errorf("failed to apply schema: %w", err)
	}

	return db, nil
}
