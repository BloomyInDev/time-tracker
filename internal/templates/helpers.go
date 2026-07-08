package templates

import "strconv"

func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}

func formatHours(h float64) string {
	return strconv.FormatFloat(h, 'f', 2, 64)
}
