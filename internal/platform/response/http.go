package response

import (
	"encoding/json"
	"net/http"
)

type Code string

type JSONResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type ErrorBody struct {
	Code    Code   `json:"code"`
	Message string `json:"message,omitempty"`
	Details any    `json:"details,omitempty"`
}

type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorBody `json:"error"`
}

// WriteJSON writes an JSON response
func WriteJSON(w http.ResponseWriter, status int, message string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(JSONResponse{
		Success: true,
		Message: message,
		Data:    nil,
	})
}

// WriteJSONWithData implements the WriteJSON function with add-on data field
func WriteJSONWithData(w http.ResponseWriter, status int, message string, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(JSONResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// WriteError writes an error response with an explicit HTTP status and code.
func WriteError(w http.ResponseWriter, status int, code Code, message string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: nil,
		},
	})
}

// WriteErrorWithDetails implements the WriteError function with add-on Details field
func WriteErrorWithDetails(w http.ResponseWriter, status int, code Code, message string, details any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// WriteErrorFromCode writes an error response, deriving the HTTP status from code's default mapping.
func WriteErrorFromCode(w http.ResponseWriter, code Code, message string) error {
	status, ok := StatusByCode[code]
	if !ok {
		status = http.StatusInternalServerError
	}

	return WriteError(w, status, code, message)
}

// WriteErrorFromCodeWithDetails implements WriteErrorFromCode function with add-on Details field
func WriteErrorFromCodeWithDetails(w http.ResponseWriter, code Code, message string, details any) error {
	status, ok := StatusByCode[code]
	if !ok {
		status = http.StatusInternalServerError
	}

	return WriteErrorWithDetails(w, status, code, message, details)
}
