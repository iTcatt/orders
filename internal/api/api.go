package api

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func SendJSON(w http.ResponseWriter, data any, code int) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		SendInternalError(w, "internal server error")
	}
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
