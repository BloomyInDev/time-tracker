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

// buildTimeView folds a user's tasks into per-day summaries, groups them by
// month and sums per-month and per-range totals. The optional from/to bounds
// (inclusive, "2006-01-02") narrow the range; an empty bound is unbounded.
func buildTimeView(conn *sql.DB, userID int64, from, to string) (templates.TimeView, error) {
	user, err := db.GetUser(conn, userID)
	if err != nil {
		return templates.TimeView{}, err
	}
	tasks, err := db.ListTasks(conn, userID)
	if err != nil {
		return templates.TimeView{}, err
	}

	// Tasks come back ordered by date desc, so we can fold consecutive
	// same-day tasks into the current day without a map. workedPositive
	// tracks, per day, whether any actual work (positive hours) was
	// logged that day, as opposed to a day made up solely of adjustment
	// entries (e.g. a recovery day off, or overtime paid out).
	var days []templates.DaySummary
	var workedPositive []bool
	for _, t := range tasks {
		key := t.Date.Format("2006-01-02")
		if n := len(days); n > 0 && days[n-1].Date.Format("2006-01-02") == key {
			days[n-1].Hours += t.HoursSpent
			if t.HoursSpent > 0 {
				workedPositive[n-1] = true
			}
			continue
		}
		days = append(days, templates.DaySummary{
			Date:   t.Date,
			Hours:  t.HoursSpent,
			Target: user.DailyHours[weekdayIndex(t.Date)],
		})
		workedPositive = append(workedPositive, t.HoursSpent > 0)
	}
	for i := range days {
		days[i].Diff = days[i].Hours - days[i].Target
		// Floor the shortfall at -Target so a day with no real work
		// (an untouched day, or a pure recovery-day-off entry) reads
		// as "-Target" rather than an unbounded negative. Days with
		// actual work logged alongside a negative adjustment (e.g.
		// overtime paid out) are left unclamped so that adjustment's
		// full value is reflected.
		if !workedPositive[i] && days[i].Diff < -(days[i].Target) {
			days[i].Diff = -(days[i].Target)
		}
	}

	filtered := from != "" || to != ""

	// An open-ended range (start given, no end) runs up to today.
	if from != "" && to == "" {
		to = time.Now().Format("2006-01-02")
	}

	// No range given at all: default to Jan 1 of this year through today.
	if from == "" && to == "" {
		now := time.Now()
		from = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		to = now.Format("2006-01-02")
	}

	view := templates.TimeView{From: from, To: to, Filtered: filtered}
	for _, d := range days {
		key := d.Date.Format("2006-01-02")
		if from != "" && key < from {
			continue
		}
		if to != "" && key > to {
			continue
		}
		n := len(view.Months)
		if n == 0 || view.Months[n-1].Year != d.Date.Year() || view.Months[n-1].Month != d.Date.Month() {
			view.Months = append(view.Months, templates.MonthSummary{Year: d.Date.Year(), Month: d.Date.Month()})
			n++
		}
		m := &view.Months[n-1]
		m.Days = append(m.Days, d)
		m.Hours += d.Hours
		m.Target += d.Target
		m.Diff += d.Diff
		view.Hours += d.Hours
		view.Target += d.Target
		view.Diff += d.Diff
	}
	return view, nil
}

// ListTimeEntries shows one row per day with the total hours logged that
// day, grouped by month, plus an optional date-to-date range filter.
func ListTimeEntries(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		view, err := buildTimeView(conn, userID, r.URL.Query().Get("from"), r.URL.Query().Get("to"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		templates.TimeEntries(view).Render(r.Context(), w)
	}
}

// TimeReport renders a print-friendly, standalone page of the same breakdown
// (no navbar) so the browser's print dialog can save it as a PDF.
func TimeReport(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.UserIDFromContext(r.Context())
		view, err := buildTimeView(conn, userID, r.URL.Query().Get("from"), r.URL.Query().Get("to"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// By default the report lists only days off target; all=1 keeps
		// the days that hit their target exactly.
		includeOnTarget := r.URL.Query().Get("all") == "1"
		templates.TimeReport(view, includeOnTarget).Render(r.Context(), w)
	}
}
