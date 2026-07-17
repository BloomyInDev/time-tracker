package db

import (
	"database/sql"

	"github.com/bloomyindev/time-tracker/internal/models"
)

func CreateClient(conn *sql.DB, userID int64, name string) (models.Client, error) {
	res, err := conn.Exec(`INSERT INTO clients (name, user_id) VALUES (?, ?)`, name, userID)
	if err != nil {
		return models.Client{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.Client{}, err
	}
	return models.Client{ID: id, UserID: userID, Name: name}, nil
}

func ListClients(conn *sql.DB, userID int64) ([]models.Client, error) {
	rows, err := conn.Query(`SELECT id, user_id, name FROM clients WHERE user_id = ? ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []models.Client
	for rows.Next() {
		var c models.Client
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, rows.Err()
}

func ListClientsOrderedByName(conn *sql.DB, userID int64) ([]models.Client, error) {
	rows, err := conn.Query(`SELECT id, user_id, name FROM clients WHERE user_id = ? ORDER BY name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []models.Client
	for rows.Next() {
		var c models.Client
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, rows.Err()
}

func GetClient(conn *sql.DB, userID, id int64) (models.Client, error) {
	var c models.Client
	err := conn.QueryRow(`SELECT id, user_id, name FROM clients WHERE id = ? AND user_id = ?`, id, userID).
		Scan(&c.ID, &c.UserID, &c.Name)
	return c, err
}

func UpdateClient(conn *sql.DB, userID, id int64, name string) error {
	_, err := conn.Exec(`UPDATE clients SET name = ? WHERE id = ? AND user_id = ?`, name, id, userID)
	return err
}

func DeleteClient(conn *sql.DB, userID, id int64) error {
	_, err := conn.Exec(`DELETE FROM clients WHERE id = ? AND user_id = ?`, id, userID)
	return err
}
