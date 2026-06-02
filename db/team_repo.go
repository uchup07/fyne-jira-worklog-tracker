// db/team_repo.go
package db

import "database/sql"

// TeamRepo handles CRUD for teams and team_members.
type TeamRepo struct{ db *sql.DB }

func (r *TeamRepo) ListTeams() ([]Team, error) {
	rows, err := r.db.Query(`SELECT id, name FROM teams ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var teams []Team
	for rows.Next() {
		var t Team
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

func (r *TeamRepo) CreateTeam(name string) (int, error) {
	res, err := r.db.Exec(`INSERT INTO teams (name) VALUES (?)`, name)
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return int(id), nil
}

func (r *TeamRepo) DeleteTeam(id int) error {
	_, err := r.db.Exec(`DELETE FROM teams WHERE id = ?`, id)
	return err
}

func (r *TeamRepo) ListMembers(teamID int) ([]TeamMember, error) {
	rows, err := r.db.Query(`
		SELECT id, team_id, user_id, join_date, leave_date
		FROM team_members WHERE team_id = ? ORDER BY user_id`, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []TeamMember
	for rows.Next() {
		var m TeamMember
		if err := rows.Scan(&m.ID, &m.TeamID, &m.UserID, &m.JoinDate, &m.LeaveDate); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *TeamRepo) AddMember(m TeamMember) error {
	_, err := r.db.Exec(`
		INSERT INTO team_members (team_id, user_id, join_date, leave_date)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(team_id, user_id) DO UPDATE SET
		  join_date  = excluded.join_date,
		  leave_date = excluded.leave_date`,
		m.TeamID, m.UserID, m.JoinDate, m.LeaveDate)
	return err
}

func (r *TeamRepo) RemoveMember(id int) error {
	_, err := r.db.Exec(`DELETE FROM team_members WHERE id = ?`, id)
	return err
}
