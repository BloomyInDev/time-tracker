package db

import (
	"database/sql"

	"github.com/bloomyindev/time-tracker/internal/models"
)

func periodIDParam(periodID int64) sql.NullInt64 {
	return sql.NullInt64{Int64: periodID, Valid: periodID != 0}
}

func scanTask(row interface{ Scan(...any) error }, t *models.Task) error {
	var periodID sql.NullInt64
	if err := row.Scan(&t.ID, &t.UserID, &t.ClientID, &t.TaskTypeID, &periodID, &t.Title, &t.HoursSpent, &t.Date); err != nil {
		return err
	}
	t.PeriodID = periodID.Int64
	return nil
}

func CreateTask(conn *sql.DB, t models.Task) (models.Task, error) {
	res, err := conn.Exec(
		`INSERT INTO tasks (user_id, client_id, task_type_id, period_id, title, hours_spent, date) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.UserID, t.ClientID, t.TaskTypeID, periodIDParam(t.PeriodID), t.Title, t.HoursSpent, t.Date,
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
		`SELECT id, user_id, client_id, task_type_id, period_id, title, hours_spent, date FROM tasks WHERE user_id = ? ORDER BY date DESC, id DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := scanTask(rows, &t); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// ListTasksByClientFiltered lists a client's tasks, optionally narrowed by
// period and/or task type. A zero id for either filter means "any".
func ListTasksByClientFiltered(conn *sql.DB, userID, clientID, periodID, taskTypeID int64) ([]models.Task, error) {
	query := `SELECT id, user_id, client_id, task_type_id, period_id, title, hours_spent, date FROM tasks
		 WHERE user_id = ? AND client_id = ?`
	args := []any{userID, clientID}
	if periodID != 0 {
		query += ` AND period_id = ?`
		args = append(args, periodID)
	}
	if taskTypeID != 0 {
		query += ` AND task_type_id = ?`
		args = append(args, taskTypeID)
	}
	query += ` ORDER BY date DESC, id DESC`

	rows, err := conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := scanTask(rows, &t); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func GetTask(conn *sql.DB, userID, id int64) (models.Task, error) {
	var t models.Task
	row := conn.QueryRow(
		`SELECT id, user_id, client_id, task_type_id, period_id, title, hours_spent, date FROM tasks WHERE id = ? AND user_id = ?`,
		id, userID,
	)
	err := scanTask(row, &t)
	return t, err
}

func UpdateTask(conn *sql.DB, t models.Task) error {
	_, err := conn.Exec(
		`UPDATE tasks SET client_id = ?, task_type_id = ?, period_id = ?, title = ?, hours_spent = ?, date = ? WHERE id = ? AND user_id = ?`,
		t.ClientID, t.TaskTypeID, periodIDParam(t.PeriodID), t.Title, t.HoursSpent, t.Date, t.ID, t.UserID,
	)
	return err
}

func DeleteTask(conn *sql.DB, userID, id int64) error {
	_, err := conn.Exec(`DELETE FROM tasks WHERE id = ? AND user_id = ?`, id, userID)
	return err
}
