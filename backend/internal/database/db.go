package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbPath string) error {
	if err := os.MkdirAll("data", 0755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	DB.SetMaxOpenConns(1)

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	if err := migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

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
		`CREATE TABLE IF NOT EXISTS app_domains (
			id TEXT PRIMARY KEY,
			application_id TEXT NOT NULL,
			host TEXT NOT NULL,
			port INTEGER NOT NULL DEFAULT 80,
			path TEXT NOT NULL DEFAULT '/',
			internal_path TEXT NOT NULL DEFAULT '/',
			strip_path INTEGER NOT NULL DEFAULT 0,
			https INTEGER NOT NULL DEFAULT 0,
			service_name TEXT NOT NULL DEFAULT '',
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

	migrations := []string{
		`ALTER TABLE applications ADD COLUMN compose_content TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE applications ADD COLUMN env_vars TEXT NOT NULL DEFAULT ''`,
	}
	for _, q := range migrations {
		DB.Exec(q)
	}

	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}
