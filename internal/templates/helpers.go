package templates

import (
	"context"
	"strconv"
	"time"

	"github.com/bloomyindev/time-tracker/internal/models"
	"github.com/invopop/ctxi18n/i18n"
)

// weekdayLabels returns the localized weekday names, index 0 = Monday ..
// 6 = Sunday, matching models.User.DailyHours ordering.
func weekdayLabels(ctx context.Context) []string {
	keys := []string{"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday"}
	labels := make([]string, len(keys))
	for i, k := range keys {
		labels[i] = i18n.T(ctx, "account."+k)
	}
	return labels
}

// monthLabel returns the localized month name followed by the year, e.g.
// "Juillet 2026". time.Month is 1-based (January = 1).
func monthLabel(ctx context.Context, m time.Month, year int) string {
	keys := []string{"january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"}
	return i18n.T(ctx, "months."+keys[int(m)-1]) + " " + strconv.Itoa(year)
}

func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}

func clientName(clients []models.Client, id int64) string {
	for _, c := range clients {
		if c.ID == id {
			return c.Name
		}
	}
	return ""
}

func taskTypeName(types []models.TaskType, id int64) string {
	for _, t := range types {
		if t.ID == id {
			return t.Name
		}
	}
	return ""
}

func periodName(periods []models.Period, id int64) string {
	for _, p := range periods {
		if p.ID == id {
			return p.Name
		}
	}
	return ""
}

func formatHours(h float64) string {
	return strconv.FormatFloat(h, 'f', 3, 64)
}

// formatDiff shows a signed hours delta, e.g. "+1.500" for overtime or
// "-2.000" for missing hours.
func formatDiff(h float64) string {
	if h > 0 {
		return "+" + formatHours(h)
	}
	return formatHours(h)
}

// diffClass colors a diff: green when overtime, red when short, muted at
// exactly on target.
func diffClass(h float64) string {
	switch {
	case h > 0:
		return "has-text-success"
	case h < 0:
		return "has-text-danger"
	default:
		return "has-text-grey"
	}
}

// printDiffClass colors a diff on the standalone print report using its own
// CSS classes (the report doesn't load Bulma).
func printDiffClass(h float64) string {
	switch {
	case h > 0:
		return "pos"
	case h < 0:
		return "neg"
	default:
		return ""
	}
}

// serializeClientTaskTypes encodes a client_id -> task_type_ids map as
// "clientID:ttID,ttID;clientID:ttID" so it can sit in a plain data
// attribute without needing JSON (and its quote-escaping headaches).
func serializeClientTaskTypes(byClient map[int64][]int64) string {
	out := ""
	for clientID, ttIDs := range byClient {
		if out != "" {
			out += ";"
		}
		out += itoa(clientID) + ":"
		for i, ttID := range ttIDs {
			if i > 0 {
				out += ","
			}
			out += itoa(ttID)
		}
	}
	return out
}
