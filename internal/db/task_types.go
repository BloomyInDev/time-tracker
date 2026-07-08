package db

import (
	"database/sql"

	"github.com/bloomyindev/time-tracker/internal/models"
)

func CreateTaskType(conn *sql.DB, userID int64, name string) (models.TaskType, error) {
	res, err := conn.Exec(`INSERT INTO task_types (name, user_id) VALUES (?, ?)`, name, userID)
	if err != nil {
		return models.TaskType{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.TaskType{}, err
	}
	return models.TaskType{ID: id, UserID: userID, Name: name}, nil
}

func ListTaskTypes(conn *sql.DB, userID int64) ([]models.TaskType, error) {
	rows, err := conn.Query(`SELECT id, user_id, name FROM task_types WHERE user_id = ? ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []models.TaskType
	for rows.Next() {
		var t models.TaskType
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name); err != nil {
			return nil, err
		}
		types = append(types, t)
	}
	return types, rows.Err()
}

func GetTaskType(conn *sql.DB, userID, id int64) (models.TaskType, error) {
	var t models.TaskType
	err := conn.QueryRow(`SELECT id, user_id, name FROM task_types WHERE id = ? AND user_id = ?`, id, userID).
		Scan(&t.ID, &t.UserID, &t.Name)
	return t, err
}

func UpdateTaskType(conn *sql.DB, userID, id int64, name string) error {
	_, err := conn.Exec(`UPDATE task_types SET name = ? WHERE id = ? AND user_id = ?`, name, id, userID)
	return err
}

func DeleteTaskType(conn *sql.DB, userID, id int64) error {
	_, err := conn.Exec(`DELETE FROM task_types WHERE id = ? AND user_id = ?`, id, userID)
	return err
}
