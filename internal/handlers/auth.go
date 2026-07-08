package handlers

import (
	"net/http"

	"github.com/bloomyindev/time-tracker/internal/templates"
)

func Login(w http.ResponseWriter, r *http.Request) {
	templates.Login().Render(r.Context(), w)
}
