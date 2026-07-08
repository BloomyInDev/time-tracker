package db

import (
	"database/sql"

	"github.com/bloomyindev/time-tracker/internal/models"
)

func GetUser(conn *sql.DB, id int64) (models.User, error) {
	var u models.User
	err := conn.QueryRow(`SELECT id, email, password_hash FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.Email, &u.PasswordHash)
	return u, err
}

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
