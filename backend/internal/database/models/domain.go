package models

import (
	"database/sql"
	"time"

	"github.com/lite-dokploy/backend/internal/database"
	"github.com/lite-dokploy/backend/internal/types"
)

func ListAppDomains(applicationID string) ([]types.AppDomain, error) {
	rows, err := database.DB.Query(`
		SELECT id, application_id, host, port, path, internal_path, strip_path, https, service_name, created_at, updated_at
		FROM app_domains WHERE application_id = ? ORDER BY created_at ASC
	`, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []types.AppDomain
	for rows.Next() {
		var d types.AppDomain
		if err := rows.Scan(&d.ID, &d.ApplicationID, &d.Host, &d.Port, &d.Path, &d.InternalPath, &d.StripPath, &d.HTTPS, &d.ServiceName, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, nil
}

func GetAppDomain(id string) (*types.AppDomain, error) {
	var d types.AppDomain
	err := database.DB.QueryRow(`
		SELECT id, application_id, host, port, path, internal_path, strip_path, https, service_name, created_at, updated_at
		FROM app_domains WHERE id = ?
	`, id).Scan(&d.ID, &d.ApplicationID, &d.Host, &d.Port, &d.Path, &d.InternalPath, &d.StripPath, &d.HTTPS, &d.ServiceName, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func CreateAppDomain(d *types.AppDomain) error {
	_, err := database.DB.Exec(`
		INSERT INTO app_domains (id, application_id, host, port, path, internal_path, strip_path, https, service_name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, d.ID, d.ApplicationID, d.Host, d.Port, d.Path, d.InternalPath, boolToInt(d.StripPath), boolToInt(d.HTTPS), d.ServiceName, time.Now(), time.Now())
	return err
}

func UpdateAppDomain(d *types.AppDomain) error {
	_, err := database.DB.Exec(`
		UPDATE app_domains SET host=?, port=?, path=?, internal_path=?, strip_path=?, https=?, service_name=?, updated_at=? WHERE id=?
	`, d.Host, d.Port, d.Path, d.InternalPath, boolToInt(d.StripPath), boolToInt(d.HTTPS), d.ServiceName, time.Now(), d.ID)
	return err
}

func DeleteAppDomain(id string) error {
	_, err := database.DB.Exec(`DELETE FROM app_domains WHERE id=?`, id)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
