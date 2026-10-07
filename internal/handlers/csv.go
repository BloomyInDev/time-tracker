package handlers

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"
)

// writeCSV sends rows as a downloadable CSV. The delimiter is ";" and
// numbers use a decimal comma (see csvHours), so a double-click opens it
// correctly in Excel or LibreOffice with a French locale.
func writeCSV(w http.ResponseWriter, filename string, rows [][]string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	cw := csv.NewWriter(w)
	cw.Comma = ';'
	// A write error here means the client went away: nothing to recover.
	_ = cw.WriteAll(rows)
}

// csvText neutralizes free text (task titles, names) that a spreadsheet
// would otherwise run as a formula: a leading = + - @ (or tab/CR) gets a
// quote in front, which Excel and LibreOffice hide.
func csvText(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}

// csvHours formats an hours value with three decimals and a decimal comma.
func csvHours(h float64) string {
	return strings.Replace(strconv.FormatFloat(h, 'f', 3, 64), ".", ",", 1)
}
