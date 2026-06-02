package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dataDir string) error {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	dbPath := filepath.Join(dataDir, "lite-dokploy.db")
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	DB.SetMaxOpenConns(1)

	if err := migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	log.Println("database initialized at", dbPath)
	return nil
}

func migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS applications (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			repo_url TEXT NOT NULL DEFAULT '',
			branch TEXT NOT NULL DEFAULT '',
			compose_path TEXT NOT NULL DEFAULT 'docker-compose.yml',
			compose_content TEXT NOT NULL DEFAULT '',
			domain TEXT NOT NULL DEFAULT '',
			env_vars TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'stopped',
			source TEXT NOT NULL DEFAULT 'git',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS deployments (
			id TEXT PRIMARY KEY,
			application_id TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'running',
			commit_sha TEXT NOT NULL DEFAULT '',
			commit_message TEXT NOT NULL DEFAULT '',
			log_path TEXT NOT NULL DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (application_id) REFERENCES applications(id) ON DELETE CASCADE
		)`,
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return err
		}
	}

	DB.Exec(`ALTER TABLE applications ADD COLUMN compose_content TEXT NOT NULL DEFAULT ''`)

	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
