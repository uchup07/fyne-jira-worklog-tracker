// db/db.go
package db

import (
	"database/sql"

	_ "modernc.org/sqlite" // pure-Go SQLite driver — no CGO required
)

// Repository is the root of all DB access. Pass it by pointer throughout the app.
type Repository struct {
	Config   *ConfigRepo
	Teams    *TeamRepo
	Holidays *HolidayRepo
	Users    *UserRepo
	db       *sql.DB
}

// Open opens (or creates) the SQLite database at path, runs the schema
// migration, and returns a fully initialised Repository.
// Use ":memory:" for tests.
func Open(path string) *Repository {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		panic("db.Open: " + err.Error())
	}
	conn.SetMaxOpenConns(1) // SQLite supports only one writer at a time
	migrate(conn)
	return &Repository{
		Config:   &ConfigRepo{conn},
		Teams:    &TeamRepo{conn},
		Holidays: &HolidayRepo{conn},
		Users:    &UserRepo{conn},
		db:       conn,
	}
}

// Close closes the underlying database connection.
func (r *Repository) Close() error {
	return r.db.Close()
}

// migrate creates all tables if they don't already exist.
// Append-only — safe to run on every startup.
func migrate(db *sql.DB) {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS app_config (
			id               INTEGER PRIMARY KEY DEFAULT 1,
			jira_domain      TEXT NOT NULL DEFAULT '',
			email            TEXT NOT NULL DEFAULT '',
			work_ref_field_id  TEXT NOT NULL DEFAULT '',
			vertical_field_id  TEXT NOT NULL DEFAULT '',
			company_field_id   TEXT NOT NULL DEFAULT '',
			uat_end_field_id   TEXT NOT NULL DEFAULT '',
			updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// api_token intentionally absent — stored in OS keychain via fyne.Preferences

		`CREATE TABLE IF NOT EXISTS teams (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE TABLE IF NOT EXISTS team_members (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			team_id    INTEGER NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
			user_id    TEXT NOT NULL,
			join_date  TEXT NOT NULL DEFAULT '',
			leave_date TEXT NOT NULL DEFAULT '',
			UNIQUE(team_id, user_id)
		)`,

		`CREATE TABLE IF NOT EXISTS public_holidays (
			id   INTEGER PRIMARY KEY AUTOINCREMENT,
			date TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL
		)`,

		`CREATE TABLE IF NOT EXISTS jira_users (
			id            TEXT PRIMARY KEY,
			display_name  TEXT NOT NULL,
			email_address TEXT NOT NULL DEFAULT '',
			avatar_url    TEXT NOT NULL DEFAULT '',
			active        INTEGER NOT NULL DEFAULT 1,
			synced_at     DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			panic("db.migrate: " + err.Error())
		}
	}
}
