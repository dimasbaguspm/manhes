package httputil

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// ErrorResponse is the standard JSON error envelope.
type ErrorResponse struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []string `json:"details"`
}

func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string, err error) {
	resp := ErrorResponse{Code: code, Message: msg}
	if err != nil {
		resp.Details = []string{err.Error()}
	}
	WriteJSON(w, status, resp)
}

// 4xx

func BadRequest(w http.ResponseWriter, msg string, err error) {
	writeError(w, http.StatusBadRequest, "BAD_REQUEST", msg, err)
}

func Unauthorized(w http.ResponseWriter, msg string, err error) {
	writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", msg, err)
}

func Forbidden(w http.ResponseWriter, msg string, err error) {
	writeError(w, http.StatusForbidden, "FORBIDDEN", msg, err)
}

func NotFound(w http.ResponseWriter, msg string, err error) {
	writeError(w, http.StatusNotFound, "NOT_FOUND", msg, err)
}

func Conflict(w http.ResponseWriter, msg string, err error) {
	writeError(w, http.StatusConflict, "CONFLICT", msg, err)
}

func UnprocessableEntity(w http.ResponseWriter, msg string, err error) {
	writeError(w, http.StatusUnprocessableEntity, "UNPROCESSABLE_ENTITY", msg, err)
}

func TooManyRequests(w http.ResponseWriter, msg string, err error) {
	writeError(w, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", msg, err)
}

// 5xx

func InternalError(w http.ResponseWriter, err error) {
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", err)
}

func ServiceUnavailable(w http.ResponseWriter, err error) {
	writeError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "service temporarily unavailable", err)
}

func IntQueryParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return defaultVal
	}
	return v
}
