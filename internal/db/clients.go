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

// scanClients collects the rows of a query selecting
// id, user_id, name, is_archived from clients.
func scanClients(rows *sql.Rows) ([]models.Client, error) {
	defer rows.Close()

	var clients []models.Client
	for rows.Next() {
		var c models.Client
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.IsArchived); err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}
	return clients, rows.Err()
}

func ListClients(conn *sql.DB, userID int64) ([]models.Client, error) {
	rows, err := conn.Query(`SELECT id, user_id, name, is_archived FROM clients WHERE user_id = ? ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	return scanClients(rows)
}

func ListClientsOrderedByName(conn *sql.DB, userID int64) ([]models.Client, error) {
	rows, err := conn.Query(`SELECT id, user_id, name, is_archived FROM clients WHERE user_id = ? ORDER BY name`, userID)
	if err != nil {
		return nil, err
	}
	return scanClients(rows)
}

func GetClient(conn *sql.DB, userID, id int64) (models.Client, error) {
	var c models.Client
	err := conn.QueryRow(`SELECT id, user_id, name, is_archived FROM clients WHERE id = ? AND user_id = ?`, id, userID).
		Scan(&c.ID, &c.UserID, &c.Name, &c.IsArchived)
	return c, err
}

// UpdateClient saves the client's editable fields. An archived client
// keeps its history but accepts no new tasks.
func UpdateClient(conn *sql.DB, userID, id int64, name string, archived bool) error {
	_, err := conn.Exec(`UPDATE clients SET name = ?, is_archived = ? WHERE id = ? AND user_id = ?`, name, archived, id, userID)
	return err
}

func DeleteClient(conn *sql.DB, userID, id int64) error {
	_, err := conn.Exec(`DELETE FROM clients WHERE id = ? AND user_id = ?`, id, userID)
	return err
}
