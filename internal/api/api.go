package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
)

type errorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func SendJSON(w http.ResponseWriter, data any, code int) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		slog.Error("failed to encode response", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(buf.Bytes())
}

func SendError(w http.ResponseWriter, msg string, code int) {
	SendJSON(w, errorResponse{
		Message: msg,
		Code:    code,
	}, code)
}

func SendValidationError(w http.ResponseWriter, msg string) {
	SendError(w, msg, http.StatusBadRequest)
}

func SendNotFoundError(w http.ResponseWriter, msg string) {
	SendError(w, msg, http.StatusNotFound)
}

func SendInternalError(w http.ResponseWriter, msg string) {
	SendError(w, msg, http.StatusInternalServerError)
}
