package handlers

import (
	"net/http/httptest"
	"testing"
)

func TestCSVText(t *testing.T) {
	for in, want := range map[string]string{
		"plain":    "plain",
		"":         "",
		"=SUM(A1)": "'=SUM(A1)",
		"+1":       "'+1",
		"-1":       "'-1",
		"@cmd":     "'@cmd",
		"a=b":      "a=b",
		"\tindent": "'\tindent",
	} {
		if got := csvText(in); got != want {
			t.Errorf("csvText(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWriteCSV(t *testing.T) {
	w := httptest.NewRecorder()
	writeCSV(w, "x.csv", [][]string{{"a", "b;c"}, {csvHours(1.5), csvHours(-2)}})

	if got, want := w.Body.String(), "a;\"b;c\"\n1,500;-2,000\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
	if got := w.Header().Get("Content-Disposition"); got != `attachment; filename="x.csv"` {
		t.Errorf("Content-Disposition = %q", got)
	}
}
