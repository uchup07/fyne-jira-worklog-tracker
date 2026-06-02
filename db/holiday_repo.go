// db/holiday_repo.go
package db

import "database/sql"

// HolidayRepo handles CRUD for public_holidays.
type HolidayRepo struct{ db *sql.DB }

func (r *HolidayRepo) List() ([]PublicHoliday, error) {
	rows, err := r.db.Query(`SELECT id, date, name FROM public_holidays ORDER BY date`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var holidays []PublicHoliday
	for rows.Next() {
		var h PublicHoliday
		if err := rows.Scan(&h.ID, &h.Date, &h.Name); err != nil {
			return nil, err
		}
		holidays = append(holidays, h)
	}
	return holidays, rows.Err()
}

func (r *HolidayRepo) Add(h PublicHoliday) error {
	_, err := r.db.Exec(`
		INSERT INTO public_holidays (date, name) VALUES (?, ?)
		ON CONFLICT(date) DO UPDATE SET name = excluded.name`,
		h.Date, h.Name)
	return err
}

func (r *HolidayRepo) Delete(id int) error {
	_, err := r.db.Exec(`DELETE FROM public_holidays WHERE id = ?`, id)
	return err
}
