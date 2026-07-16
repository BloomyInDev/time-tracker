package models

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	// DailyHours is the expected hours target per weekday, index 0 =
	// Monday .. 6 = Sunday. Defaults to 0 for every day.
	DailyHours [7]float64
}
