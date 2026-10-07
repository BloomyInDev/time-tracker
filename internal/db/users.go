package db

import (
	"database/sql"

	"github.com/bloomyindev/time-tracker/internal/models"
)

const userColumns = `id, email, password_hash,
	hours_mon, hours_tue, hours_wed, hours_thu, hours_fri, hours_sat, hours_sun, time_start_date, vocabulary`

func scanUser(row interface{ Scan(...any) error }, u *models.User) error {
	var start sql.NullString
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash,
		&u.DailyHours[0], &u.DailyHours[1], &u.DailyHours[2], &u.DailyHours[3],
		&u.DailyHours[4], &u.DailyHours[5], &u.DailyHours[6], &start, &u.Vocabulary)
	u.TimeStartDate = start.String
	return err
}

func GetUser(conn *sql.DB, id int64) (models.User, error) {
	var u models.User
	err := scanUser(conn.QueryRow(`SELECT `+userColumns+` FROM users WHERE id = ?`, id), &u)
	return u, err
}

func ListUsers(conn *sql.DB) ([]models.User, error) {
	rows, err := conn.Query(`SELECT ` + userColumns + ` FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := scanUser(rows, &u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// UpdateTimeSettings saves a user's per-weekday hours target (index 0 =
// Monday .. 6 = Sunday) and the default start date of the /time page. An
// empty startDate clears it.
func UpdateTimeSettings(conn *sql.DB, id int64, hours [7]float64, startDate string) error {
	var start any
	if startDate != "" {
		start = startDate
	}
	_, err := conn.Exec(
		`UPDATE users SET hours_mon = ?, hours_tue = ?, hours_wed = ?, hours_thu = ?, hours_fri = ?, hours_sat = ?, hours_sun = ?, time_start_date = ? WHERE id = ?`,
		hours[0], hours[1], hours[2], hours[3], hours[4], hours[5], hours[6], start, id,
	)
	return err
}

// UpdateVocabulary saves the wording preset a user picked. The caller
// validates the name.
func UpdateVocabulary(conn *sql.DB, id int64, name string) error {
	_, err := conn.Exec(`UPDATE users SET vocabulary = ? WHERE id = ?`, name, id)
	return err
}
