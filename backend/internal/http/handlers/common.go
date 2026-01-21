package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func parseIDParam(r *http.Request, key string) (int64, bool) {
	val := chi.URLParam(r, key)
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func parsePagination(r *http.Request) (int, int, error) {
	limit := 20
	offset := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			return 0, 0, errInvalid("limit")
		}
		limit = value
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return 0, 0, errInvalid("offset")
		}
		offset = value
	}
	if limit > 50 {
		limit = 50
	}
	return limit, offset, nil
}

type invalidParamError struct{ field string }

func (e invalidParamError) Error() string { return "invalid " + e.field }

func errInvalid(field string) error { return invalidParamError{field: field} }
