package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/http/middleware"
)

type errorResponse struct {
	Error string `json:"error"`
}

type userProfile = middleware.UserProfile

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
