package models

import "time"

type Task struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	ClientID   int64     `json:"client_id"`
	TaskTypeID int64     `json:"task_type_id"`
	PeriodID   int64     `json:"period_id"`
	Title      string    `json:"title"`
	HoursSpent float64   `json:"hours_spent"`
	Date       time.Time `json:"date"`
}
