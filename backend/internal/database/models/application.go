package models

import (
	"database/sql"
	"time"

	"github.com/lite-dokploy/backend/internal/database"
	"github.com/lite-dokploy/backend/internal/types"
)

func ListApplications() ([]types.Application, error) {
	rows, err := database.DB.Query(`
		SELECT id, name, repo_url, branch, compose_path, compose_content, env_vars, status, source, created_at, updated_at
		FROM applications ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []types.Application
	for rows.Next() {
		var a types.Application
		if err := rows.Scan(&a.ID, &a.Name, &a.RepoURL, &a.Branch, &a.ComposePath, &a.ComposeContent, &a.EnvVars, &a.Status, &a.Source, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		apps = append(apps, a)
	}
	return apps, nil
}

func GetApplication(id string) (*types.Application, error) {
	var a types.Application
	err := database.DB.QueryRow(`
		SELECT id, name, repo_url, branch, compose_path, compose_content, env_vars, status, source, created_at, updated_at
		FROM applications WHERE id = ?
	`, id).Scan(&a.ID, &a.Name, &a.RepoURL, &a.Branch, &a.ComposePath, &a.ComposeContent, &a.EnvVars, &a.Status, &a.Source, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &a, nil
}

func CreateApplication(a *types.Application) error {
	_, err := database.DB.Exec(`
		INSERT INTO applications (id, name, repo_url, branch, compose_path, compose_content, env_vars, status, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, a.ID, a.Name, a.RepoURL, a.Branch, a.ComposePath, a.ComposeContent, a.EnvVars, a.Status, a.Source, time.Now(), time.Now())
	return err
}

func UpdateApplication(a *types.Application) error {
	_, err := database.DB.Exec(`
		UPDATE applications SET name=?, repo_url=?, branch=?, compose_path=?, compose_content=?, env_vars=?, status=?, updated_at=? WHERE id=?
	`, a.Name, a.RepoURL, a.Branch, a.ComposePath, a.ComposeContent, a.EnvVars, a.Status, time.Now(), a.ID)
	return err
}

func UpdateAppStatus(id string, status string) error {
	_, err := database.DB.Exec(`UPDATE applications SET status=?, updated_at=? WHERE id=?`, status, time.Now(), id)
	return err
}

func DeleteApplication(id string) error {
	_, err := database.DB.Exec(`DELETE FROM applications WHERE id=?`, id)
	return err
}
