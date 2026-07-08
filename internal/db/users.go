package db

import (
	"database/sql"

	"github.com/bloomyindev/time-tracker/internal/models"
)

func ListUsers(conn *sql.DB) ([]models.User, error) {
	rows, err := conn.Query(`SELECT id, email, password_hash FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
