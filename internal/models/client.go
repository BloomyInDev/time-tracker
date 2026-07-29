package models

type Client struct {
	ID         int64  `json:"id"`
	UserID     int64  `json:"user_id"`
	Name       string `json:"name"`
	IsArchived bool   `json:"is_archived"`
}
