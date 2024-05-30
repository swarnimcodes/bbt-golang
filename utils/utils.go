package utils

import (
	"encoding/json"
	"net/http"
)

func SendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := ErrorResponse{Error: message}
	json.NewEncoder(w).Encode(response)
	return
}

func SendMessageResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := MessageResponse{Message: message}
	json.NewEncoder(w).Encode(response)
	return
}

// utils.SendJSONMessageResponse(w, string(customerSummaryJSON), http.StatusOK)

func SendJSONMessageResponse(w http.ResponseWriter, jsonData interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(jsonData)
	return
}
