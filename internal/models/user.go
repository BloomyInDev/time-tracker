package models

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	// DailyHours is the expected hours target per weekday, index 0 =
	// Monday .. 6 = Sunday. Defaults to 0 for every day.
	DailyHours [7]float64
	// TimeStartDate ("2006-01-02") is where the /time page starts when no
	// range is given. Empty means January 1 of the current year.
	TimeStartDate string
	// Vocabulary names the wording preset of the UI, one of i18n.Presets.
	Vocabulary string
}
