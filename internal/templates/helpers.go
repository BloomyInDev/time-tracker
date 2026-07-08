package templates

import "strconv"

func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}

func formatHours(h float64) string {
	return strconv.FormatFloat(h, 'f', 2, 64)
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
