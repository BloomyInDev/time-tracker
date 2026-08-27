package db

import (
	"database/sql"

	"github.com/bloomyindev/time-tracker/internal/models"
)

func AssignTaskTypeToClient(conn *sql.DB, clientID, taskTypeID int64) error {
	_, err := conn.Exec(
		`INSERT OR IGNORE INTO task_types_for_client (client_id, task_type_id) VALUES (?, ?)`,
		clientID, taskTypeID,
	)
	return err
}

func UnassignTaskTypeFromClient(conn *sql.DB, clientID, taskTypeID int64) error {
	_, err := conn.Exec(
		`DELETE FROM task_types_for_client WHERE client_id = ? AND task_type_id = ?`,
		clientID, taskTypeID,
	)
	return err
}

// ListTaskTypesForClient returns the task types assigned to one of the
// user's clients. It is scoped by user on both sides of the assignment, so
// it stays empty for a client the user doesn't own.
func ListTaskTypesForClient(conn *sql.DB, userID, clientID int64) ([]models.TaskType, error) {
	rows, err := conn.Query(`
		SELECT tt.id, tt.user_id, tt.name
		FROM task_types tt
		JOIN task_types_for_client ttc ON ttc.task_type_id = tt.id
		JOIN clients c ON c.id = ttc.client_id
		WHERE ttc.client_id = ? AND c.user_id = ? AND tt.user_id = ?
		ORDER BY tt.name`, clientID, userID, userID)
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

// ListTaskTypesByClient returns, for a user's clients, the set of task type
// IDs assigned to each client. Used to drive client-side filtering of the
// task type dropdown on the task creation form.
func ListTaskTypesByClient(conn *sql.DB, userID int64) (map[int64][]int64, error) {
	rows, err := conn.Query(`
		SELECT ttc.client_id, ttc.task_type_id
		FROM task_types_for_client ttc
		JOIN clients c ON c.id = ttc.client_id
		WHERE c.user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]int64)
	for rows.Next() {
		var clientID, taskTypeID int64
		if err := rows.Scan(&clientID, &taskTypeID); err != nil {
			return nil, err
		}
		result[clientID] = append(result[clientID], taskTypeID)
	}
	return result, rows.Err()
}

// ListTaskTypeIDsUsedByClient returns the task type IDs a client's existing
// tasks carry. Assignments can change after the fact, so a filter over a
// client's tasks has to account for types that are no longer assigned.
func ListTaskTypeIDsUsedByClient(conn *sql.DB, userID, clientID int64) ([]int64, error) {
	rows, err := conn.Query(
		`SELECT DISTINCT task_type_id FROM tasks WHERE user_id = ? AND client_id = ?`,
		userID, clientID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
