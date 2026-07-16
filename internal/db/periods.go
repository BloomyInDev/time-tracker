package db

import (
	"database/sql"

	"github.com/bloomyindev/time-tracker/internal/models"
)

func CreatePeriod(conn *sql.DB, userID int64, name string) (models.Period, error) {
	res, err := conn.Exec(`INSERT INTO periods (user_id, name, is_default) VALUES (?, ?, 0)`, userID, name)
	if err != nil {
		return models.Period{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.Period{}, err
	}
	return models.Period{ID: id, UserID: userID, Name: name}, nil
}

func ListPeriods(conn *sql.DB, userID int64) ([]models.Period, error) {
	rows, err := conn.Query(`SELECT id, user_id, name, is_default FROM periods WHERE user_id = ? ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var periods []models.Period
	for rows.Next() {
		var p models.Period
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.IsDefault); err != nil {
			return nil, err
		}
		periods = append(periods, p)
	}
	return periods, rows.Err()
}

func GetDefaultPeriod(conn *sql.DB, userID int64) (models.Period, error) {
	var p models.Period
	err := conn.QueryRow(`SELECT id, user_id, name, is_default FROM periods WHERE user_id = ? AND is_default = 1`, userID).
		Scan(&p.ID, &p.UserID, &p.Name, &p.IsDefault)
	return p, err
}

func SetDefaultPeriod(conn *sql.DB, userID, id int64) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE periods SET is_default = 0 WHERE user_id = ?`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE periods SET is_default = 1 WHERE id = ? AND user_id = ?`, id, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func GetPeriod(conn *sql.DB, userID, id int64) (models.Period, error) {
	var p models.Period
	err := conn.QueryRow(`SELECT id, user_id, name, is_default FROM periods WHERE id = ? AND user_id = ?`, id, userID).
		Scan(&p.ID, &p.UserID, &p.Name, &p.IsDefault)
	return p, err
}

func UpdatePeriod(conn *sql.DB, userID, id int64, name string) error {
	_, err := conn.Exec(`UPDATE periods SET name = ? WHERE id = ? AND user_id = ?`, name, id, userID)
	return err
}

func DeletePeriod(conn *sql.DB, userID, id int64) error {
	_, err := conn.Exec(`DELETE FROM periods WHERE id = ? AND user_id = ?`, id, userID)
	return err
}
