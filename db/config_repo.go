// db/config_repo.go
package db

import "database/sql"

// ConfigRepo handles CRUD for app_config.
// NOTE: ApiToken is never written to or read from this table.
// The caller is responsible for reading/writing the token via fyne.Preferences.
type ConfigRepo struct{ db *sql.DB }

// Get retrieves the stored config. Returns nil (no error) if not yet configured.
func (r *ConfigRepo) Get() (*AppConfig, error) {
	row := r.db.QueryRow(`
		SELECT jira_domain, email, work_ref_field_id,
		       vertical_field_id, company_field_id, uat_end_field_id
		FROM app_config WHERE id = 1`)
	cfg := &AppConfig{}
	err := row.Scan(
		&cfg.JiraDomain, &cfg.Email,
		&cfg.WorkRefFieldID, &cfg.VerticalFieldID,
		&cfg.CompanyFieldID, &cfg.UatEndFieldID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// Save inserts or updates the config record (upsert on id=1).
// ApiToken in cfg is intentionally ignored — use fyne.Preferences for that.
func (r *ConfigRepo) Save(cfg *AppConfig) error {
	_, err := r.db.Exec(`
		INSERT INTO app_config
		  (id, jira_domain, email, work_ref_field_id,
		   vertical_field_id, company_field_id, uat_end_field_id, updated_at)
		VALUES (1, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(id) DO UPDATE SET
		  jira_domain        = excluded.jira_domain,
		  email              = excluded.email,
		  work_ref_field_id  = excluded.work_ref_field_id,
		  vertical_field_id  = excluded.vertical_field_id,
		  company_field_id   = excluded.company_field_id,
		  uat_end_field_id   = excluded.uat_end_field_id,
		  updated_at         = excluded.updated_at`,
		cfg.JiraDomain, cfg.Email,
		cfg.WorkRefFieldID, cfg.VerticalFieldID,
		cfg.CompanyFieldID, cfg.UatEndFieldID,
	)
	return err
}
