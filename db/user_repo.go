// db/user_repo.go
package db

import (
	"database/sql"
	"time"
)

// UserRepo handles caching of Jira user records.
type UserRepo struct{ db *sql.DB }

// Upsert bulk-inserts or updates Jira user records.
func (r *UserRepo) Upsert(users []JiraUser) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO jira_users (id, display_name, email_address, avatar_url, active, synced_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
		  display_name  = excluded.display_name,
		  email_address = excluded.email_address,
		  avatar_url    = excluded.avatar_url,
		  active        = excluded.active,
		  synced_at     = excluded.synced_at`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	for _, u := range users {
		active := 0
		if u.Active {
			active = 1
		}
		if _, err := stmt.Exec(u.ID, u.DisplayName, u.EmailAddress, u.AvatarURL, active, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ListActive returns all users marked active, ordered by display name.
func (r *UserRepo) ListActive() ([]JiraUser, error) {
	rows, err := r.db.Query(`
		SELECT id, display_name, email_address, avatar_url, active, synced_at
		FROM jira_users WHERE active = 1 ORDER BY display_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []JiraUser
	for rows.Next() {
		var u JiraUser
		var active int
		var syncedAt string
		if err := rows.Scan(&u.ID, &u.DisplayName, &u.EmailAddress, &u.AvatarURL, &active, &syncedAt); err != nil {
			return nil, err
		}
		u.Active = active == 1
		u.SyncedAt, _ = time.Parse("2006-01-02T15:04:05Z", syncedAt)
		users = append(users, u)
	}
	return users, rows.Err()
}
