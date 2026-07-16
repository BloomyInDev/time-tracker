package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/bloomyindev/time-tracker/internal/db"
	"github.com/bloomyindev/time-tracker/internal/service/auth"
	"github.com/bloomyindev/time-tracker/internal/templates"
)

// weekdayIndex maps a date to models.User.DailyHours order (0 = Monday
// .. 6 = Sunday). Go's time.Weekday has Sunday = 0, so shift by 6.
func weekdayIndex(t time.Time) int {
	return (int(t.Weekday()) + 6) % 7
}

// ListTimeEntries shows one row per day with the total hours logged that
// day. Tasks come back ordered by date desc, so we can fold consecutive
// same-day tasks into the current day without a map.
func ListTimeEntries(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		user, err := db.GetUser(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tasks, err := db.ListTasks(conn, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var days []templates.DaySummary
		for _, t := range tasks {
			key := t.Date.Format("2006-01-02")
			if n := len(days); n > 0 && days[n-1].Date.Format("2006-01-02") == key {
				days[n-1].Hours += t.HoursSpent
				continue
			}
			days = append(days, templates.DaySummary{
				Date:   t.Date,
				Hours:  t.HoursSpent,
				Target: user.DailyHours[weekdayIndex(t.Date)],
			})
		}
		for i := range days {
			days[i].Diff = days[i].Hours - days[i].Target
			if days[i].Diff < -(days[i].Target) {
				days[i].Diff = -(days[i].Target)
			}
		}

		templates.TimeEntries(days).Render(r.Context(), w)
	}
}
