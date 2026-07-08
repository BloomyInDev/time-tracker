package templates

import (
	"strconv"

	"github.com/bloomyindev/time-tracker/internal/models"
)

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
