package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorEnvelope struct {
	Error Error `json:"error"`
}

const (
	CodeNotFound         = "not_found"
	CodeMethodNotAllowed = "method_not_allowed"
	CodeBadRequest       = "bad_request"
	CodePayloadTooLarge  = "payload_too_large"
	CodeInternal         = "internal_error"
	CodeUnavailable      = "unavailable"
	CodeUnauthorized     = "unauthorized"
	CodeForbidden        = "forbidden"
	CodeTooManyAttempts  = "too_many_attempts"
)

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("writing response body failed", "error", err)
	}
}

func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, errorEnvelope{Error: Error{Code: code, Message: message}})
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotFound, CodeNotFound, "This endpoint does not exist.")
}

func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusMethodNotAllowed, CodeMethodNotAllowed,
		"This endpoint does not accept "+r.Method+" requests.")
}
