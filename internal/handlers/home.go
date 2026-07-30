package handlers

import "net/http"

func (h *Handlers) home(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/tasks", http.StatusSeeOther)
}
