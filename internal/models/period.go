package models

type Period struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}
