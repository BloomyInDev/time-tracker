package db

import (
	"database/sql"

	"github.com/bloomyindev/time-tracker/internal/models"
)

func CreateTask(conn *sql.DB, t models.Task) (models.Task, error) {
	res, err := conn.Exec(
		`INSERT INTO tasks (user_id, client_id, task_type_id, title, hours_spent, date) VALUES (?, ?, ?, ?, ?, ?)`,
		t.UserID, t.ClientID, t.TaskTypeID, t.Title, t.HoursSpent, t.Date,
	)
	if err != nil {
		return models.Task{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return models.Task{}, err
	}
	t.ID = id
	return t, nil
}

func ListTasks(conn *sql.DB, userID int64) ([]models.Task, error) {
	rows, err := conn.Query(
		`SELECT id, user_id, client_id, task_type_id, title, hours_spent, date FROM tasks WHERE user_id = ? ORDER BY date DESC, id DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.ClientID, &t.TaskTypeID, &t.Title, &t.HoursSpent, &t.Date); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func ListTasksByClient(conn *sql.DB, userID, clientID int64) ([]models.Task, error) {
	rows, err := conn.Query(
		`SELECT id, user_id, client_id, task_type_id, title, hours_spent, date FROM tasks
		 WHERE user_id = ? AND client_id = ? ORDER BY date DESC, id DESC`,
		userID, clientID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.ClientID, &t.TaskTypeID, &t.Title, &t.HoursSpent, &t.Date); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func GetTask(conn *sql.DB, userID, id int64) (models.Task, error) {
	var t models.Task
	err := conn.QueryRow(
		`SELECT id, user_id, client_id, task_type_id, title, hours_spent, date FROM tasks WHERE id = ? AND user_id = ?`,
		id, userID,
	).Scan(&t.ID, &t.UserID, &t.ClientID, &t.TaskTypeID, &t.Title, &t.HoursSpent, &t.Date)
	return t, err
}

func UpdateTask(conn *sql.DB, t models.Task) error {
	_, err := conn.Exec(
		`UPDATE tasks SET client_id = ?, task_type_id = ?, title = ?, hours_spent = ?, date = ? WHERE id = ? AND user_id = ?`,
		t.ClientID, t.TaskTypeID, t.Title, t.HoursSpent, t.Date, t.ID, t.UserID,
	)
	return err
}

func DeleteTask(conn *sql.DB, userID, id int64) error {
	_, err := conn.Exec(`DELETE FROM tasks WHERE id = ? AND user_id = ?`, id, userID)
	return err
}
