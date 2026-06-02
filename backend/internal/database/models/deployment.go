package models

import (
	"database/sql"
	"time"

	"github.com/lite-dokploy/backend/internal/database"
	"github.com/lite-dokploy/backend/internal/types"
)

func ListDeployments(appID string) ([]types.Deployment, error) {
	rows, err := database.DB.Query(`
		SELECT id, application_id, status, commit_sha, commit_message, log_path, created_at, updated_at
		FROM deployments WHERE application_id = ? ORDER BY created_at DESC
	`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deps []types.Deployment
	for rows.Next() {
		var d types.Deployment
		if err := rows.Scan(&d.ID, &d.ApplicationID, &d.Status, &d.CommitSHA, &d.CommitMessage, &d.LogPath, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		deps = append(deps, d)
	}
	return deps, nil
}

func GetDeployment(id string) (*types.Deployment, error) {
	var d types.Deployment
	err := database.DB.QueryRow(`
		SELECT id, application_id, status, commit_sha, commit_message, log_path, created_at, updated_at
		FROM deployments WHERE id = ?
	`, id).Scan(&d.ID, &d.ApplicationID, &d.Status, &d.CommitSHA, &d.CommitMessage, &d.LogPath, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func GetLatestDeployment(appID string) (*types.Deployment, error) {
	var d types.Deployment
	err := database.DB.QueryRow(`
		SELECT id, application_id, status, commit_sha, commit_message, log_path, created_at, updated_at
		FROM deployments WHERE application_id = ? ORDER BY created_at DESC LIMIT 1
	`, appID).Scan(&d.ID, &d.ApplicationID, &d.Status, &d.CommitSHA, &d.CommitMessage, &d.LogPath, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func CreateDeployment(d *types.Deployment) error {
	_, err := database.DB.Exec(`
		INSERT INTO deployments (id, application_id, status, commit_sha, commit_message, log_path, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, d.ID, d.ApplicationID, d.Status, d.CommitSHA, d.CommitMessage, d.LogPath, time.Now(), time.Now())
	return err
}

func UpdateDeploymentStatus(id string, status types.DeployStatus) error {
	_, err := database.DB.Exec(`UPDATE deployments SET status=?, updated_at=? WHERE id=?`, status, time.Now(), id)
	return err
}

func UpdateDeployment(d *types.Deployment) error {
	_, err := database.DB.Exec(`
		UPDATE deployments SET status=?, commit_sha=?, commit_message=?, log_path=?, updated_at=? WHERE id=?
	`, d.Status, d.CommitSHA, d.CommitMessage, d.LogPath, time.Now(), d.ID)
	return err
}
