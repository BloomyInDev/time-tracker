package db

import (
	"database/sql"

	"github.com/bloomyindev/time-tracker/internal/models"
)

const userColumns = `id, email, password_hash,
	hours_mon, hours_tue, hours_wed, hours_thu, hours_fri, hours_sat, hours_sun`

func scanUser(row interface{ Scan(...any) error }, u *models.User) error {
	return row.Scan(&u.ID, &u.Email, &u.PasswordHash,
		&u.DailyHours[0], &u.DailyHours[1], &u.DailyHours[2], &u.DailyHours[3],
		&u.DailyHours[4], &u.DailyHours[5], &u.DailyHours[6])
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

// UpdateDailyHours saves a user's per-weekday hours target (index 0 =
// Monday .. 6 = Sunday).
func UpdateDailyHours(conn *sql.DB, id int64, hours [7]float64) error {
	_, err := conn.Exec(
		`UPDATE users SET hours_mon = ?, hours_tue = ?, hours_wed = ?, hours_thu = ?, hours_fri = ?, hours_sat = ?, hours_sun = ? WHERE id = ?`,
		hours[0], hours[1], hours[2], hours[3], hours[4], hours[5], hours[6], id,
	)
	return err
}
